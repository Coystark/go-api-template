package orderrepo

import (
	"time"

	orderdomain "github.com/caiohenrique/go-api-template/internal/features/order/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type orderModel struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Title         string    `gorm:"type:text;not null"`
	Subject       *string   `gorm:"type:text"`
	Code          *string   `gorm:"type:text"`
	SentAt        *time.Time
	ConvertedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt      `gorm:"index"`
	OrderServices []orderServiceModel `gorm:"foreignKey:OrderID"`
}

func (orderModel) TableName() string { return "orders" }

type orderServiceModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OrderID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Title        string     `gorm:"type:text;not null"`
	StartDate    time.Time  `gorm:"column:start_date;not null"`
	EndDate      *time.Time `gorm:"column:end_date"`
	Observations *string    `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (orderServiceModel) TableName() string { return "order_services" }

func toDomainOrder(m *orderModel) *orderdomain.Order {
	if m == nil {
		return nil
	}
	out := &orderdomain.Order{
		ID:          m.ID,
		Title:       m.Title,
		Subject:     m.Subject,
		Code:        m.Code,
		SentAt:      m.SentAt,
		ConvertedAt: m.ConvertedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	for i := range m.OrderServices {
		out.OrderServices = append(out.OrderServices, *toDomainOrderService(&m.OrderServices[i]))
	}
	return out
}

func fromDomainOrder(o *orderdomain.Order) *orderModel {
	if o == nil {
		return nil
	}
	m := &orderModel{
		ID:          o.ID,
		Title:       o.Title,
		Subject:     o.Subject,
		Code:        o.Code,
		SentAt:      o.SentAt,
		ConvertedAt: o.ConvertedAt,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
	for i := range o.OrderServices {
		m.OrderServices = append(m.OrderServices, *fromDomainOrderService(&o.OrderServices[i]))
	}
	return m
}

func toDomainOrderService(m *orderServiceModel) *orderdomain.OrderService {
	if m == nil {
		return nil
	}
	return &orderdomain.OrderService{
		ID:           m.ID,
		OrderID:      m.OrderID,
		Title:        m.Title,
		StartDate:    m.StartDate,
		EndDate:      m.EndDate,
		Observations: m.Observations,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func fromDomainOrderService(s *orderdomain.OrderService) *orderServiceModel {
	if s == nil {
		return nil
	}
	return &orderServiceModel{
		ID:           s.ID,
		OrderID:      s.OrderID,
		Title:        s.Title,
		StartDate:    s.StartDate,
		EndDate:      s.EndDate,
		Observations: s.Observations,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}
