package domain

import (
	"time"
)

type Product struct {
	ID          string    `json:"id"`
	SKU         string    `json:"sku" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Stock uses version-based optimistic locking to prevent race conditions
type Stock struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id" validate:"required"`
	Quantity  int64     `json:"quantity" validate:"gte=0"`
	Version   int64     `json:"version"` // Incremented on every update
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StockMutation provides audit trail for all inventory changes
type StockMutation struct {
	ID            string        `json:"id"`
	ProductID     string        `json:"product_id" validate:"required"`
	MutationType  MutationType  `json:"mutation_type" validate:"required"`
	Quantity      int64         `json:"quantity" validate:"required"`
	PreviousQty   int64         `json:"previous_qty"`
	NewQty        int64         `json:"new_qty"`
	ReferenceID   string        `json:"reference_id"`
	ReferenceType ReferenceType `json:"reference_type"`
	Notes         string        `json:"notes"`
	CreatedBy     string        `json:"created_by"`
	CreatedAt     time.Time     `json:"created_at"`
}

type MutationType string

const (
	MutationTypeInbound     MutationType = "inbound"
	MutationTypeOutbound    MutationType = "outbound"
	MutationTypeAdjustment  MutationType = "adjustment"
	MutationTypeReservation MutationType = "reservation"
	MutationTypeRelease     MutationType = "release"
)

type ReferenceType string

const (
	ReferenceTypeOrder      ReferenceType = "order"
	ReferenceTypePurchase   ReferenceType = "purchase"
	ReferenceTypeAdjustment ReferenceType = "adjustment"
)
