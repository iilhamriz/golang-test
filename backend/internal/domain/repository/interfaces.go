package repository

import (
	"context"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ItemRepository interface {StockOutUseCase
	Create(ctx context.Context, tx pgx.Tx, item *entity.Item) error
	Update(ctx context.Context, tx pgx.Tx, item *entity.Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Item, error)
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.Item, error)
	List(ctx context.Context, filter ItemFilter) ([]entity.Item, int, error)
	UpdatePhysicalStock(ctx context.Context, tx pgx.Tx, id uuid.UUID, newStock int) error
	GetAvailableStock(ctx context.Context, tx pgx.Tx, id uuid.UUID) (int, error)
}

type ItemFilter struct {
	Name       string
	SKU        string
	CustomerID *uuid.UUID
	Page       int
	Limit      int
}

type CustomerRepository interface {
	Create(ctx context.Context, customer *entity.Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error)
	List(ctx context.Context, page, limit int) ([]entity.Customer, int, error)
}

type StockInRepository interface {
	Create(ctx context.Context, tx pgx.Tx, txn *entity.StockInTransaction) error
	CreateItem(ctx context.Context, tx pgx.Tx, item *entity.StockInItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.StockInTransaction, error)
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.StockInTransaction, error)
	UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error
	List(ctx context.Context, status string, page, limit int) ([]entity.StockInTransaction, int, error)
	GetItems(ctx context.Context, txnID uuid.UUID) ([]entity.StockInItem, error)
}

type StockOutRepository interface {
	Create(ctx context.Context, tx pgx.Tx, txn *entity.StockOutTransaction) error
	CreateItem(ctx context.Context, tx pgx.Tx, item *entity.StockOutItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.StockOutTransaction, error)
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.StockOutTransaction, error)
	UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error
	List(ctx context.Context, status string, page, limit int) ([]entity.StockOutTransaction, int, error)
	GetItems(ctx context.Context, txnID uuid.UUID) ([]entity.StockOutItem, error)
}

type TransactionLogRepository interface {
	Create(ctx context.Context, tx pgx.Tx, log *entity.TransactionLog) error
	GetByTransaction(ctx context.Context, txnType string, txnID uuid.UUID) ([]entity.TransactionLog, error)
}

type StockAdjustmentRepository interface {
	Create(ctx context.Context, tx pgx.Tx, adj *entity.StockAdjustment) error
	GetByItemID(ctx context.Context, itemID uuid.UUID) ([]entity.StockAdjustment, error)
}

type ReportRepository interface {
	ListDoneTransactions(ctx context.Context, txnType string, page, limit int) ([]entity.TransactionLog, int, error)
}
