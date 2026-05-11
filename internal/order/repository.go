package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository define persistência de pedidos (consumido pelo Service).
type Repository interface {
	Create(ctx context.Context, o *Order, services []OrderService) error
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	Update(ctx context.Context, orderID uuid.UUID, title string, services []OrderService) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository cria implementação GORM do repositório.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, o *Order, services []OrderService) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(o).Error; err != nil {
			return fmt.Errorf("create order: %w", err)
		}
		if len(services) == 0 {
			return nil
		}
		if err := tx.Create(&services).Error; err != nil {
			return fmt.Errorf("create order services: %w", err)
		}
		return nil
	})
}

func (r *gormRepository) FindByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	var o Order
	if err := r.db.WithContext(ctx).Preload("OrderServices").Where("id = ?", id).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find order by id: %w", err)
	}
	return &o, nil
}

func (r *gormRepository) Update(ctx context.Context, orderID uuid.UUID, title string, services []OrderService) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Order{}).Where("id = ?", orderID).Updates(map[string]any{
			"title":      title,
			"updated_at": gorm.Expr("NOW()"),
		})
		if res.Error != nil {
			return fmt.Errorf("update order: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		if err := tx.Where("order_id = ?", orderID).Delete(&OrderService{}).Error; err != nil {
			return fmt.Errorf("delete order services: %w", err)
		}
		if len(services) == 0 {
			return nil
		}
		if err := tx.Create(&services).Error; err != nil {
			return fmt.Errorf("create order services: %w", err)
		}
		return nil
	})
}
