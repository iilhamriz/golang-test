package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	repo "github.com/calyx/smart-inventory/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresItemRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresItemRepo(pool *pgxpool.Pool) *PostgresItemRepo {
	return &PostgresItemRepo{pool: pool}
}

func (r *PostgresItemRepo) Create(ctx context.Context, tx pgx.Tx, item *entity.Item) error {
	item.ID = uuid.New()
	return tx.QueryRow(ctx,
		`INSERT INTO items (id, sku, name, description, physical_stock) VALUES ($1,$2,$3,$4,$5) RETURNING created_at, updated_at`,
		item.ID, item.SKU, item.Name, item.Description, item.PhysicalStock,
	).Scan(&item.CreatedAt, &item.UpdatedAt)
}

func (r *PostgresItemRepo) Update(ctx context.Context, tx pgx.Tx, item *entity.Item) error {
	_, err := tx.Exec(ctx,
		`UPDATE items SET sku=$1, name=$2, description=$3, updated_at=now() WHERE id=$4`,
		item.SKU, item.Name, item.Description, item.ID,
	)
	return err
}

func (r *PostgresItemRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Item, error) {
	item := &entity.Item{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, sku, name, description, physical_stock, available_stock, created_at, updated_at FROM v_item_stock WHERE id=$1`, id,
	).Scan(&item.ID, &item.SKU, &item.Name, &item.Description, &item.PhysicalStock, &item.AvailableStock, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresItemRepo) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.Item, error) {
	item := &entity.Item{}
	err := tx.QueryRow(ctx,
		`SELECT id, sku, name, description, physical_stock, created_at, updated_at FROM items WHERE id=$1 FOR UPDATE`, id,
	).Scan(&item.ID, &item.SKU, &item.Name, &item.Description, &item.PhysicalStock, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresItemRepo) List(ctx context.Context, filter repo.ItemFilter) ([]entity.Item, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filter.Name != "" {
		where = append(where, fmt.Sprintf("v.name ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Name+"%")
		argIdx++
	}
	if filter.SKU != "" {
		where = append(where, fmt.Sprintf("v.sku ILIKE $%d", argIdx))
		args = append(args, "%"+filter.SKU+"%")
		argIdx++
	}
	if filter.CustomerID != nil {
		where = append(where, fmt.Sprintf(`v.id IN (
			SELECT soi.item_id FROM stock_out_items soi
			JOIN stock_out_transactions sot ON sot.id = soi.stock_out_transaction_id
			WHERE sot.customer_id = $%d
		)`, argIdx))
		args = append(args, *filter.CustomerID)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM v_item_stock v WHERE %s", whereClause)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}
	offset := (filter.Page - 1) * filter.Limit

	query := fmt.Sprintf(
		"SELECT v.id, v.sku, v.name, v.description, v.physical_stock, v.available_stock, v.created_at, v.updated_at FROM v_item_stock v WHERE %s ORDER BY v.created_at DESC LIMIT $%d OFFSET $%d",
		whereClause, argIdx, argIdx+1,
	)
	args = append(args, filter.Limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []entity.Item
	for rows.Next() {
		var it entity.Item
		if err := rows.Scan(&it.ID, &it.SKU, &it.Name, &it.Description, &it.PhysicalStock, &it.AvailableStock, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}

	return items, total, nil
}

func (r *PostgresItemRepo) UpdatePhysicalStock(ctx context.Context, tx pgx.Tx, id uuid.UUID, newStock int) error {
	_, err := tx.Exec(ctx, `UPDATE items SET physical_stock=$1, updated_at=now() WHERE id=$2`, newStock, id)
	return err
}

func (r *PostgresItemRepo) GetAvailableStock(ctx context.Context, tx pgx.Tx, id uuid.UUID) (int, error) {
	var avail int
	err := tx.QueryRow(ctx, `SELECT available_stock FROM v_item_stock WHERE id=$1`, id).Scan(&avail)
	return avail, err
}
