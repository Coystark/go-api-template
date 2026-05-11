package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/caiohenrique/go-api-template/internal/platform/queue"
	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"golang.org/x/crypto/bcrypt"
)

// taskEnqueuer abstrai o cliente asynq para testes.
type taskEnqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// Service contém regras de negócio do domínio de usuário.
type Service struct {
	repo      Repository
	validator *validator.Validator
	enqueue   taskEnqueuer
}

// NewService injeta dependências do domínio user.
func NewService(repo Repository, v *validator.Validator, client taskEnqueuer) *Service {
	return &Service{
		repo:      repo,
		validator: v,
		enqueue:   client,
	}
}

// Create valida, persiste e enfileira email de boas-vindas.
func (s *Service) Create(ctx context.Context, in CreateDTO) (*ResponseDTO, error) {
	if err := s.validator.Struct(in); err != nil {
		return nil, fmt.Errorf("validate create dto: %w", err)
	}

	existing, err := s.repo.FindByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("check email exists: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &User{
		ID:           uuid.New(),
		Email:        in.Email,
		PasswordHash: string(hash),
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("persist user: %w", err)
	}

	payload, err := json.Marshal(queue.WelcomeEmailPayload{Email: u.Email})
	if err != nil {
		return nil, fmt.Errorf("marshal welcome payload: %w", err)
	}
	task := asynq.NewTask(queue.TaskTypeWelcomeEmail, payload)
	if _, err := s.enqueue.Enqueue(task); err != nil {
		return nil, fmt.Errorf("enqueue welcome email: %w", err)
	}

	out := toResponseDTO(u)
	return &out, nil
}

// GetByID retorna usuário por ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*ResponseDTO, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := toResponseDTO(u)
	return &out, nil
}

func toResponseDTO(u *User) ResponseDTO {
	return ResponseDTO{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
