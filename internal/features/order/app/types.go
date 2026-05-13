package orderapp

import (
	"time"

	"github.com/caiohenrique/go-api-template/internal/platform/pagination"
	"github.com/google/uuid"
)

// OrderServiceItem represents a service line in create/update bodies.
// On update, a present ID means UPDATE; absent means INSERT.
// On create, ID is ignored. On UPDATE, omitted or null end_date and observations persist as NULL.
type OrderServiceItem struct {
	ID           *uuid.UUID `json:"id,omitempty"`
	Title        string     `json:"title" validate:"required"`
	StartDate    time.Time  `json:"start_date" validate:"required"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Observations *string    `json:"observations,omitempty" validate:"omitempty,min=1"`
}

// CreateOrderInput is input for creating an order.
type CreateOrderInput struct {
	Title         string             `json:"title" validate:"required"`
	Subject       *string            `json:"subject,omitempty" validate:"omitempty,min=1"`
	Code          *string            `json:"code,omitempty" validate:"omitempty,min=1"`
	SentAt        *time.Time         `json:"sent_at,omitempty"`
	ConvertedAt   *time.Time         `json:"converted_at,omitempty"`
	OrderServices []OrderServiceItem `json:"order_services" validate:"dive"`
}

// UpdateOrderInput is input for updating an order (PUT).
type UpdateOrderInput struct {
	Title             string             `json:"title" validate:"required,min=1"`
	Subject           *string            `json:"subject,omitempty" validate:"omitempty,min=1"`
	Code              *string            `json:"code,omitempty" validate:"omitempty,min=1"`
	SentAt            *time.Time         `json:"sent_at,omitempty"`
	ConvertedAt       *time.Time         `json:"converted_at,omitempty"`
	OrderServices     []OrderServiceItem `json:"order_services" validate:"dive"`
	RemovedServiceIDs []uuid.UUID        `json:"removed_service_ids"`
}

// ListOrdersInput is query input for listing orders.
type ListOrdersInput struct {
	pagination.Query
	Code  string `form:"code" validate:"omitempty,max=255"`
	Title string `form:"title" validate:"omitempty,max=255"`
}

// ListOrderServicesInput is query input for listing order services.
type ListOrderServicesInput struct {
	pagination.Query
	OrderID string `form:"order_id" validate:"omitempty,uuid"`
}

// OrderServiceView is the public read model for an order service line.
type OrderServiceView struct {
	ID           uuid.UUID  `json:"id"`
	OrderID      uuid.UUID  `json:"order_id"`
	Title        string     `json:"title"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Observations *string    `json:"observations,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// OrderView is the public read model for an order.
type OrderView struct {
	ID            uuid.UUID          `json:"id"`
	Title         string             `json:"title"`
	Subject       *string            `json:"subject,omitempty"`
	Code          *string            `json:"code,omitempty"`
	SentAt        *time.Time         `json:"sent_at,omitempty"`
	ConvertedAt   *time.Time         `json:"converted_at,omitempty"`
	OrderServices []OrderServiceView `json:"order_services"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// OrdersPage is a paginated list of orders.
type OrdersPage struct {
	Items []OrderView `json:"items"`
	pagination.Meta
}

// OrderServicesPage is a paginated list of order service rows.
type OrderServicesPage struct {
	Items []OrderServiceView `json:"items"`
	pagination.Meta
}
