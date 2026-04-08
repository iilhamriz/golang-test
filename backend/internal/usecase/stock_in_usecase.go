package usecase

import (
	"context"
	"fmt"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/calyx/smart-inventory/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockInUseCase struct {
	pool     *pgxpool.Pool
	siRepo   repository.StockInRepository
	itemRepo repository.ItemRepository
	logRepo  repository.TransactionLogRepository
}

func NewStockInUseCase(pool *pgxpool.Pool, siRepo repository.StockInRepository, itemRepo repository.ItemRepository, logRepo repository.TransactionLogRepository) *StockInUseCase {
	return &StockInUseCase{pool: pool, siRepo: siRepo, itemRepo: itemRepo, logRepo: logRepo}
}

type CreateStockInInput struct {
	ReferenceNo string                   `json:"reference_no"`
	Notes       string                   `json:"notes"`
	CreatedBy   string                   `json:"created_by"`
	Items       []CreateStockInItemInput `json:"items"`
}

type CreateStockInItemInput struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity int       `json:"quantity"`
}

func (u *StockInUseCase) Create(ctx context.Context, input CreateStockInInput) (*entity.StockInTransaction, error) {
	if input.ReferenceNo == "" || len(input.Items) == 0 {
		return nil, ErrInvalidInput
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	txn := &entity.StockInTransaction{
		ReferenceNo: input.ReferenceNo,
		Status:      entity.StockInStatusCreated,
		Notes:       input.Notes,
		CreatedBy:   input.CreatedBy,
	}

	if err := u.siRepo.Create(ctx, tx, txn); err != nil {
		return nil, fmt.Errorf("create stock in: %w", err)
	}

	for _, it := range input.Items {
		if it.Quantity <= 0 {
			return nil, ErrInvalidInput
		}
		item := &entity.StockInItem{
			StockInTransactionID: txn.ID,
			ItemID:               it.ItemID,
			Quantity:             it.Quantity,
		}
		if err := u.siRepo.CreateItem(ctx, tx, item); err != nil {
			return nil, fmt.Errorf("create stock in item: %w", err)
		}
		txn.Items = append(txn.Items, *item)
	}

	log := &entity.TransactionLog{
		TransactionType: "STOCK_IN",
		TransactionID:   txn.ID,
		FromStatus:      "",
		ToStatus:        entity.StockInStatusCreated,
		Notes:           "Transaction created",
		CreatedBy:       input.CreatedBy,
	}
	if err := u.logRepo.Create(ctx, tx, log); err != nil {
		return nil, fmt.Errorf("create log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return txn, nil
}

func (u *StockInUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus string) (*entity.StockInTransaction, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	txn, err := u.siRepo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	allowed, ok := entity.StockInAllowedTransitions[txn.Status]
	if !ok {
		return nil, ErrInvalidTransition
	}
	valid := false
	for _, s := range allowed {
		if s == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, txn.Status, newStatus)
	}

	// If transitioning to DONE: increase physical stock
	if newStatus == entity.StockInStatusDone {
		items, err := u.siRepo.GetItems(ctx, id)
		if err != nil {
			return nil, err
		}
		for _, it := range items {
			item, err := u.itemRepo.GetByIDForUpdate(ctx, tx, it.ItemID)
			if err != nil {
				return nil, err
			}
			if err := u.itemRepo.UpdatePhysicalStock(ctx, tx, it.ItemID, item.PhysicalStock+it.Quantity); err != nil {
				return nil, err
			}
		}
	}

	if err := u.siRepo.UpdateStatus(ctx, tx, id, newStatus); err != nil {
		return nil, err
	}

	log := &entity.TransactionLog{
		TransactionType: "STOCK_IN",
		TransactionID:   id,
		FromStatus:      txn.Status,
		ToStatus:        newStatus,
		Notes:           fmt.Sprintf("Status changed from %s to %s", txn.Status, newStatus),
	}
	if err := u.logRepo.Create(ctx, tx, log); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	txn.Status = newStatus
	return txn, nil
}

func (u *StockInUseCase) Cancel(ctx context.Context, id uuid.UUID) (*entity.StockInTransaction, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	txn, err := u.siRepo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	if txn.Status == entity.StockInStatusDone {
		return nil, ErrCannotCancelDone
	}
	if txn.Status == entity.StockInStatusCancelled {
		return nil, fmt.Errorf("%w: already cancelled", ErrInvalidTransition)
	}

	if err := u.siRepo.UpdateStatus(ctx, tx, id, entity.StockInStatusCancelled); err != nil {
		return nil, err
	}

	log := &entity.TransactionLog{
		TransactionType: "STOCK_IN",
		TransactionID:   id,
		FromStatus:      txn.Status,
		ToStatus:        entity.StockInStatusCancelled,
		Notes:           "Transaction cancelled",
	}
	if err := u.logRepo.Create(ctx, tx, log); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	txn.Status = entity.StockInStatusCancelled
	return txn, nil
}

func (u *StockInUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.StockInTransaction, error) {
	txn, err := u.siRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	items, _ := u.siRepo.GetItems(ctx, id)
	txn.Items = items
	logs, _ := u.logRepo.GetByTransaction(ctx, "STOCK_IN", id)
	txn.Logs = logs
	return txn, nil
}

func (u *StockInUseCase) List(ctx context.Context, status string, page, limit int) ([]entity.StockInTransaction, int, error) {
	return u.siRepo.List(ctx, status, page, limit)
}
