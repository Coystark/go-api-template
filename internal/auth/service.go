package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/caiohenrique/go-api-template/internal/user"
	"golang.org/x/crypto/bcrypt"
)

// Service contém regras de autenticação.
type Service struct {
	users     user.Repository
	tokens    *TokenManager
	validator *validator.Validator
}

// NewService injeta dependências do domínio auth.
func NewService(users user.Repository, tokens *TokenManager, v *validator.Validator) *Service {
	return &Service{
		users:     users,
		tokens:    tokens,
		validator: v,
	}
}

// Login valida credenciais e retorna JWT.
func (s *Service) Login(ctx context.Context, in LoginDTO) (*TokenResponseDTO, error) {
	if err := s.validator.Struct(in); err != nil {
		return nil, fmt.Errorf("validate login dto: %w", err)
	}

	u, err := s.users.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(u.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &TokenResponseDTO{AccessToken: token}, nil
}
