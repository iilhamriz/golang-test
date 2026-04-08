package repository

import (
	"context"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCustomerRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresCustomerRepo(pool *pgxpool.Pool) *PostgresCustomerRepo {
	return &PostgresCustomerRepo{pool: pool}
}

func (r *PostgresCustomerRepo) Create(ctx context.Context, c *entity.Customer) error {
	c.ID = uuid.New()
	return r.pool.QueryRow(ctx,
		`INSERT INTO customers (id, name, email, phone, address) VALUES ($1,$2,$3,$4,$5) RETURNING created_at, updated_at`,
		c.ID, c.Name, c.Email, c.Phone, c.Address,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
}

func (r *PostgresCustomerRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error) {
	c := &entity.Customer{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, phone, address, created_at, updated_at FROM customers WHERE id=$1`, id,
	).Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Address, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *PostgresCustomerRepo) List(ctx context.Context, page, limit int) ([]entity.Customer, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM customers`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, name, email, phone, address, created_at, updated_at FROM customers ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, (page-1)*limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var customers []entity.Customer
	for rows.Next() {
		var c entity.Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Address, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		customers = append(customers, c)
	}

	return customers, total, nil
}
