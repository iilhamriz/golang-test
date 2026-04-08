package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	StockInStatusCreated    = "CREATED"
	StockInStatusInProgress = "IN_PROGRESS"
	StockInStatusDone       = "DONE"
	StockInStatusCancelled  = "CANCELLED"
)

var StockInAllowedTransitions = map[string][]string{
	StockInStatusCreated:    {StockInStatusInProgress, StockInStatusCancelled},
	StockInStatusInProgress: {StockInStatusDone, StockInStatusCancelled},
}

type StockInTransaction struct {
	ID          uuid.UUID      `json:"id"`
	ReferenceNo string         `json:"reference_no"`
	Status      string         `json:"status"`
	Notes       string         `json:"notes"`
	CreatedBy   string         `json:"created_by"`
	Items       []StockInItem  `json:"items,omitempty"`
	Logs        []TransactionLog `json:"logs,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type StockInItem struct {
	ID                    uuid.UUID `json:"id"`
	StockInTransactionID  uuid.UUID `json:"stock_in_transaction_id"`
	ItemID                uuid.UUID `json:"item_id"`
	Quantity              int       `json:"quantity"`
	ItemName              string    `json:"item_name,omitempty"`
	ItemSKU               string    `json:"item_sku,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}
