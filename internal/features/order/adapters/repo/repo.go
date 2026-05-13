package orderrepo

import (
	"context"
	"errors"
	"fmt"

	orderapp "github.com/caiohenrique/go-api-template/internal/features/order/app"
	orderdomain "github.com/caiohenrique/go-api-template/internal/features/order/domain"
	"github.com/caiohenrique/go-api-template/internal/platform/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

// NewRepository creates the GORM-backed order repository.
func NewRepository(db *gorm.DB) orderapp.OrderRepository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, o *orderdomain.Order, services []orderdomain.OrderService) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		om := fromDomainOrder(o)
		// Persist header only first; nested services are passed separately for create flow
		om.OrderServices = nil
		if err := tx.Create(om).Error; err != nil {
			return fmt.Errorf("create order: %w", err)
		}
		o.CreatedAt = om.CreatedAt
		o.UpdatedAt = om.UpdatedAt
		if len(services) == 0 {
			return nil
		}
		models := make([]orderServiceModel, 0, len(services))
		for i := range services {
			models = append(models, *fromDomainOrderService(&services[i]))
		}
		if err := tx.Create(&models).Error; err != nil {
			return fmt.Errorf("create order services: %w", err)
		}
		for i := range services {
			services[i].CreatedAt = models[i].CreatedAt
			services[i].UpdatedAt = models[i].UpdatedAt
		}
		return nil
	})
}

func (r *gormRepository) FindByID(ctx context.Context, id uuid.UUID) (*orderdomain.Order, error) {
	var m orderModel
	if err := r.db.WithContext(ctx).Preload("OrderServices").Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, orderdomain.ErrNotFound
		}
		return nil, fmt.Errorf("find order by id: %w", err)
	}
	return toDomainOrder(&m), nil
}

func (r *gormRepository) List(ctx context.Context, in orderapp.ListOrdersRepoInput) ([]orderdomain.Order, int64, error) {
	db := r.db.WithContext(ctx).Model(&orderModel{})
	if in.Code != "" {
		db = db.Where("code ILIKE ?", "%"+in.Code+"%")
	}
	if in.Title != "" {
		db = db.Where("title ILIKE ?", "%"+in.Title+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	var rows []orderModel
	if err := db.Order("created_at DESC").
		Limit(in.PageSize).
		Offset(pagination.Offset(in.Query)).
		Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	out := make([]orderdomain.Order, 0, len(rows))
	for i := range rows {
		out = append(out, *toDomainOrder(&rows[i]))
	}
	return out, total, nil
}

func (r *gormRepository) ListOrderServices(ctx context.Context, in orderapp.ListOrderServicesRepoInput) ([]orderdomain.OrderService, int64, error) {
	db := r.db.WithContext(ctx).Model(&orderServiceModel{}).
		Joins("JOIN orders ON orders.id = order_services.order_id AND orders.deleted_at IS NULL")
	if in.OrderID != nil {
		db = db.Where("order_services.order_id = ?", *in.OrderID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count order services: %w", err)
	}

	var models []orderServiceModel
	if err := db.Order("order_services.created_at DESC").
		Limit(in.PageSize).
		Offset(pagination.Offset(in.Query)).
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("list order services: %w", err)
	}
	out := make([]orderdomain.OrderService, 0, len(models))
	for i := range models {
		out = append(out, *toDomainOrderService(&models[i]))
	}
	return out, total, nil
}

func (r *gormRepository) Update(ctx context.Context, orderID uuid.UUID, in orderapp.UpdateOrderRepoInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		orderUpdates := map[string]any{
			"title":        in.Title,
			"subject":      in.Subject,
			"code":         in.Code,
			"sent_at":      in.SentAt,
			"converted_at": in.ConvertedAt,
			"updated_at":   gorm.Expr("NOW()"),
		}
		res := tx.Model(&orderModel{}).Where("id = ?", orderID).Updates(orderUpdates)
		if res.Error != nil {
			return fmt.Errorf("update order: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return orderdomain.ErrNotFound
		}

		for _, svc := range in.Updates {
			res := tx.Model(&orderServiceModel{}).
				Where("id = ? AND order_id = ?", svc.ID, orderID).
				Updates(map[string]any{
					"title":        svc.Title,
					"start_date":   svc.StartDate,
					"end_date":     svc.EndDate,
					"observations": svc.Observations,
					"updated_at":   gorm.Expr("NOW()"),
				})
			if res.Error != nil {
				return fmt.Errorf("update order service %s: %w", svc.ID, res.Error)
			}
			if res.RowsAffected == 0 {
				return orderdomain.ErrServiceNotInOrder
			}
		}

		if len(in.Creates) > 0 {
			models := make([]orderServiceModel, 0, len(in.Creates))
			for i := range in.Creates {
				models = append(models, *fromDomainOrderService(&in.Creates[i]))
			}
			if err := tx.Create(&models).Error; err != nil {
				return fmt.Errorf("create order services: %w", err)
			}
		}

		if len(in.RemoveIDs) > 0 {
			res := tx.Where("order_id = ? AND id IN ?", orderID, in.RemoveIDs).Delete(&orderServiceModel{})
			if res.Error != nil {
				return fmt.Errorf("delete order services: %w", res.Error)
			}
			if res.RowsAffected != int64(len(in.RemoveIDs)) {
				return orderdomain.ErrServiceNotInOrder
			}
		}

		return nil
	})
}

func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&orderModel{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("delete order: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return orderdomain.ErrNotFound
	}
	return nil
}
