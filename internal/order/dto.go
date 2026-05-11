package order

import (
	"time"

	"github.com/caiohenrique/go-api-template/internal/platform/pagination"
	"github.com/google/uuid"
)

// OrderServiceInput representa um serviço no corpo de create/update.
// No update, `ID` presente indica UPDATE; ausente indica INSERT.
// No create, `ID` é ignorado. Em UPDATE, `end_date` e `observations` omitidos ou null
// gravam NULL; reenvie os valores para mantê-los.
type OrderServiceInput struct {
	ID           *uuid.UUID `json:"id,omitempty"`
	Title        string     `json:"title" validate:"required"`
	StartDate    time.Time  `json:"start_date" validate:"required"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Observations *string    `json:"observations,omitempty" validate:"omitempty,min=1"`
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

// UpdateOrderDTO representa entrada para atualização do pedido (PUT).
// Cabeçalho: replace total — `title` obrigatório; `subject`, `code`, `sent_at` e `converted_at`
// omitidos ou `null` gravam NULL (reenvie valores para mantê-los).
// Itens em `OrderServices` com `ID` são UPDATE; sem `ID` são INSERT. `RemovedServiceIDs` remove na mesma transação.
// Linhas não listadas em `OrderServices` permanecem inalteradas.
type UpdateOrderDTO struct {
	Title             string              `json:"title" validate:"required,min=1"`
	Subject           *string             `json:"subject,omitempty" validate:"omitempty,min=1"`
	Code              *string             `json:"code,omitempty" validate:"omitempty,min=1"`
	SentAt            *time.Time          `json:"sent_at,omitempty"`
	ConvertedAt       *time.Time          `json:"converted_at,omitempty"`
	OrderServices     []OrderServiceInput `json:"order_services" validate:"dive"`
	RemovedServiceIDs []uuid.UUID         `json:"removed_service_ids"`
}

// ListQueryDTO representa parâmetros de paginação para listagem de pedidos.
type ListQueryDTO = pagination.Query

// ListResponseDTO representa uma página de pedidos.
type ListResponseDTO struct {
	Items []ResponseDTO `json:"items"`
	pagination.Meta
}

// OrderServiceResponse é a visão pública de um serviço do pedido.
type OrderServiceResponse struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Observations *string    `json:"observations,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
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
