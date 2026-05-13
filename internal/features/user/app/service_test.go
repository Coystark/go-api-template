package userapp

import (
	"context"
	"testing"

	userdomain "github.com/caiohenrique/go-api-template/internal/features/user/domain"
	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	byEmail map[string]*userdomain.User
	byID    map[uuid.UUID]*userdomain.User
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		byEmail: make(map[string]*userdomain.User),
		byID:    make(map[uuid.UUID]*userdomain.User),
	}
}

func (f *fakeRepo) Create(ctx context.Context, u *userdomain.User) error {
	_ = ctx
	cp := *u
	f.byEmail[u.Email] = &cp
	f.byID[u.ID] = &cp
	return nil
}

func (f *fakeRepo) FindByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	_ = ctx
	u, ok := f.byEmail[email]
	if !ok {
		return nil, userdomain.ErrNotFound
	}
	return u, nil
}

func (f *fakeRepo) FindByID(ctx context.Context, id uuid.UUID) (*userdomain.User, error) {
	_ = ctx
	u, ok := f.byID[id]
	if !ok {
		return nil, userdomain.ErrNotFound
	}
	return u, nil
}

type fakeMailer struct {
	lastEmails []string
}

func (f *fakeMailer) SendWelcome(ctx context.Context, email string) error {
	_ = ctx
	f.lastEmails = append(f.lastEmails, email)
	return nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(plain string) (string, error) {
	return "hash:" + plain, nil
}

func (fakeHasher) Compare(hash, plain string) error {
	_ = hash
	_ = plain
	return nil
}

func TestService_Create_Success(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	mailer := &fakeMailer{}
	svc := NewService(repo, validator.New(), mailer, fakeHasher{})

	out, err := svc.Create(ctx, CreateUserInput{
		Email:    "a@example.com",
		Password: "password1",
	})
	require.NoError(t, err)
	require.Equal(t, "a@example.com", out.Email)
	require.NotEqual(t, uuid.Nil, out.ID)

	stored, err := repo.FindByEmail(ctx, "a@example.com")
	require.NoError(t, err)
	require.Equal(t, "hash:password1", stored.PasswordHash)
	require.Len(t, mailer.lastEmails, 1)
	require.Equal(t, "a@example.com", mailer.lastEmails[0])
}

func TestService_Create_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New(), &fakeMailer{}, fakeHasher{})

	_, err := svc.Create(ctx, CreateUserInput{Email: "b@example.com", Password: "password1"})
	require.NoError(t, err)

	_, err = svc.Create(ctx, CreateUserInput{Email: "b@example.com", Password: "password2"})
	require.ErrorIs(t, err, userdomain.ErrEmailAlreadyExists)
}
