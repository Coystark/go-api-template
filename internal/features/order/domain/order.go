package orderdomain

import (
	"time"

	"github.com/google/uuid"
)

// Order is the order aggregate root (pure domain).
type Order struct {
	ID            uuid.UUID
	Title         string
	Subject       *string
	Code          *string
	SentAt        *time.Time
	ConvertedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	OrderServices []OrderService
}

// OrderService is a line item belonging to an order.
type OrderService struct {
	ID           uuid.UUID
	OrderID      uuid.UUID
	Title        string
	StartDate    time.Time
	EndDate      *time.Time
	Observations *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
