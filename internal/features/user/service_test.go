package user

import (
	"context"
	"testing"

	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type fakeRepo struct {
	byEmail map[string]*User
	byID    map[uuid.UUID]*User
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		byEmail: make(map[string]*User),
		byID:    make(map[uuid.UUID]*User),
	}
}

func (f *fakeRepo) Create(ctx context.Context, u *User) error {
	_ = ctx
	cp := *u
	f.byEmail[u.Email] = &cp
	f.byID[u.ID] = &cp
	return nil
}

func (f *fakeRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	_ = ctx
	u, ok := f.byEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (f *fakeRepo) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	_ = ctx
	u, ok := f.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

type noopEnqueue struct{}

func (noopEnqueue) Enqueue(t *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	_, _ = t, len(opts)
	return nil, nil
}

func TestService_Create_Success(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New(), noopEnqueue{})

	out, err := svc.Create(ctx, CreateDTO{
		Email:    "a@example.com",
		Password: "password1",
	})
	require.NoError(t, err)
	require.Equal(t, "a@example.com", out.Email)
	require.NotEqual(t, uuid.Nil, out.ID)

	stored, err := repo.FindByEmail(ctx, "a@example.com")
	require.NoError(t, err)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("password1")))
}

func TestService_Create_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New(), noopEnqueue{})

	_, err := svc.Create(ctx, CreateDTO{Email: "b@example.com", Password: "password1"})
	require.NoError(t, err)

	_, err = svc.Create(ctx, CreateDTO{Email: "b@example.com", Password: "password2"})
	require.ErrorIs(t, err, ErrEmailAlreadyExists)
}
