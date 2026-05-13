package orderapp

import (
	"context"
	"fmt"

	orderdomain "github.com/caiohenrique/go-api-template/internal/features/order/domain"
	"github.com/caiohenrique/go-api-template/internal/platform/id"
	"github.com/caiohenrique/go-api-template/internal/platform/pagination"
	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/google/uuid"
)

// Service contains order use cases.
type Service struct {
	repo      OrderRepository
	validator *validator.Validator
}

// NewService wires order application services.
func NewService(repo OrderRepository, v *validator.Validator) *Service {
	return &Service{repo: repo, validator: v}
}

// Create validates and persists an order with services.
func (s *Service) Create(ctx context.Context, in CreateOrderInput) (*OrderView, error) {
	if err := s.validator.Struct(in); err != nil {
		return nil, fmt.Errorf("validate create order input: %w", err)
	}

	orderID, err := id.New()
	if err != nil {
		return nil, fmt.Errorf("generate order id: %w", err)
	}
	o := &orderdomain.Order{
		ID:          orderID,
		Title:       in.Title,
		Subject:     in.Subject,
		Code:        in.Code,
		SentAt:      in.SentAt,
		ConvertedAt: in.ConvertedAt,
	}
	services := make([]orderdomain.OrderService, 0, len(in.OrderServices))
	for _, row := range in.OrderServices {
		svcID, err := id.New()
		if err != nil {
			return nil, fmt.Errorf("generate order service id: %w", err)
		}
		services = append(services, orderdomain.OrderService{
			ID:           svcID,
			OrderID:      orderID,
			Title:        row.Title,
			StartDate:    row.StartDate,
			EndDate:      row.EndDate,
			Observations: row.Observations,
		})
	}

	if err := s.repo.Create(ctx, o, services); err != nil {
		return nil, fmt.Errorf("persist order: %w", err)
	}

	stored, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("load order after create: %w", err)
	}
	out := toOrderView(stored)
	return &out, nil
}

// GetByID returns an order by ID with services.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*OrderView, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := toOrderView(o)
	return &out, nil
}

// List returns a paginated page of orders (summary without nested services in list).
func (s *Service) List(ctx context.Context, in ListOrdersInput) (*OrdersPage, error) {
	in.Query = pagination.Normalize(in.Query)
	if err := s.validator.Struct(in); err != nil {
		return nil, fmt.Errorf("validate list orders input: %w", err)
	}

	params := ListOrdersRepoInput(in)
	orders, total, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	items := make([]OrderView, 0, len(orders))
	for i := range orders {
		items = append(items, toOrderView(&orders[i]))
	}

	meta := pagination.NewMeta(in.Query, total)
	return &OrdersPage{
		Items: items,
		Meta:  meta,
	}, nil
}

// ListOrderServices returns order_service rows with pagination.
func (s *Service) ListOrderServices(ctx context.Context, in ListOrderServicesInput) (*OrderServicesPage, error) {
	in.Query = pagination.Normalize(in.Query)
	if err := s.validator.Struct(in); err != nil {
		return nil, fmt.Errorf("validate list order services input: %w", err)
	}

	var orderID *uuid.UUID
	if in.OrderID != "" {
		id, err := uuid.Parse(in.OrderID)
		if err != nil {
			return nil, fmt.Errorf("parse order_id: %w", err)
		}
		orderID = &id
	}

	rows, total, err := s.repo.ListOrderServices(ctx, ListOrderServicesRepoInput{
		Query:   in.Query,
		OrderID: orderID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]OrderServiceView, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		items = append(items, OrderServiceView{
			ID:           row.ID,
			OrderID:      row.OrderID,
			Title:        row.Title,
			StartDate:    row.StartDate,
			EndDate:      row.EndDate,
			Observations: row.Observations,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}
	meta := pagination.NewMeta(in.Query, total)
	return &OrderServicesPage{
		Items: items,
		Meta:  meta,
	}, nil
}

// Update applies full header replace and service upserts/removals in one repository transaction.
func (s *Service) Update(ctx context.Context, orderID uuid.UUID, in UpdateOrderInput) (*OrderView, error) {
	if err := s.validator.Struct(in); err != nil {
		return nil, fmt.Errorf("validate update order input: %w", err)
	}

	removeSet := make(map[uuid.UUID]struct{}, len(in.RemovedServiceIDs))
	for _, rid := range in.RemovedServiceIDs {
		removeSet[rid] = struct{}{}
	}

	creates := make([]orderdomain.OrderService, 0, len(in.OrderServices))
	updates := make([]orderdomain.OrderService, 0, len(in.OrderServices))
	for _, row := range in.OrderServices {
		if row.ID == nil {
			svcID, err := id.New()
			if err != nil {
				return nil, fmt.Errorf("generate order service id: %w", err)
			}
			creates = append(creates, orderdomain.OrderService{
				ID:           svcID,
				OrderID:      orderID,
				Title:        row.Title,
				StartDate:    row.StartDate,
				EndDate:      row.EndDate,
				Observations: row.Observations,
			})
			continue
		}
		if _, conflict := removeSet[*row.ID]; conflict {
			return nil, fmt.Errorf("service %s is both updated and removed: %w", *row.ID, orderdomain.ErrServiceNotInOrder)
		}
		updates = append(updates, orderdomain.OrderService{
			ID:           *row.ID,
			OrderID:      orderID,
			Title:        row.Title,
			StartDate:    row.StartDate,
			EndDate:      row.EndDate,
			Observations: row.Observations,
		})
	}

	repoIn := UpdateOrderRepoInput{
		Title:       in.Title,
		Subject:     in.Subject,
		Code:        in.Code,
		SentAt:      in.SentAt,
		ConvertedAt: in.ConvertedAt,
		Creates:     creates,
		Updates:     updates,
		RemoveIDs:   in.RemovedServiceIDs,
	}

	if err := s.repo.Update(ctx, orderID, repoIn); err != nil {
		return nil, fmt.Errorf("persist order update: %w", err)
	}

	stored, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("load order after update: %w", err)
	}
	out := toOrderView(stored)
	return &out, nil
}

// Delete soft-deletes an order.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func toOrderView(o *orderdomain.Order) OrderView {
	svc := make([]OrderServiceView, 0, len(o.OrderServices))
	for i := range o.OrderServices {
		row := &o.OrderServices[i]
		svc = append(svc, OrderServiceView{
			ID:           row.ID,
			OrderID:      row.OrderID,
			Title:        row.Title,
			StartDate:    row.StartDate,
			EndDate:      row.EndDate,
			Observations: row.Observations,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}
	return OrderView{
		ID:            o.ID,
		Title:         o.Title,
		Subject:       o.Subject,
		Code:          o.Code,
		SentAt:        o.SentAt,
		ConvertedAt:   o.ConvertedAt,
		OrderServices: svc,
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
	}
}
