package repository

import (
	"context"
	"fmt"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStockInRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresStockInRepo(pool *pgxpool.Pool) *PostgresStockInRepo {
	return &PostgresStockInRepo{pool: pool}
}

func (r *PostgresStockInRepo) Create(ctx context.Context, tx pgx.Tx, txn *entity.StockInTransaction) error {
	txn.ID = uuid.New()
	return tx.QueryRow(ctx,
		`INSERT INTO stock_in_transactions (id, reference_no, status, notes, created_by) VALUES ($1,$2,$3,$4,$5) RETURNING created_at, updated_at`,
		txn.ID, txn.ReferenceNo, txn.Status, txn.Notes, txn.CreatedBy,
	).Scan(&txn.CreatedAt, &txn.UpdatedAt)
}

func (r *PostgresStockInRepo) CreateItem(ctx context.Context, tx pgx.Tx, item *entity.StockInItem) error {
	item.ID = uuid.New()
	return tx.QueryRow(ctx,
		`INSERT INTO stock_in_items (id, stock_in_transaction_id, item_id, quantity) VALUES ($1,$2,$3,$4) RETURNING created_at`,
		item.ID, item.StockInTransactionID, item.ItemID, item.Quantity,
	).Scan(&item.CreatedAt)
}

func (r *PostgresStockInRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.StockInTransaction, error) {
	txn := &entity.StockInTransaction{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, reference_no, status, notes, created_by, created_at, updated_at FROM stock_in_transactions WHERE id=$1`, id,
	).Scan(&txn.ID, &txn.ReferenceNo, &txn.Status, &txn.Notes, &txn.CreatedBy, &txn.CreatedAt, &txn.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

func (r *PostgresStockInRepo) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.StockInTransaction, error) {
	txn := &entity.StockInTransaction{}
	err := tx.QueryRow(ctx,
		`SELECT id, reference_no, status, notes, created_by, created_at, updated_at FROM stock_in_transactions WHERE id=$1 FOR UPDATE`, id,
	).Scan(&txn.ID, &txn.ReferenceNo, &txn.Status, &txn.Notes, &txn.CreatedBy, &txn.CreatedAt, &txn.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

func (r *PostgresStockInRepo) UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error {
	_, err := tx.Exec(ctx, `UPDATE stock_in_transactions SET status=$1, updated_at=now() WHERE id=$2`, status, id)
	return err
}

func (r *PostgresStockInRepo) List(ctx context.Context, status string, page, limit int) ([]entity.StockInTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	argIdx := 1
	if status != "" {
		where = fmt.Sprintf("status=$%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM stock_in_transactions WHERE %s", where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(
		"SELECT id, reference_no, status, notes, created_by, created_at, updated_at FROM stock_in_transactions WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, (page-1)*limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txns []entity.StockInTransaction
	for rows.Next() {
		var t entity.StockInTransaction
		if err := rows.Scan(&t.ID, &t.ReferenceNo, &t.Status, &t.Notes, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		txns = append(txns, t)
	}

	return txns, total, nil
}

func (r *PostgresStockInRepo) GetItems(ctx context.Context, txnID uuid.UUID) ([]entity.StockInItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT si.id, si.stock_in_transaction_id, si.item_id, si.quantity, i.name, i.sku, si.created_at
		 FROM stock_in_items si JOIN items i ON i.id = si.item_id WHERE si.stock_in_transaction_id=$1`, txnID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.StockInItem
	for rows.Next() {
		var it entity.StockInItem
		if err := rows.Scan(&it.ID, &it.StockInTransactionID, &it.ItemID, &it.Quantity, &it.ItemName, &it.ItemSKU, &it.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}
