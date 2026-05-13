package userapp

import (
	"time"

	"github.com/google/uuid"
)

// CreateUserInput is input for creating a user.
type CreateUserInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateUserInput is reserved for future user updates in the template.
type UpdateUserInput struct {
	Email string `json:"email" validate:"omitempty,email"`
}

// UserView is the public read model (no password).
type UserView struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
