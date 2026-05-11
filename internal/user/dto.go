package user

import (
	"time"

	"github.com/google/uuid"
)

// CreateDTO representa entrada para criação de usuário.
type CreateDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateDTO reservado para evoluções futuras do template.
type UpdateDTO struct {
	Email string `json:"email" validate:"omitempty,email"`
}

// ResponseDTO é a visão pública do usuário (sem senha).
type ResponseDTO struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
