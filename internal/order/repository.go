package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UpdateOrderInput agrega cabeçalho (replace total) e mudanças de linhas em uma única transação.
type UpdateOrderInput struct {
	Title       string
	Subject     *string
	Code        *string
	SentAt      *time.Time
	ConvertedAt *time.Time
	Creates     []OrderService
	Updates     []OrderService
	RemoveIDs   []uuid.UUID
}

// Repository define persistência de pedidos (consumido pelo Service).
type Repository interface {
	Create(ctx context.Context, o *Order, services []OrderService) error
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	Update(ctx context.Context, orderID uuid.UUID, in UpdateOrderInput) error
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

func (r *gormRepository) Update(ctx context.Context, orderID uuid.UUID, in UpdateOrderInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		orderUpdates := map[string]any{
			"title":         in.Title,
			"subject":       in.Subject,
			"code":          in.Code,
			"sent_at":       in.SentAt,
			"converted_at":  in.ConvertedAt,
			"updated_at":    gorm.Expr("NOW()"),
		}
		res := tx.Model(&Order{}).Where("id = ?", orderID).Updates(orderUpdates)
		if res.Error != nil {
			return fmt.Errorf("update order: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}

		for _, svc := range in.Updates {
			res := tx.Model(&OrderService{}).
				Where("id = ? AND order_id = ?", svc.ID, orderID).
				Updates(map[string]any{
					"title":          svc.Title,
					"start_date":     svc.StartDate,
					"end_date":       svc.EndDate,
					"observations":   svc.Observations,
					"updated_at":     gorm.Expr("NOW()"),
				})
			if res.Error != nil {
				return fmt.Errorf("update order service %s: %w", svc.ID, res.Error)
			}
			if res.RowsAffected == 0 {
				return ErrServiceNotInOrder
			}
		}

		if len(in.Creates) > 0 {
			if err := tx.Create(&in.Creates).Error; err != nil {
				return fmt.Errorf("create order services: %w", err)
			}
		}

		if len(in.RemoveIDs) > 0 {
			res := tx.Where("order_id = ? AND id IN ?", orderID, in.RemoveIDs).Delete(&OrderService{})
			if res.Error != nil {
				return fmt.Errorf("delete order services: %w", res.Error)
			}
			if res.RowsAffected != int64(len(in.RemoveIDs)) {
				return ErrServiceNotInOrder
			}
		}

		return nil
	})
}
