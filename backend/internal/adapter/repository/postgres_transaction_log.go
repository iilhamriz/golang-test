package repository

import (
	"context"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTransactionLogRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresTransactionLogRepo(pool *pgxpool.Pool) *PostgresTransactionLogRepo {
	return &PostgresTransactionLogRepo{pool: pool}
}

func (r *PostgresTransactionLogRepo) Create(ctx context.Context, tx pgx.Tx, log *entity.TransactionLog) error {
	log.ID = uuid.New()
	return tx.QueryRow(ctx,
		`INSERT INTO transaction_logs (id, transaction_type, transaction_id, from_status, to_status, notes, created_by) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING created_at`,
		log.ID, log.TransactionType, log.TransactionID, log.FromStatus, log.ToStatus, log.Notes, log.CreatedBy,
	).Scan(&log.CreatedAt)
}

func (r *PostgresTransactionLogRepo) GetByTransaction(ctx context.Context, txnType string, txnID uuid.UUID) ([]entity.TransactionLog, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, transaction_type, transaction_id, from_status, to_status, notes, created_by, created_at
		 FROM transaction_logs WHERE transaction_type=$1 AND transaction_id=$2 ORDER BY created_at ASC`, txnType, txnID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []entity.TransactionLog
	for rows.Next() {
		var l entity.TransactionLog
		if err := rows.Scan(&l.ID, &l.TransactionType, &l.TransactionID, &l.FromStatus, &l.ToStatus, &l.Notes, &l.CreatedBy, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
