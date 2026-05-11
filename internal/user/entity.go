package user

import (
	"time"

	"github.com/google/uuid"
)

// User é o modelo persistido no banco.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName define o nome da tabela para o GORM.
func (User) TableName() string {
	return "users"
}
