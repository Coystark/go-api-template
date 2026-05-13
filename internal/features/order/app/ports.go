package orderapp

import (
	"context"
	"time"

	orderdomain "github.com/caiohenrique/go-api-template/internal/features/order/domain"
	"github.com/caiohenrique/go-api-template/internal/platform/pagination"
	"github.com/google/uuid"
)

// UpdateOrderRepoInput aggregates header and line changes for a single repository transaction.
type UpdateOrderRepoInput struct {
	Title       string
	Subject     *string
	Code        *string
	SentAt      *time.Time
	ConvertedAt *time.Time
	Creates     []orderdomain.OrderService
	Updates     []orderdomain.OrderService
	RemoveIDs   []uuid.UUID
}

// ListOrdersRepoInput lists orders with pagination and optional filters.
type ListOrdersRepoInput struct {
	pagination.Query
	Code  string
	Title string
}

// ListOrderServicesRepoInput lists order service rows with optional order filter.
type ListOrderServicesRepoInput struct {
	pagination.Query
	OrderID *uuid.UUID
}

// OrderRepository persists orders (driven port).
type OrderRepository interface {
	Create(ctx context.Context, o *orderdomain.Order, services []orderdomain.OrderService) error
	FindByID(ctx context.Context, id uuid.UUID) (*orderdomain.Order, error)
	List(ctx context.Context, in ListOrdersRepoInput) ([]orderdomain.Order, int64, error)
	ListOrderServices(ctx context.Context, in ListOrderServicesRepoInput) ([]orderdomain.OrderService, int64, error)
	Update(ctx context.Context, orderID uuid.UUID, in UpdateOrderRepoInput) error
	Delete(ctx context.Context, id uuid.UUID) error
}
