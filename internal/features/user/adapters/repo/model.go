package userrepo

import (
	"time"

	userdomain "github.com/caiohenrique/go-api-template/internal/features/user/domain"
	"github.com/google/uuid"
)

type userModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (userModel) TableName() string { return "users" }

func toDomain(m *userModel) *userdomain.User {
	if m == nil {
		return nil
	}
	return &userdomain.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func fromDomain(u *userdomain.User) *userModel {
	if u == nil {
		return nil
	}
	return &userModel{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
