package order

import (
	"context"
	"fmt"

	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/google/uuid"
)

// Service contém regras de negócio do domínio de pedidos.
type Service struct {
	repo      Repository
	validator *validator.Validator
}

// NewService injeta dependências do domínio order.
func NewService(repo Repository, v *validator.Validator) *Service {
	return &Service{repo: repo, validator: v}
}

// Create valida e persiste pedido com serviços.
func (s *Service) Create(ctx context.Context, in CreateOrderDTO) (*ResponseDTO, error) {
	if err := s.validator.Struct(in); err != nil {
		return nil, fmt.Errorf("validate create dto: %w", err)
	}

	orderID := uuid.New()
	o := &Order{
		ID:          orderID,
		Title:       in.Title,
		Subject:     in.Subject,
		Code:        in.Code,
		SentAt:      in.SentAt,
		ConvertedAt: in.ConvertedAt,
	}
	services := make([]OrderService, 0, len(in.OrderServices))
	for _, row := range in.OrderServices {
		services = append(services, OrderService{
			ID:      uuid.New(),
			OrderID: orderID,
			Title:   row.Title,
		})
	}

	if err := s.repo.Create(ctx, o, services); err != nil {
		return nil, fmt.Errorf("persist order: %w", err)
	}

	stored, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("load order after create: %w", err)
	}
	out := toResponseDTO(stored)
	return &out, nil
}

// GetByID retorna pedido por ID com serviços.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*ResponseDTO, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := toResponseDTO(o)
	return &out, nil
}

// Update aplica uma atualização parcial: itens com ID viram UPDATE, sem ID viram INSERT,
// e RemovedServiceIDs são deletados na mesma transação. Omissões não apagam nada.
func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpdateOrderDTO) (*ResponseDTO, error) {
	if err := s.validator.Struct(in); err != nil {
		return nil, fmt.Errorf("validate update dto: %w", err)
	}

	removeSet := make(map[uuid.UUID]struct{}, len(in.RemovedServiceIDs))
	for _, rid := range in.RemovedServiceIDs {
		removeSet[rid] = struct{}{}
	}

	creates := make([]OrderService, 0, len(in.OrderServices))
	updates := make([]OrderService, 0, len(in.OrderServices))
	for _, row := range in.OrderServices {
		if row.ID == nil {
			creates = append(creates, OrderService{
				ID:      uuid.New(),
				OrderID: id,
				Title:   row.Title,
			})
			continue
		}
		if _, conflict := removeSet[*row.ID]; conflict {
			return nil, fmt.Errorf("service %s is both updated and removed: %w", *row.ID, ErrServiceNotInOrder)
		}
		updates = append(updates, OrderService{
			ID:      *row.ID,
			OrderID: id,
			Title:   row.Title,
		})
	}

	repoIn := UpdateOrderInput{
		Title:       in.Title,
		Subject:     in.Subject,
		Code:        in.Code,
		SentAt:      in.SentAt,
		ConvertedAt: in.ConvertedAt,
		Creates:     creates,
		Updates:     updates,
		RemoveIDs:   in.RemovedServiceIDs,
	}

	if err := s.repo.Update(ctx, id, repoIn); err != nil {
		return nil, fmt.Errorf("persist order update: %w", err)
	}

	stored, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load order after update: %w", err)
	}
	out := toResponseDTO(stored)
	return &out, nil
}

func toResponseDTO(o *Order) ResponseDTO {
	svc := make([]OrderServiceResponse, 0, len(o.OrderServices))
	for i := range o.OrderServices {
		row := &o.OrderServices[i]
		svc = append(svc, OrderServiceResponse{
			ID:        row.ID,
			Title:     row.Title,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return ResponseDTO{
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
