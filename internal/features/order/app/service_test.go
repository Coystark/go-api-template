package orderapp

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	orderdomain "github.com/caiohenrique/go-api-template/internal/features/order/domain"
	"github.com/caiohenrique/go-api-template/internal/platform/pagination"
	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	orders  map[uuid.UUID]*orderdomain.Order
	deleted map[uuid.UUID]struct{}
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		orders:  make(map[uuid.UUID]*orderdomain.Order),
		deleted: make(map[uuid.UUID]struct{}),
	}
}

func (f *fakeRepo) Create(ctx context.Context, o *orderdomain.Order, services []orderdomain.OrderService) error {
	_ = ctx
	cp := *o
	cp.OrderServices = append([]orderdomain.OrderService(nil), services...)
	f.orders[o.ID] = &cp
	return nil
}

func (f *fakeRepo) FindByID(ctx context.Context, id uuid.UUID) (*orderdomain.Order, error) {
	_ = ctx
	o, ok := f.orders[id]
	if !ok || f.isDeleted(id) {
		return nil, orderdomain.ErrNotFound
	}
	cp := *o
	cp.OrderServices = append([]orderdomain.OrderService(nil), o.OrderServices...)
	return &cp, nil
}

func (f *fakeRepo) List(ctx context.Context, in ListOrdersRepoInput) ([]orderdomain.Order, int64, error) {
	_ = ctx
	orders := make([]orderdomain.Order, 0, len(f.orders))
	for id, o := range f.orders {
		if f.isDeleted(id) {
			continue
		}
		if in.Title != "" && !strings.Contains(strings.ToLower(o.Title), strings.ToLower(in.Title)) {
			continue
		}
		if in.Code != "" {
			if o.Code == nil || !strings.Contains(strings.ToLower(*o.Code), strings.ToLower(in.Code)) {
				continue
			}
		}
		cp := *o
		cp.OrderServices = nil
		orders = append(orders, cp)
	}

	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})

	total := int64(len(orders))
	start := pagination.Offset(in.Query)
	if start >= len(orders) {
		return []orderdomain.Order{}, total, nil
	}
	end := start + in.PageSize
	if end > len(orders) {
		end = len(orders)
	}
	return orders[start:end], total, nil
}

func (f *fakeRepo) ListOrderServices(ctx context.Context, in ListOrderServicesRepoInput) ([]orderdomain.OrderService, int64, error) {
	_ = ctx
	var rows []orderdomain.OrderService
	for oid, o := range f.orders {
		if f.isDeleted(oid) {
			continue
		}
		if in.OrderID != nil && oid != *in.OrderID {
			continue
		}
		for _, svc := range o.OrderServices {
			cp := svc
			rows = append(rows, cp)
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		ci, cj := rows[i].CreatedAt, rows[j].CreatedAt
		if !ci.Equal(cj) {
			return ci.After(cj)
		}
		return rows[i].ID.String() < rows[j].ID.String()
	})

	total := int64(len(rows))
	start := pagination.Offset(in.Query)
	if start >= len(rows) {
		return []orderdomain.OrderService{}, total, nil
	}
	end := start + in.PageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end], total, nil
}

func (f *fakeRepo) Update(ctx context.Context, orderID uuid.UUID, in UpdateOrderRepoInput) error {
	_ = ctx
	o, ok := f.orders[orderID]
	if !ok || f.isDeleted(orderID) {
		return orderdomain.ErrNotFound
	}

	o.Title = in.Title
	o.Subject = in.Subject
	o.Code = in.Code
	o.SentAt = in.SentAt
	o.ConvertedAt = in.ConvertedAt

	for _, svc := range in.Updates {
		idx := indexOfService(o.OrderServices, svc.ID)
		if idx < 0 {
			return orderdomain.ErrServiceNotInOrder
		}
		o.OrderServices[idx].Title = svc.Title
		o.OrderServices[idx].StartDate = svc.StartDate
		o.OrderServices[idx].EndDate = svc.EndDate
		o.OrderServices[idx].Observations = svc.Observations
	}

	if len(in.Creates) > 0 {
		o.OrderServices = append(o.OrderServices, in.Creates...)
	}

	for _, rid := range in.RemoveIDs {
		idx := indexOfService(o.OrderServices, rid)
		if idx < 0 {
			return orderdomain.ErrServiceNotInOrder
		}
		o.OrderServices = append(o.OrderServices[:idx], o.OrderServices[idx+1:]...)
	}

	return nil
}

