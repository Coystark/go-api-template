package userapp

import (
	"context"
	"errors"
	"fmt"

	userdomain "github.com/caiohenrique/go-api-template/internal/features/user/domain"
	"github.com/caiohenrique/go-api-template/internal/platform/id"
	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/google/uuid"
)

// Service contains user use cases.
type Service struct {
	repo      UserRepository
	validator *validator.Validator
	mailer    Mailer
	hasher    PasswordHasher
}

// NewService wires user application services.
func NewService(repo UserRepository, v *validator.Validator, mailer Mailer, hasher PasswordHasher) *Service {
	return &Service{
		repo:      repo,
		validator: v,
		mailer:    mailer,
		hasher:    hasher,
	}
}

// Create validates, persists, and enqueues welcome email.
func (s *Service) Create(ctx context.Context, in CreateUserInput) (*UserView, error) {
	if err := s.validator.Struct(in); err != nil {
		return nil, fmt.Errorf("validate create user input: %w", err)
	}

	existing, err := s.repo.FindByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, userdomain.ErrNotFound) {
		return nil, fmt.Errorf("check email exists: %w", err)
	}
	if existing != nil {
		return nil, userdomain.ErrEmailAlreadyExists
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	uid, err := id.New()
	if err != nil {
		return nil, fmt.Errorf("generate user id: %w", err)
	}

	u := &userdomain.User{
		ID:           uid,
		Email:        in.Email,
		PasswordHash: hash,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("persist user: %w", err)
	}

	if err := s.mailer.SendWelcome(ctx, u.Email); err != nil {
		return nil, fmt.Errorf("enqueue welcome email: %w", err)
	}

	out := toUserView(u)
	return &out, nil
}

// GetByID returns a user by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*UserView, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := toUserView(u)
	return &out, nil
}

func toUserView(u *userdomain.User) UserView {
	return UserView{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
