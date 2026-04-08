package entity

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID             uuid.UUID `json:"id"`
	SKU            string    `json:"sku"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	PhysicalStock  int       `json:"physical_stock"`
	AvailableStock int       `json:"available_stock"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
