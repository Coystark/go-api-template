package userrepo

import (
	"context"
	"errors"
	"fmt"

	userapp "github.com/caiohenrique/go-api-template/internal/features/user/app"
	userdomain "github.com/caiohenrique/go-api-template/internal/features/user/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

// NewRepository creates the GORM-backed user repository.
func NewRepository(db *gorm.DB) userapp.UserRepository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, u *userdomain.User) error {
	m := fromDomain(u)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	u.CreatedAt = m.CreatedAt
	u.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *gormRepository) FindByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, userdomain.ErrNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return toDomain(&m), nil
}

func (r *gormRepository) FindByID(ctx context.Context, id uuid.UUID) (*userdomain.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, userdomain.ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return toDomain(&m), nil
}
