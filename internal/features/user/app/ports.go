package userapp

import (
	"context"

	userdomain "github.com/caiohenrique/go-api-template/internal/features/user/domain"
	"github.com/google/uuid"
)

// UserRepository persists users (driven port).
type UserRepository interface {
	Create(ctx context.Context, u *userdomain.User) error
	FindByEmail(ctx context.Context, email string) (*userdomain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*userdomain.User, error)
}

// Mailer sends async notifications (driven port).
type Mailer interface {
	SendWelcome(ctx context.Context, email string) error
}

// PasswordHasher hashes passwords for storage (driven port).
type PasswordHasher interface {
	Hash(plain string) (string, error)
}
