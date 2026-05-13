package authapp

import (
	"context"

	userdomain "github.com/caiohenrique/go-api-template/internal/features/user/domain"
	"github.com/google/uuid"
)

// UserFinder resolves users by email (driven port; satisfied by user repo).
type UserFinder interface {
	FindByEmail(ctx context.Context, email string) (*userdomain.User, error)
}

// TokenIssuer issues access tokens (driven port).
type TokenIssuer interface {
	Generate(userID uuid.UUID) (string, error)
}

// PasswordHasher compares plaintext passwords to stored hashes (driven port).
type PasswordHasher interface {
	Compare(hash, plain string) error
}
