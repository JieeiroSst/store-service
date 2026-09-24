package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

type memRepo[T any] struct {
	items map[uint]*T
	next  uint
}

func newMemRepo[T any]() *memRepo[T] { return &memRepo[T]{items: map[uint]*T{}} }

func (m *memRepo[T]) Create(_ context.Context, e *T) error {
	m.next++
	m.items[m.next] = e
	return nil
}
func (m *memRepo[T]) GetByID(_ context.Context, id uint) (*T, error) {
	if e, ok := m.items[id]; ok {
		return e, nil
	}
	return nil, common.ErrNotFound
}
func (m *memRepo[T]) List(_ context.Context, q port.ListQuery) ([]T, int64, error) {
	var out []T
	for id := uint(1); id <= m.next; id++ {
		if e, ok := m.items[id]; ok {
			out = append(out, *e)
		}
	}
	total := int64(len(out))
	if q.Offset >= len(out) {
		return nil, total, nil
	}
	out = out[q.Offset:]
	if q.Limit > 0 && len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out, total, nil
}
func (m *memRepo[T]) Update(_ context.Context, id uint, e *T) error {
	m.items[id] = e
	return nil
}
func (m *memRepo[T]) Delete(_ context.Context, id uint) error {
	delete(m.items, id)
	return nil
}

type fakeNotifier struct {
	got []model.Notification
	err error
}

func (f *fakeNotifier) Notify(_ context.Context, n model.Notification) error {
	f.got = append(f.got, n)
	return f.err
}

func TestCRUD_UpdateAndDeleteRequireExisting(t *testing.T) {
	uc := NewCRUDService[model.Account](newMemRepo[model.Account]())
	ctx := context.Background()

	if _, err := uc.Update(ctx, 1, &model.Account{}); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("update missing: got %v, want ErrNotFound", err)
	}
	if err := uc.Delete(ctx, 1); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("delete missing: got %v, want ErrNotFound", err)
	}

	if _, err := uc.Create(ctx, &model.Account{Name: "acme"}); err != nil {
		t.Fatal(err)
	}
	got, err := uc.Update(ctx, 1, &model.Account{Name: "acme 2"})
	if err != nil || got.Name != "acme 2" {
		t.Fatalf("update: got %+v, %v", got, err)
	}
}

func TestLeadService_NotifiesOnCreate(t *testing.T) {
	n := &fakeNotifier{}
	uc := NewLeadService(newMemRepo[model.Lead](), n)

	if _, err := uc.Create(context.Background(), &model.Lead{FirstName: "Ada", Surname: "Lovelace"}); err != nil {
		t.Fatal(err)
	}
	if len(n.got) != 1 || n.got[0].Title != "New lead" {
		t.Fatalf("notifications = %+v", n.got)
	}
}

func TestLeadService_NotifierFailureDoesNotFailCreate(t *testing.T) {
	n := &fakeNotifier{err: errors.New("down")}
	uc := NewLeadService(newMemRepo[model.Lead](), n)

	if _, err := uc.Create(context.Background(), &model.Lead{FirstName: "Ada", Surname: "L"}); err != nil {
		t.Fatalf("create should succeed, got %v", err)
	}
}

func TestLeadService_ForcesStatusAndKeepsManagedFieldsOnUpdate(t *testing.T) {
	repo := newMemRepo[model.Lead]()
	uc := NewLeadService(repo, &fakeNotifier{})
	ctx := context.Background()

	created, err := uc.Create(ctx, &model.Lead{FirstName: "Ada", Surname: "L", Status: model.LeadConverted})
	if err != nil || created.Status != model.LeadNew {
		t.Fatalf("create: status = %q, err = %v; want %q", created.Status, err, model.LeadNew)
	}

	repo.items[1].Status = model.LeadConverted // as if a conversion had happened
	updated, err := uc.Update(ctx, 1, &model.Lead{FirstName: "Ada", Surname: "Byron", Status: model.LeadNew})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != model.LeadConverted || updated.Surname != "Byron" {
		t.Fatalf("update = %+v", updated)
	}
}
