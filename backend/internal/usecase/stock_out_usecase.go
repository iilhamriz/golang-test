package usecase

import (
	"context"
	"fmt"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/calyx/smart-inventory/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockOutUseCase struct {
	pool     *pgxpool.Pool
	soRepo   repository.StockOutRepository
	itemRepo repository.ItemRepository
	logRepo  repository.TransactionLogRepository
}

func NewStockOutUseCase(pool *pgxpool.Pool, soRepo repository.StockOutRepository, itemRepo repository.ItemRepository, logRepo repository.TransactionLogRepository) *StockOutUseCase {
	return &StockOutUseCase{pool: pool, soRepo: soRepo, itemRepo: itemRepo, logRepo: logRepo}
}

type CreateStockOutInput struct {
	ReferenceNo string                    `json:"reference_no"`
	CustomerID  *uuid.UUID                `json:"customer_id"`
	Notes       string                    `json:"notes"`
	CreatedBy   string                    `json:"created_by"`
	Items       []CreateStockOutItemInput `json:"items"`
}

type CreateStockOutItemInput struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity int       `json:"quantity"`
}

// CreateDraft implements Phase 1: Allocation
func (u *StockOutUseCase) CreateDraft(ctx context.Context, input CreateStockOutInput) (*entity.StockOutTransaction, error) {
	if input.ReferenceNo == "" || len(input.Items) == 0 {
		return nil, ErrInvalidInput
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Check available stock for each item (with row lock)
	for _, it := range input.Items {
		if it.Quantity <= 0 {
			return nil, ErrInvalidInput
		}
		// Lock item row to prevent concurrent allocation
		_, err := u.itemRepo.GetByIDForUpdate(ctx, tx, it.ItemID)
		if err != nil {
			return nil, fmt.Errorf("item %s not found: %w", it.ItemID, ErrNotFound)
		}
		avail, err := u.itemRepo.GetAvailableStock(ctx, tx, it.ItemID)
		if err != nil {
			return nil, err
		}
		if it.Quantity > avail {
			return nil, fmt.Errorf("%w: item %s needs %d but only %d available", ErrInsufficientStock, it.ItemID, it.Quantity, avail)
		}
	}

	// All checks passed — create transaction
	txn := &entity.StockOutTransaction{
		ReferenceNo: input.ReferenceNo,
		CustomerID:  input.CustomerID,
		Status:      entity.StockOutStatusDraft,
		Notes:       input.Notes,
		CreatedBy:   input.CreatedBy,
	}

	if err := u.soRepo.Create(ctx, tx, txn); err != nil {
		return nil, fmt.Errorf("create stock out: %w", err)
	}

	for _, it := range input.Items {
		item := &entity.StockOutItem{
			StockOutTransactionID: txn.ID,
			ItemID:                it.ItemID,
			Quantity:              it.Quantity,
		}
		if err := u.soRepo.CreateItem(ctx, tx, item); err != nil {
			return nil, fmt.Errorf("create stock out item: %w", err)
		}
		txn.Items = append(txn.Items, *item)
	}

	log := &entity.TransactionLog{
		TransactionType: "STOCK_OUT",
		TransactionID:   txn.ID,
		FromStatus:      "",
		ToStatus:        entity.StockOutStatusDraft,
		Notes:           "Draft created (stock allocated)",
		CreatedBy:       input.CreatedBy,
	}
	if err := u.logRepo.Create(ctx, tx, log); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return txn, nil
}

// UpdateStatus implements Phase 2: DRAFT -> IN_PROGRESS -> DONE
func (u *StockOutUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus string) (*entity.StockOutTransaction, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	txn, err := u.soRepo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	allowed, ok := entity.StockOutAllowedTransitions[txn.Status]
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

	// If transitioning to DONE: decrease physical stock
	if newStatus == entity.StockOutStatusDone {
		items, err := u.soRepo.GetItems(ctx, id)
		if err != nil {
			return nil, err
		}
		for _, it := range items {
			item, err := u.itemRepo.GetByIDForUpdate(ctx, tx, it.ItemID)
			if err != nil {
				return nil, err
			}
			newStock := item.PhysicalStock - it.Quantity
			if newStock < 0 {
				return nil, fmt.Errorf("%w: item %s physical stock would go negative", ErrInsufficientStock, it.ItemID)
			}
			if err := u.itemRepo.UpdatePhysicalStock(ctx, tx, it.ItemID, newStock); err != nil {
				return nil, err
			}
		}
	}

	if err := u.soRepo.UpdateStatus(ctx, tx, id, newStatus); err != nil {
		return nil, err
	}

	log := &entity.TransactionLog{
		TransactionType: "STOCK_OUT",
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

// Cancel implements rollback — just changes status, reservation released automatically
func (u *StockOutUseCase) Cancel(ctx context.Context, id uuid.UUID) (*entity.StockOutTransaction, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	txn, err := u.soRepo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	if txn.Status == entity.StockOutStatusDone {
		return nil, ErrCannotCancelDone
	}
	if txn.Status == entity.StockOutStatusCancelled {
		return nil, fmt.Errorf("%w: already cancelled", ErrInvalidTransition)
	}

	if err := u.soRepo.UpdateStatus(ctx, tx, id, entity.StockOutStatusCancelled); err != nil {
		return nil, err
	}

	log := &entity.TransactionLog{
		TransactionType: "STOCK_OUT",
		TransactionID:   id,
		FromStatus:      txn.Status,
		ToStatus:        entity.StockOutStatusCancelled,
		Notes:           "Transaction cancelled (stock reservation released)",
	}
	if err := u.logRepo.Create(ctx, tx, log); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	txn.Status = entity.StockOutStatusCancelled
	return txn, nil
}

func (u *StockOutUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.StockOutTransaction, error) {
	txn, err := u.soRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	items, _ := u.soRepo.GetItems(ctx, id)
	txn.Items = items
	logs, _ := u.logRepo.GetByTransaction(ctx, "STOCK_OUT", id)
	txn.Logs = logs
	return txn, nil
}

func (u *StockOutUseCase) List(ctx context.Context, status string, page, limit int) ([]entity.StockOutTransaction, int, error) {
	return u.soRepo.List(ctx, status, page, limit)
}
