package order

import (
	"context"
	"testing"

	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	orders map[uuid.UUID]*Order
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{orders: make(map[uuid.UUID]*Order)}
}

func (f *fakeRepo) Create(ctx context.Context, o *Order, services []OrderService) error {
	_ = ctx
	cp := *o
	cp.OrderServices = append([]OrderService(nil), services...)
	f.orders[o.ID] = &cp
	return nil
}

func (f *fakeRepo) FindByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	_ = ctx
	o, ok := f.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *o
	cp.OrderServices = append([]OrderService(nil), o.OrderServices...)
	return &cp, nil
}

func (f *fakeRepo) Update(ctx context.Context, orderID uuid.UUID, title string, services []OrderService) error {
	_ = ctx
	o, ok := f.orders[orderID]
	if !ok {
		return ErrNotFound
	}
	o.Title = title
	o.OrderServices = append([]OrderService(nil), services...)
	return nil
}

func TestService_Create_WithServices(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	out, err := svc.Create(ctx, CreateOrderDTO{
		Title: "Pedido A",
		OrderServices: []OrderServiceInput{
			{Title: "Serviço 1"},
			{Title: "Serviço 2"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Pedido A", out.Title)
	require.Len(t, out.OrderServices, 2)
	require.Equal(t, "Serviço 1", out.OrderServices[0].Title)
	require.Equal(t, "Serviço 2", out.OrderServices[1].Title)
	require.NotEqual(t, uuid.Nil, out.ID)
}

func TestService_Create_EmptyServices(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	out, err := svc.Create(ctx, CreateOrderDTO{
		Title:         "Sem serviços",
		OrderServices: nil,
	})
	require.NoError(t, err)
	require.Empty(t, out.OrderServices)
}

func TestService_Update_ReplaceServices(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	created, err := svc.Create(ctx, CreateOrderDTO{
		Title: "Original",
		OrderServices: []OrderServiceInput{
			{Title: "Antigo"},
		},
	})
	require.NoError(t, err)
	oldSvcID := created.OrderServices[0].ID

	out, err := svc.Update(ctx, created.ID, UpdateOrderDTO{
		Title: "Atualizado",
		OrderServices: []OrderServiceInput{
			{Title: "Novo 1"},
			{Title: "Novo 2"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Atualizado", out.Title)
	require.Len(t, out.OrderServices, 2)
	require.NotEqual(t, oldSvcID, out.OrderServices[0].ID)
	require.Equal(t, "Novo 1", out.OrderServices[0].Title)
	require.Equal(t, "Novo 2", out.OrderServices[1].Title)
}

func TestService_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	_, err := svc.Update(ctx, uuid.New(), UpdateOrderDTO{
		Title: "X",
		OrderServices: []OrderServiceInput{
			{Title: "Y"},
		},
	})
	require.ErrorIs(t, err, ErrNotFound)
}
