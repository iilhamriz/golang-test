package usecase

import (
	"context"
	"fmt"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/calyx/smart-inventory/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryUseCase struct {
	pool    *pgxpool.Pool
	itemRepo repository.ItemRepository
	adjRepo  repository.StockAdjustmentRepository
}

func NewInventoryUseCase(pool *pgxpool.Pool, itemRepo repository.ItemRepository, adjRepo repository.StockAdjustmentRepository) *InventoryUseCase {
	return &InventoryUseCase{pool: pool, itemRepo: itemRepo, adjRepo: adjRepo}
}

type CreateItemInput struct {
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (u *InventoryUseCase) CreateItem(ctx context.Context, input CreateItemInput) (*entity.Item, error) {
	if input.SKU == "" || input.Name == "" {
		return nil, ErrInvalidInput
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	item := &entity.Item{
		SKU:         input.SKU,
		Name:        input.Name,
		Description: input.Description,
	}

	if err := u.itemRepo.Create(ctx, tx, item); err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return item, nil
}

type UpdateItemInput struct {
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (u *InventoryUseCase) UpdateItem(ctx context.Context, id uuid.UUID, input UpdateItemInput) (*entity.Item, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	item, err := u.itemRepo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	if input.SKU != "" {
		item.SKU = input.SKU
	}
	if input.Name != "" {
		item.Name = input.Name
	}
	item.Description = input.Description

	if err := u.itemRepo.Update(ctx, tx, item); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return u.itemRepo.GetByID(ctx, id)
}

func (u *InventoryUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Item, error) {
	return u.itemRepo.GetByID(ctx, id)
}

func (u *InventoryUseCase) List(ctx context.Context, filter repository.ItemFilter) ([]entity.Item, int, error) {
	return u.itemRepo.List(ctx, filter)
}

type AdjustStockInput struct {
	Quantity  int    `json:"quantity"`
	Reason    string `json:"reason"`
	CreatedBy string `json:"created_by"`
}

func (u *InventoryUseCase) AdjustStock(ctx context.Context, itemID uuid.UUID, input AdjustStockInput) (*entity.Item, error) {
	if input.Reason == "" {
		return nil, ErrInvalidInput
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	item, err := u.itemRepo.GetByIDForUpdate(ctx, tx, itemID)
	if err != nil {
		return nil, ErrNotFound
	}

	newStock := item.PhysicalStock + input.Quantity
	if newStock < 0 {
		return nil, ErrStockCannotBeNeg
	}

	if err := u.itemRepo.UpdatePhysicalStock(ctx, tx, itemID, newStock); err != nil {
		return nil, err
	}

	adj := &entity.StockAdjustment{
		ItemID:    itemID,
		Quantity:  input.Quantity,
		Reason:    input.Reason,
		CreatedBy: input.CreatedBy,
	}
	if err := u.adjRepo.Create(ctx, tx, adj); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return u.itemRepo.GetByID(ctx, itemID)
}
