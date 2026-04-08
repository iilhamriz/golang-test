package entity

import (
	"time"

	"github.com/google/uuid"
)

type TransactionLog struct {
	ID              uuid.UUID `json:"id"`
	TransactionType string    `json:"transaction_type"`
	TransactionID   uuid.UUID `json:"transaction_id"`
	FromStatus      string    `json:"from_status"`
	ToStatus        string    `json:"to_status"`
	Notes           string    `json:"notes"`
	CreatedBy       string    `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
}
