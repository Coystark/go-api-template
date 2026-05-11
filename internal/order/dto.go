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
// No update, `ID` presente indica UPDATE; ausente indica INSERT.
// No create, `ID` é ignorado.
type OrderServiceInput struct {
	ID    *uuid.UUID `json:"id,omitempty"`
	Title string     `json:"title" validate:"required"`
}

// CreateOrderDTO representa entrada para criação de pedido.
type CreateOrderDTO struct {
	Title         string              `json:"title" validate:"required"`
	Subject       *string             `json:"subject,omitempty" validate:"omitempty,min=1"`
	Code          *string             `json:"code,omitempty" validate:"omitempty,min=1"`
	SentAt        *time.Time          `json:"sent_at,omitempty"`
	ConvertedAt   *time.Time          `json:"converted_at,omitempty"`
	OrderServices []OrderServiceInput `json:"order_services" validate:"dive"`
}

// UpdateOrderDTO representa entrada para atualização parcial de pedido.
// Itens em `OrderServices` com `ID` são atualizados; sem `ID` são criados.
// IDs em `RemovedServiceIDs` são removidos na mesma transação.
// Serviços omitidos permanecem inalterados.
type UpdateOrderDTO struct {
	Title             *string             `json:"title,omitempty" validate:"omitempty,min=1"`
	Subject           *string             `json:"subject,omitempty" validate:"omitempty,min=1"`
	Code              *string             `json:"code,omitempty" validate:"omitempty,min=1"`
	SentAt            *time.Time          `json:"sent_at,omitempty"`
	ConvertedAt       *time.Time          `json:"converted_at,omitempty"`
	OrderServices     []OrderServiceInput `json:"order_services" validate:"dive"`
	RemovedServiceIDs []uuid.UUID         `json:"removed_service_ids"`
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
	Subject       *string                `json:"subject,omitempty"`
	Code          *string                `json:"code,omitempty"`
	SentAt        *time.Time             `json:"sent_at,omitempty"`
	ConvertedAt   *time.Time             `json:"converted_at,omitempty"`
	OrderServices []OrderServiceResponse `json:"order_services"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}
