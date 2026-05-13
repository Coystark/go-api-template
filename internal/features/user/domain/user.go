package userdomain

import (
	"time"

	"github.com/google/uuid"
)

// User is the user aggregate root (pure domain; no persistence tags).
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
