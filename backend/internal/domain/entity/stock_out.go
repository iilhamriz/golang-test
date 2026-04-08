package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	StockOutStatusDraft      = "DRAFT"
	StockOutStatusInProgress = "IN_PROGRESS"
	StockOutStatusDone       = "DONE"
	StockOutStatusCancelled  = "CANCELLED"
)

var StockOutAllowedTransitions = map[string][]string{
	StockOutStatusDraft:      {StockOutStatusInProgress, StockOutStatusCancelled},
	StockOutStatusInProgress: {StockOutStatusDone, StockOutStatusCancelled},
}

type StockOutTransaction struct {
	ID           uuid.UUID        `json:"id"`
	ReferenceNo  string           `json:"reference_no"`
	CustomerID   *uuid.UUID       `json:"customer_id"`
	CustomerName string           `json:"customer_name,omitempty"`
	Status       string           `json:"status"`
	Notes        string           `json:"notes"`
	CreatedBy    string           `json:"created_by"`
	Items        []StockOutItem   `json:"items,omitempty"`
	Logs         []TransactionLog `json:"logs,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type StockOutItem struct {
	ID                     uuid.UUID `json:"id"`
	StockOutTransactionID  uuid.UUID `json:"stock_out_transaction_id"`
	ItemID                 uuid.UUID `json:"item_id"`
	Quantity               int       `json:"quantity"`
	ItemName               string    `json:"item_name,omitempty"`
	ItemSKU                string    `json:"item_sku,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
}
