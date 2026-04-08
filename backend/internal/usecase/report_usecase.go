package usecase

import (
	"context"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/calyx/smart-inventory/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportUseCase struct {
	pool    *pgxpool.Pool
	siRepo  repository.StockInRepository
	soRepo  repository.StockOutRepository
	logRepo repository.TransactionLogRepository
}

func NewReportUseCase(pool *pgxpool.Pool, siRepo repository.StockInRepository, soRepo repository.StockOutRepository, logRepo repository.TransactionLogRepository) *ReportUseCase {
	return &ReportUseCase{pool: pool, siRepo: siRepo, soRepo: soRepo, logRepo: logRepo}
}

type ReportTransaction struct {
	Type        string `json:"type"`
	ID          uuid.UUID `json:"id"`
	ReferenceNo string    `json:"reference_no"`
	Status      string    `json:"status"`
	Notes       string    `json:"notes"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

func (u *ReportUseCase) ListDoneTransactions(ctx context.Context, txnType string, page, limit int) ([]ReportTransaction, int, error) {
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }

	var results []ReportTransaction
	var total int

	if txnType == "" || txnType == "stock-in" {
		txns, cnt, err := u.siRepo.List(ctx, entity.StockInStatusDone, page, limit)
		if err != nil {
			return nil, 0, err
		}
		total += cnt
		for _, t := range txns {
			results = append(results, ReportTransaction{
				Type: "STOCK_IN", ID: t.ID, ReferenceNo: t.ReferenceNo,
				Status: t.Status, Notes: t.Notes,
				CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	if txnType == "" || txnType == "stock-out" {
		txns, cnt, err := u.soRepo.List(ctx, entity.StockOutStatusDone, page, limit)
		if err != nil {
			return nil, 0, err
		}
		total += cnt
		for _, t := range txns {
			results = append(results, ReportTransaction{
				Type: "STOCK_OUT", ID: t.ID, ReferenceNo: t.ReferenceNo,
				Status: t.Status, Notes: t.Notes,
				CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	return results, total, nil
}

type ReportDetail struct {
	Transaction interface{}          `json:"transaction"`
	Logs        []entity.TransactionLog `json:"logs"`
}

func (u *ReportUseCase) GetTransactionDetail(ctx context.Context, txnType string, id uuid.UUID) (*ReportDetail, error) {
	switch txnType {
	case "stock-in":
		txn, err := u.siRepo.GetByID(ctx, id)
		if err != nil {
			return nil, ErrNotFound
		}
		items, _ := u.siRepo.GetItems(ctx, id)
		txn.Items = items
		logs, _ := u.logRepo.GetByTransaction(ctx, "STOCK_IN", id)
		return &ReportDetail{Transaction: txn, Logs: logs}, nil

	case "stock-out":
		txn, err := u.soRepo.GetByID(ctx, id)
		if err != nil {
			return nil, ErrNotFound
		}
		items, _ := u.soRepo.GetItems(ctx, id)
		txn.Items = items
		logs, _ := u.logRepo.GetByTransaction(ctx, "STOCK_OUT", id)
		return &ReportDetail{Transaction: txn, Logs: logs}, nil

	default:
		return nil, ErrInvalidInput
	}
}
