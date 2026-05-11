package order

import (
	"context"
	"testing"
	"time"

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

func (f *fakeRepo) Update(ctx context.Context, orderID uuid.UUID, in UpdateOrderInput) error {
	_ = ctx
	o, ok := f.orders[orderID]
	if !ok {
		return ErrNotFound
	}

	if in.Title != nil {
		o.Title = *in.Title
	}
	if in.Subject != nil {
		o.Subject = in.Subject
	}
	if in.Code != nil {
		o.Code = in.Code
	}
	if in.SentAt != nil {
		o.SentAt = in.SentAt
	}
	if in.ConvertedAt != nil {
		o.ConvertedAt = in.ConvertedAt
	}

	for _, svc := range in.Updates {
		idx := indexOfService(o.OrderServices, svc.ID)
		if idx < 0 {
			return ErrServiceNotInOrder
		}
		o.OrderServices[idx].Title = svc.Title
	}

	if len(in.Creates) > 0 {
		o.OrderServices = append(o.OrderServices, in.Creates...)
	}

	for _, rid := range in.RemoveIDs {
		idx := indexOfService(o.OrderServices, rid)
		if idx < 0 {
			return ErrServiceNotInOrder
		}
		o.OrderServices = append(o.OrderServices[:idx], o.OrderServices[idx+1:]...)
	}

	return nil
}

func indexOfService(services []OrderService, id uuid.UUID) int {
	for i := range services {
		if services[i].ID == id {
			return i
		}
	}
	return -1
}

func ptrStr(s string) *string { return &s }
func ptrTime(t time.Time) *time.Time { return &t }

func TestService_Create_WithServices(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	sentAt := time.Date(2026, 5, 11, 10, 30, 0, 0, time.UTC)
	out, err := svc.Create(ctx, CreateOrderDTO{
		Title:   "Pedido A",
		Subject: ptrStr("Assunto A"),
		Code:    ptrStr("COD-123"),
		SentAt:  ptrTime(sentAt),
		OrderServices: []OrderServiceInput{
			{Title: "Serviço 1"},
			{Title: "Serviço 2"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Pedido A", out.Title)
	require.Equal(t, "Assunto A", *out.Subject)
	require.Equal(t, "COD-123", *out.Code)
	require.NotNil(t, out.SentAt)
	require.True(t, out.SentAt.Equal(sentAt))
	require.Nil(t, out.ConvertedAt)
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

func TestService_Update_PartialUpsert(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	created, err := svc.Create(ctx, CreateOrderDTO{
		Title: "Original",
		OrderServices: []OrderServiceInput{
			{Title: "Antigo 1"},
			{Title: "Antigo 2"},
			{Title: "Antigo 3"},
		},
	})
	require.NoError(t, err)
	require.Len(t, created.OrderServices, 3)
	keepID := created.OrderServices[0].ID
	editID := created.OrderServices[1].ID
	untouchedID := created.OrderServices[2].ID

	convertedAt := time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)
	out, err := svc.Update(ctx, created.ID, UpdateOrderDTO{
		Title: ptrStr("Atualizado"),
		ConvertedAt: ptrTime(convertedAt),
		OrderServices: []OrderServiceInput{
			{ID: &editID, Title: "Editado"},
			{Title: "Novo"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Atualizado", out.Title)
	require.NotNil(t, out.ConvertedAt)
	require.True(t, out.ConvertedAt.Equal(convertedAt))
	require.Len(t, out.OrderServices, 4)

	got := indexByID(out.OrderServices)
	require.Equal(t, "Antigo 1", got[keepID].Title)
	require.Equal(t, "Editado", got[editID].Title)
	require.Equal(t, "Antigo 3", got[untouchedID].Title)

	var newCount int
	for id := range got {
		if id != keepID && id != editID && id != untouchedID {
			newCount++
			require.Equal(t, "Novo", got[id].Title)
		}
	}
	require.Equal(t, 1, newCount)
}

func TestService_Update_TitleOnly(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	created, err := svc.Create(ctx, CreateOrderDTO{
		Title: "Original",
		OrderServices: []OrderServiceInput{
			{Title: "S1"},
		},
	})
	require.NoError(t, err)

	out, err := svc.Update(ctx, created.ID, UpdateOrderDTO{
		Title: ptrStr("Só título"),
	})
	require.NoError(t, err)
	require.Equal(t, "Só título", out.Title)
	require.Len(t, out.OrderServices, 1)
	require.Equal(t, "S1", out.OrderServices[0].Title)
}

func TestService_Update_RemovedServiceIDs(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	created, err := svc.Create(ctx, CreateOrderDTO{
		Title: "Pedido",
		OrderServices: []OrderServiceInput{
			{Title: "Manter"},
			{Title: "Remover"},
		},
	})
	require.NoError(t, err)
	keepID := created.OrderServices[0].ID
	removeID := created.OrderServices[1].ID

	out, err := svc.Update(ctx, created.ID, UpdateOrderDTO{
		RemovedServiceIDs: []uuid.UUID{removeID},
	})
	require.NoError(t, err)
	require.Len(t, out.OrderServices, 1)
	require.Equal(t, keepID, out.OrderServices[0].ID)
}

func TestService_Update_RemoveAndUpdateSameID_Fails(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	created, err := svc.Create(ctx, CreateOrderDTO{
		Title: "Pedido",
		OrderServices: []OrderServiceInput{
			{Title: "S1"},
		},
	})
	require.NoError(t, err)
	id := created.OrderServices[0].ID

	_, err = svc.Update(ctx, created.ID, UpdateOrderDTO{
		OrderServices: []OrderServiceInput{
			{ID: &id, Title: "X"},
		},
		RemovedServiceIDs: []uuid.UUID{id},
	})
	require.ErrorIs(t, err, ErrServiceNotInOrder)
}

func TestService_Update_ServiceFromAnotherOrderFails(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	other, err := svc.Create(ctx, CreateOrderDTO{
		Title: "Outro",
		OrderServices: []OrderServiceInput{
			{Title: "Alheio"},
		},
	})
	require.NoError(t, err)
	foreignID := other.OrderServices[0].ID

	target, err := svc.Create(ctx, CreateOrderDTO{
		Title:         "Alvo",
		OrderServices: []OrderServiceInput{{Title: "Próprio"}},
	})
	require.NoError(t, err)

	_, err = svc.Update(ctx, target.ID, UpdateOrderDTO{
		OrderServices: []OrderServiceInput{
			{ID: &foreignID, Title: "Tentativa"},
		},
	})
	require.ErrorIs(t, err, ErrServiceNotInOrder)
}

func TestService_Update_RemoveUnknownIDFails(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	created, err := svc.Create(ctx, CreateOrderDTO{
		Title:         "Pedido",
		OrderServices: []OrderServiceInput{{Title: "S1"}},
	})
	require.NoError(t, err)

	_, err = svc.Update(ctx, created.ID, UpdateOrderDTO{
		RemovedServiceIDs: []uuid.UUID{uuid.New()},
	})
	require.ErrorIs(t, err, ErrServiceNotInOrder)
}

func TestService_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	_, err := svc.Update(ctx, uuid.New(), UpdateOrderDTO{
		Title: ptrStr("X"),
		OrderServices: []OrderServiceInput{
			{Title: "Y"},
		},
	})
	require.ErrorIs(t, err, ErrNotFound)
}

func indexByID(services []OrderServiceResponse) map[uuid.UUID]OrderServiceResponse {
	out := make(map[uuid.UUID]OrderServiceResponse, len(services))
	for _, s := range services {
		out[s.ID] = s
	}
	return out
}
