package authapp

import (
	"context"
	"errors"
	"fmt"

	authdomain "github.com/caiohenrique/go-api-template/internal/features/auth/domain"
	userdomain "github.com/caiohenrique/go-api-template/internal/features/user/domain"
	"github.com/caiohenrique/go-api-template/internal/platform/validator"
)

// Service contains authentication use cases.
type Service struct {
	users     UserFinder
	tokens    TokenIssuer
	hasher    PasswordHasher
	validator *validator.Validator
}

// NewService wires auth application services.
func NewService(users UserFinder, tokens TokenIssuer, hasher PasswordHasher, v *validator.Validator) *Service {
	return &Service{
		users:     users,
		tokens:    tokens,
		hasher:    hasher,
		validator: v,
	}
}

// Login validates credentials and returns a JWT.
func (s *Service) Login(ctx context.Context, in LoginInput) (*TokenView, error) {
	if err := s.validator.Struct(in); err != nil {
		return nil, fmt.Errorf("validate login input: %w", err)
	}

	u, err := s.users.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, userdomain.ErrNotFound) {
			return nil, authdomain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := s.hasher.Compare(u.PasswordHash, in.Password); err != nil {
		return nil, authdomain.ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(u.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &TokenView{AccessToken: token}, nil
}
