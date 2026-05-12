package order

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Order é o modelo persistido no banco.
type Order struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Title         string    `gorm:"type:text;not null"`
	Subject       *string   `gorm:"type:text"`
	Code          *string   `gorm:"type:text"`
	SentAt        *time.Time
	ConvertedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	OrderServices []OrderService `gorm:"foreignKey:OrderID"`
}

// TableName define o nome da tabela para o GORM.
func (Order) TableName() string {
	return "orders"
}

// OrderService é um serviço vinculado a um pedido.
type OrderService struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OrderID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Title        string     `gorm:"type:text;not null"`
	StartDate    time.Time  `gorm:"column:start_date;not null"`
	EndDate      *time.Time `gorm:"column:end_date"`
	Observations *string    `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName define o nome da tabela para o GORM.
func (OrderService) TableName() string {
	return "order_services"
}
