package authapp

import (
	"context"
	"errors"
	"testing"
	"time"

	authdomain "github.com/caiohenrique/go-api-template/internal/features/auth/domain"
	userdomain "github.com/caiohenrique/go-api-template/internal/features/user/domain"
	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeUserFinder struct {
	byEmail map[string]*userdomain.User
}

func newFakeUserFinder(u *userdomain.User) *fakeUserFinder {
	m := make(map[string]*userdomain.User)
	if u != nil {
		m[u.Email] = u
	}
	return &fakeUserFinder{byEmail: m}
}

func (f *fakeUserFinder) FindByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	_ = ctx
	u, ok := f.byEmail[email]
	if !ok {
		return nil, userdomain.ErrNotFound
	}
	return u, nil
}

type fakeIssuer struct{}

func (fakeIssuer) Generate(userID uuid.UUID) (string, error) {
	return "token:" + userID.String(), nil
}

type fakeAuthHasher struct{}

func (fakeAuthHasher) Compare(hash, plain string) error {
	if hash != "hash:"+plain {
		return errors.New("mismatch")
	}
	return nil
}

func TestService_Login_Success(t *testing.T) {
	ctx := context.Background()
	uid := uuid.Must(uuid.NewV7())
	u := &userdomain.User{
		ID:           uid,
		Email:        "x@example.com",
		PasswordHash: "hash:secret",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	svc := NewService(newFakeUserFinder(u), fakeIssuer{}, fakeAuthHasher{}, validator.New())

	out, err := svc.Login(ctx, LoginInput{Email: "x@example.com", Password: "secret"})
	require.NoError(t, err)
	require.Equal(t, "token:"+uid.String(), out.AccessToken)
}

func TestService_Login_WrongPassword(t *testing.T) {
	ctx := context.Background()
	uid := uuid.Must(uuid.NewV7())
	u := &userdomain.User{
		ID:           uid,
		Email:        "x@example.com",
		PasswordHash: "hash:secret",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	svc := NewService(newFakeUserFinder(u), fakeIssuer{}, fakeAuthHasher{}, validator.New())

	_, err := svc.Login(ctx, LoginInput{Email: "x@example.com", Password: "wrong"})
	require.ErrorIs(t, err, authdomain.ErrInvalidCredentials)
}

func TestService_Login_UnknownEmail(t *testing.T) {
	ctx := context.Background()
	svc := NewService(newFakeUserFinder(nil), fakeIssuer{}, fakeAuthHasher{}, validator.New())

	_, err := svc.Login(ctx, LoginInput{Email: "nope@example.com", Password: "secret"})
	require.ErrorIs(t, err, authdomain.ErrInvalidCredentials)
}
