package repository

import (
	"context"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStockAdjustmentRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresStockAdjustmentRepo(pool *pgxpool.Pool) *PostgresStockAdjustmentRepo {
	return &PostgresStockAdjustmentRepo{pool: pool}
}

func (r *PostgresStockAdjustmentRepo) Create(ctx context.Context, tx pgx.Tx, adj *entity.StockAdjustment) error {
	adj.ID = uuid.New()
	return tx.QueryRow(ctx,
		`INSERT INTO stock_adjustments (id, item_id, quantity, reason, created_by) VALUES ($1,$2,$3,$4,$5) RETURNING created_at`,
		adj.ID, adj.ItemID, adj.Quantity, adj.Reason, adj.CreatedBy,
	).Scan(&adj.CreatedAt)
}

func (r *PostgresStockAdjustmentRepo) GetByItemID(ctx context.Context, itemID uuid.UUID) ([]entity.StockAdjustment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, item_id, quantity, reason, created_by, created_at FROM stock_adjustments WHERE item_id=$1 ORDER BY created_at DESC`, itemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adjs []entity.StockAdjustment
	for rows.Next() {
		var a entity.StockAdjustment
		if err := rows.Scan(&a.ID, &a.ItemID, &a.Quantity, &a.Reason, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		adjs = append(adjs, a)
	}
	return adjs, nil
}