func (f *fakeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_ = ctx
	if _, ok := f.orders[id]; !ok || f.isDeleted(id) {
		return orderdomain.ErrNotFound
	}
	f.deleted[id] = struct{}{}
	return nil
}

func (f *fakeRepo) isDeleted(id uuid.UUID) bool {
	_, ok := f.deleted[id]
	return ok
}

func indexOfService(services []orderdomain.OrderService, id uuid.UUID) int {
	for i := range services {
		if services[i].ID == id {
			return i
		}
	}
	return -1
}

func ptrStr(s string) *string        { return &s }
func ptrTime(t time.Time) *time.Time { return &t }

// datas fixas para serviços nos testes
var testSvcStart = time.Date(2026, 5, 11, 9, 0, 0, 0, time.UTC)
var testSvcStart2 = time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
var testSvcEnd = time.Date(2026, 5, 20, 18, 0, 0, 0, time.UTC)

func TestService_Create_WithServices(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	sentAt := time.Date(2026, 5, 11, 10, 30, 0, 0, time.UTC)
	out, err := svc.Create(ctx, CreateOrderInput{
		Title:   "Pedido A",
		Subject: ptrStr("Assunto A"),
		Code:    ptrStr("COD-123"),
		SentAt:  ptrTime(sentAt),
		OrderServices: []OrderServiceItem{
			{Title: "Serviço 1", StartDate: testSvcStart},
			{Title: "Serviço 2", StartDate: testSvcStart2, EndDate: ptrTime(testSvcEnd), Observations: ptrStr("nota B")},
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
	require.True(t, out.OrderServices[0].StartDate.Equal(testSvcStart))
	require.Nil(t, out.OrderServices[0].EndDate)
	require.Nil(t, out.OrderServices[0].Observations)
	require.Equal(t, "Serviço 2", out.OrderServices[1].Title)
	require.True(t, out.OrderServices[1].StartDate.Equal(testSvcStart2))
	require.NotNil(t, out.OrderServices[1].EndDate)
	require.True(t, out.OrderServices[1].EndDate.Equal(testSvcEnd))
	require.Equal(t, "nota B", *out.OrderServices[1].Observations)
	require.NotEqual(t, uuid.Nil, out.ID)
}

func TestService_Create_EmptyServices(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	out, err := svc.Create(ctx, CreateOrderInput{
		Title:         "Sem serviços",
		OrderServices: nil,
	})
	require.NoError(t, err)
	require.Empty(t, out.OrderServices)
}

func TestService_List_PaginatesAndSummarizes(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	first, err := svc.Create(ctx, CreateOrderInput{
		Title:         "Primeiro",
		OrderServices: []OrderServiceItem{{Title: "S1", StartDate: testSvcStart}},
	})
	require.NoError(t, err)
	second, err := svc.Create(ctx, CreateOrderInput{
		Title:         "Segundo",
		OrderServices: []OrderServiceItem{{Title: "S2", StartDate: testSvcStart}},
	})
	require.NoError(t, err)
	third, err := svc.Create(ctx, CreateOrderInput{
		Title:         "Terceiro",
		OrderServices: []OrderServiceItem{{Title: "S3", StartDate: testSvcStart}},
	})
	require.NoError(t, err)

	repo.orders[first.ID].CreatedAt = time.Date(2026, 5, 11, 9, 0, 0, 0, time.UTC)
	repo.orders[second.ID].CreatedAt = time.Date(2026, 5, 12, 9, 0, 0, 0, time.UTC)
	repo.orders[third.ID].CreatedAt = time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)

	out, err := svc.List(ctx, ListOrdersInput{Query: pagination.Query{Page: 1, PageSize: 2}})
	require.NoError(t, err)
	require.Equal(t, 1, out.Page)
	require.Equal(t, 2, out.PageSize)
	require.Equal(t, int64(3), out.Total)
	require.Equal(t, 2, out.TotalPages)
	require.Len(t, out.Items, 2)
	require.Equal(t, third.ID, out.Items[0].ID)
	require.Equal(t, second.ID, out.Items[1].ID)
	require.Empty(t, out.Items[0].OrderServices)
	require.Empty(t, out.Items[1].OrderServices)

	out, err = svc.List(ctx, ListOrdersInput{Query: pagination.Query{Page: 2, PageSize: 2}})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	require.Equal(t, first.ID, out.Items[0].ID)
	require.Empty(t, out.Items[0].OrderServices)
}

func TestService_List_FiltersByCodeAndTitle(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	_, err := svc.Create(ctx, CreateOrderInput{
		Title:         "Alpha Pedido",
		Code:          ptrStr("ORD-100"),
		OrderServices: []OrderServiceItem{{Title: "S", StartDate: testSvcStart}},
	})
	require.NoError(t, err)
	_, err = svc.Create(ctx, CreateOrderInput{
		Title:         "Beta outro",
		Code:          ptrStr("ORD-200"),
		OrderServices: []OrderServiceItem{{Title: "S", StartDate: testSvcStart}},
	})
	require.NoError(t, err)
	_, err = svc.Create(ctx, CreateOrderInput{
		Title:         "Gamma",
		OrderServices: []OrderServiceItem{{Title: "S", StartDate: testSvcStart}},
	})
	require.NoError(t, err)

	byCode, err := svc.List(ctx, ListOrdersInput{Query: pagination.Query{Page: 1, PageSize: 10}, Code: "ord-1"})
	require.NoError(t, err)
	require.Equal(t, int64(1), byCode.Total)
	require.Equal(t, "Alpha Pedido", byCode.Items[0].Title)

	byTitle, err := svc.List(ctx, ListOrdersInput{Query: pagination.Query{Page: 1, PageSize: 10}, Title: "BETA"})
	require.NoError(t, err)
	require.Equal(t, int64(1), byTitle.Total)
	require.Equal(t, "Beta outro", byTitle.Items[0].Title)

	both, err := svc.List(ctx, ListOrdersInput{Query: pagination.Query{Page: 1, PageSize: 10}, Code: "ORD", Title: "pha"})
	require.NoError(t, err)
	require.Equal(t, int64(1), both.Total)
	require.Equal(t, "Alpha Pedido", both.Items[0].Title)
}

func TestService_List_DefaultsAndCapsPageSize(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	out, err := svc.List(ctx, ListOrdersInput{Query: pagination.Query{PageSize: 500}})
	require.NoError(t, err)
	require.Equal(t, 1, out.Page)
	require.Equal(t, 100, out.PageSize)
	require.Equal(t, int64(0), out.Total)
	require.Equal(t, 0, out.TotalPages)
}

func TestService_List_InvalidPageFails(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	_, err := svc.List(ctx, ListOrdersInput{Query: pagination.Query{Page: -1, PageSize: 20}})
	require.Error(t, err)
}

func TestService_Delete_SoftDeletesOrder(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	created, err := svc.Create(ctx, CreateOrderInput{
		Title:         "Pedido",
		OrderServices: []OrderServiceItem{{Title: "S1", StartDate: testSvcStart}},
	})
	require.NoError(t, err)

	require.NoError(t, svc.Delete(ctx, created.ID))

	_, err = svc.GetByID(ctx, created.ID)
	require.ErrorIs(t, err, orderdomain.ErrNotFound)

	list, err := svc.List(ctx, ListOrdersInput{})
	require.NoError(t, err)
	require.Empty(t, list.Items)
}

func TestService_Delete_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	err := svc.Delete(ctx, uuid.Must(uuid.NewV7()))
	require.ErrorIs(t, err, orderdomain.ErrNotFound)
}

func TestService_Update_PartialUpsert(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	created, err := svc.Create(ctx, CreateOrderInput{
		Title: "Original",
		OrderServices: []OrderServiceItem{
			{Title: "Antigo 1", StartDate: testSvcStart},
			{Title: "Antigo 2", StartDate: testSvcStart},
			{Title: "Antigo 3", StartDate: testSvcStart},
		},
	})
	require.NoError(t, err)
	require.Len(t, created.OrderServices, 3)
	keepID := created.OrderServices[0].ID
	editID := created.OrderServices[1].ID
	untouchedID := created.OrderServices[2].ID

	convertedAt := time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)
	out, err := svc.Update(ctx, created.ID, UpdateOrderInput{
		Title:       "Atualizado",
		ConvertedAt: ptrTime(convertedAt),
		OrderServices: []OrderServiceItem{
			{ID: &editID, Title: "Editado", StartDate: testSvcStart2},
			{Title: "Novo", StartDate: testSvcStart},
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

	created, err := svc.Create(ctx, CreateOrderInput{
		Title: "Original",
		OrderServices: []OrderServiceItem{
			{Title: "S1", StartDate: testSvcStart},
		},
	})
	require.NoError(t, err)

	out, err := svc.Update(ctx, created.ID, UpdateOrderInput{
		Title: "Só título",
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

	created, err := svc.Create(ctx, CreateOrderInput{
		Title: "Pedido",
		OrderServices: []OrderServiceItem{
			{Title: "Manter", StartDate: testSvcStart},
			{Title: "Remover", StartDate: testSvcStart},
		},
	})
	require.NoError(t, err)
	keepID := created.OrderServices[0].ID
	removeID := created.OrderServices[1].ID

	out, err := svc.Update(ctx, created.ID, UpdateOrderInput{
		Title:             "Pedido",
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

	created, err := svc.Create(ctx, CreateOrderInput{
		Title: "Pedido",
		OrderServices: []OrderServiceItem{
			{Title: "S1", StartDate: testSvcStart},
		},
	})
	require.NoError(t, err)
	id := created.OrderServices[0].ID

	_, err = svc.Update(ctx, created.ID, UpdateOrderInput{
		Title: "Pedido",
		OrderServices: []OrderServiceItem{
			{ID: &id, Title: "X", StartDate: testSvcStart2},
		},
		RemovedServiceIDs: []uuid.UUID{id},
	})
	require.ErrorIs(t, err, orderdomain.ErrServiceNotInOrder)
}

func TestService_Update_ServiceFromAnotherOrderFails(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	other, err := svc.Create(ctx, CreateOrderInput{
		Title: "Outro",
		OrderServices: []OrderServiceItem{
			{Title: "Alheio", StartDate: testSvcStart},
		},
	})
	require.NoError(t, err)
	foreignID := other.OrderServices[0].ID

	target, err := svc.Create(ctx, CreateOrderInput{
		Title:         "Alvo",
		OrderServices: []OrderServiceItem{{Title: "Próprio", StartDate: testSvcStart}},
	})
	require.NoError(t, err)

	_, err = svc.Update(ctx, target.ID, UpdateOrderInput{
		Title: "Alvo",
		OrderServices: []OrderServiceItem{
			{ID: &foreignID, Title: "Tentativa", StartDate: testSvcStart},
		},
	})
	require.ErrorIs(t, err, orderdomain.ErrServiceNotInOrder)
}

func TestService_Update_RemoveUnknownIDFails(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	created, err := svc.Create(ctx, CreateOrderInput{
		Title:         "Pedido",
		OrderServices: []OrderServiceItem{{Title: "S1", StartDate: testSvcStart}},
	})
	require.NoError(t, err)

	_, err = svc.Update(ctx, created.ID, UpdateOrderInput{
		Title:             "Pedido",
		RemovedServiceIDs: []uuid.UUID{uuid.Must(uuid.NewV7())},
	})
	require.ErrorIs(t, err, orderdomain.ErrServiceNotInOrder)
}

func TestService_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	_, err := svc.Update(ctx, uuid.Must(uuid.NewV7()), UpdateOrderInput{
		Title: "X",
		OrderServices: []OrderServiceItem{
			{Title: "Y", StartDate: testSvcStart},
		},
	})
	require.ErrorIs(t, err, orderdomain.ErrNotFound)
}

func TestService_Update_ClearsOptionalServiceFieldsWhenOmitted(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	end := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	created, err := svc.Create(ctx, CreateOrderInput{
		Title: "Pedido",
		OrderServices: []OrderServiceItem{
			{
				Title:        "S1",
				StartDate:    testSvcStart,
				EndDate:      ptrTime(end),
				Observations: ptrStr("será limpo"),
			},
		},
	})
	require.NoError(t, err)
	sid := created.OrderServices[0].ID

	out, err := svc.Update(ctx, created.ID, UpdateOrderInput{
		Title: "Pedido",
		OrderServices: []OrderServiceItem{
			{ID: &sid, Title: "S1 renomeado", StartDate: testSvcStart2},
		},
	})
	require.NoError(t, err)
	require.Len(t, out.OrderServices, 1)
	require.Equal(t, "S1 renomeado", out.OrderServices[0].Title)
	require.True(t, out.OrderServices[0].StartDate.Equal(testSvcStart2))
	require.Nil(t, out.OrderServices[0].EndDate)
	require.Nil(t, out.OrderServices[0].Observations)
}

func TestService_ListOrderServices_WithoutOrderID(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	a, err := svc.Create(ctx, CreateOrderInput{
		Title:         "A",
		OrderServices: []OrderServiceItem{{Title: "Sa", StartDate: testSvcStart}},
	})
	require.NoError(t, err)
	b, err := svc.Create(ctx, CreateOrderInput{
		Title:         "B",
		OrderServices: []OrderServiceItem{{Title: "Sb", StartDate: testSvcStart}},
	})
	require.NoError(t, err)

	tEarly := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	tLate := time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)
	repo.orders[a.ID].OrderServices[0].CreatedAt = tEarly
	repo.orders[b.ID].OrderServices[0].CreatedAt = tLate

	out, err := svc.ListOrderServices(ctx, ListOrderServicesInput{
		Query: pagination.Query{Page: 1, PageSize: 10},
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), out.Total)
	require.Len(t, out.Items, 2)
	require.Equal(t, "Sb", out.Items[0].Title)
	require.Equal(t, "Sa", out.Items[1].Title)
}

func TestService_ListOrderServices_WithOrderID(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	a, err := svc.Create(ctx, CreateOrderInput{
		Title:         "A",
		OrderServices: []OrderServiceItem{{Title: "S1", StartDate: testSvcStart}},
	})
	require.NoError(t, err)
	_, err = svc.Create(ctx, CreateOrderInput{
		Title:         "B",
		OrderServices: []OrderServiceItem{{Title: "S2", StartDate: testSvcStart}},
	})
	require.NoError(t, err)

	out, err := svc.ListOrderServices(ctx, ListOrderServicesInput{
		Query:   pagination.Query{Page: 1, PageSize: 10},
		OrderID: a.ID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), out.Total)
	require.Len(t, out.Items, 1)
	require.Equal(t, "S1", out.Items[0].Title)
	require.Equal(t, a.OrderServices[0].ID, out.Items[0].ID)
}

func TestService_ListOrderServices_Paginates(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	o, err := svc.Create(ctx, CreateOrderInput{
		Title: "O",
		OrderServices: []OrderServiceItem{
			{Title: "S1", StartDate: testSvcStart},
			{Title: "S2", StartDate: testSvcStart},
			{Title: "S3", StartDate: testSvcStart},
		},
	})
	require.NoError(t, err)
	t1 := time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	repo.orders[o.ID].OrderServices[0].CreatedAt = t1
	repo.orders[o.ID].OrderServices[1].CreatedAt = t2
	repo.orders[o.ID].OrderServices[2].CreatedAt = t3

	out, err := svc.ListOrderServices(ctx, ListOrderServicesInput{
		Query:   pagination.Query{Page: 1, PageSize: 2},
		OrderID: o.ID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, int64(3), out.Total)
	require.Len(t, out.Items, 2)
	require.Equal(t, "S3", out.Items[0].Title)
	require.Equal(t, "S2", out.Items[1].Title)

	out, err = svc.ListOrderServices(ctx, ListOrderServicesInput{
		Query:   pagination.Query{Page: 2, PageSize: 2},
		OrderID: o.ID.String(),
	})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	require.Equal(t, "S1", out.Items[0].Title)
}

func TestService_ListOrderServices_DeletedOrderExcluded(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewService(repo, validator.New())

	o, err := svc.Create(ctx, CreateOrderInput{
		Title:         "Será removido",
		OrderServices: []OrderServiceItem{{Title: "Sx", StartDate: testSvcStart}},
	})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(ctx, o.ID))

	out, err := svc.ListOrderServices(ctx, ListOrderServicesInput{
		Query: pagination.Query{Page: 1, PageSize: 10},
	})
	require.NoError(t, err)
	require.Empty(t, out.Items)
	require.Equal(t, int64(0), out.Total)

	out, err = svc.ListOrderServices(ctx, ListOrderServicesInput{
		Query:   pagination.Query{Page: 1, PageSize: 10},
		OrderID: o.ID.String(),
	})
	require.NoError(t, err)
	require.Empty(t, out.Items)
	require.Equal(t, int64(0), out.Total)
}

func TestService_ListOrderServices_InvalidOrderID(t *testing.T) {
	ctx := context.Background()
	svc := NewService(newFakeRepo(), validator.New())

	_, err := svc.ListOrderServices(ctx, ListOrderServicesInput{
		Query:   pagination.Query{Page: 1, PageSize: 10},
		OrderID: "not-a-uuid",
	})
	require.Error(t, err)
}

func indexByID(services []OrderServiceView) map[uuid.UUID]OrderServiceView {
	out := make(map[uuid.UUID]OrderServiceView, len(services))
	for _, s := range services {
		out[s.ID] = s
	}
	return out
}
