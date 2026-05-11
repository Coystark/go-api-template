package order

import (
	"time"

	"github.com/google/uuid"
)

// ErrorResponse representa erro JSON da API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// OrderServiceInput representa um serviço no corpo de create/update.
type OrderServiceInput struct {
	Title string `json:"title" validate:"required"`
}

// CreateOrderDTO representa entrada para criação de pedido.
type CreateOrderDTO struct {
	Title         string              `json:"title" validate:"required"`
	OrderServices []OrderServiceInput `json:"order_services" validate:"dive"`
}

// UpdateOrderDTO representa entrada para atualização de pedido.
type UpdateOrderDTO struct {
	Title         string              `json:"title" validate:"required"`
	OrderServices []OrderServiceInput `json:"order_services" validate:"dive"`
}

// OrderServiceResponse é a visão pública de um serviço do pedido.
type OrderServiceResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResponseDTO é a visão pública do pedido.
type ResponseDTO struct {
	ID            uuid.UUID              `json:"id"`
	Title         string                 `json:"title"`
	OrderServices []OrderServiceResponse `json:"order_services"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}
