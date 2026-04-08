package repository

import (
	"context"
	"fmt"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStockOutRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresStockOutRepo(pool *pgxpool.Pool) *PostgresStockOutRepo {
	return &PostgresStockOutRepo{pool: pool}
}

func (r *PostgresStockOutRepo) Create(ctx context.Context, tx pgx.Tx, txn *entity.StockOutTransaction) error {
	txn.ID = uuid.New()
	return tx.QueryRow(ctx,
		`INSERT INTO stock_out_transactions (id, reference_no, customer_id, status, notes, created_by) VALUES ($1,$2,$3,$4,$5,$6) RETURNING created_at, updated_at`,
		txn.ID, txn.ReferenceNo, txn.CustomerID, txn.Status, txn.Notes, txn.CreatedBy,
	).Scan(&txn.CreatedAt, &txn.UpdatedAt)
}

func (r *PostgresStockOutRepo) CreateItem(ctx context.Context, tx pgx.Tx, item *entity.StockOutItem) error {
	item.ID = uuid.New()
	return tx.QueryRow(ctx,
		`INSERT INTO stock_out_items (id, stock_out_transaction_id, item_id, quantity) VALUES ($1,$2,$3,$4) RETURNING created_at`,
		item.ID, item.StockOutTransactionID, item.ItemID, item.Quantity,
	).Scan(&item.CreatedAt)
}

func (r *PostgresStockOutRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.StockOutTransaction, error) {
	txn := &entity.StockOutTransaction{}
	err := r.pool.QueryRow(ctx,
		`SELECT sot.id, sot.reference_no, sot.customer_id, COALESCE(c.name,''), sot.status, sot.notes, sot.created_by, sot.created_at, sot.updated_at
		 FROM stock_out_transactions sot LEFT JOIN customers c ON c.id = sot.customer_id WHERE sot.id=$1`, id,
	).Scan(&txn.ID, &txn.ReferenceNo, &txn.CustomerID, &txn.CustomerName, &txn.Status, &txn.Notes, &txn.CreatedBy, &txn.CreatedAt, &txn.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

func (r *PostgresStockOutRepo) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.StockOutTransaction, error) {
	txn := &entity.StockOutTransaction{}
	err := tx.QueryRow(ctx,
		`SELECT id, reference_no, customer_id, status, notes, created_by, created_at, updated_at FROM stock_out_transactions WHERE id=$1 FOR UPDATE`, id,
	).Scan(&txn.ID, &txn.ReferenceNo, &txn.CustomerID, &txn.Status, &txn.Notes, &txn.CreatedBy, &txn.CreatedAt, &txn.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

func (r *PostgresStockOutRepo) UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error {
	_, err := tx.Exec(ctx, `UPDATE stock_out_transactions SET status=$1, updated_at=now() WHERE id=$2`, status, id)
	return err
}

func (r *PostgresStockOutRepo) List(ctx context.Context, status string, page, limit int) ([]entity.StockOutTransaction, int, error) {
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
		where = fmt.Sprintf("sot.status=$%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM stock_out_transactions sot WHERE %s", where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(
		`SELECT sot.id, sot.reference_no, sot.customer_id, COALESCE(c.name,''), sot.status, sot.notes, sot.created_by, sot.created_at, sot.updated_at
		 FROM stock_out_transactions sot LEFT JOIN customers c ON c.id = sot.customer_id
		 WHERE %s ORDER BY sot.created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, (page-1)*limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txns []entity.StockOutTransaction
	for rows.Next() {
		var t entity.StockOutTransaction
		if err := rows.Scan(&t.ID, &t.ReferenceNo, &t.CustomerID, &t.CustomerName, &t.Status, &t.Notes, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		txns = append(txns, t)
	}

	return txns, total, nil
}

func (r *PostgresStockOutRepo) GetItems(ctx context.Context, txnID uuid.UUID) ([]entity.StockOutItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT so.id, so.stock_out_transaction_id, so.item_id, so.quantity, i.name, i.sku, so.created_at
		 FROM stock_out_items so JOIN items i ON i.id = so.item_id WHERE so.stock_out_transaction_id=$1`, txnID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.StockOutItem
	for rows.Next() {
		var it entity.StockOutItem
		if err := rows.Scan(&it.ID, &it.StockOutTransactionID, &it.ItemID, &it.Quantity, &it.ItemName, &it.ItemSKU, &it.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}
