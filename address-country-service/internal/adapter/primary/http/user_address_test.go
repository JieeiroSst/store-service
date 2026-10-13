package http

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"github.com/JIeeiroSst/address-country-service/pkg/authn"
	"gofr.dev/pkg/gofr"
)

type fakeRequest struct {
	ctx    context.Context
	params map[string]string
	body   any
}

func (r fakeRequest) Context() context.Context  { return r.ctx }
func (r fakeRequest) Param(string) string       { return "" }
func (r fakeRequest) PathParam(k string) string { return r.params[k] }
func (r fakeRequest) HostName() string          { return "" }
func (r fakeRequest) Params(string) []string    { return nil }
func (r fakeRequest) Bind(v interface{}) error {
	b, _ := json.Marshal(r.body)
	return json.Unmarshal(b, v)
}

type fakeAddresses struct {
	rows    map[int]model.UserAddress
	created *model.UserAddress
	deleted []int
}

func (f *fakeAddresses) Create(_ context.Context, a *model.UserAddress) error {
	f.created = a
	return nil
}
func (f *fakeAddresses) Get(_ context.Context, id int) (*model.UserAddress, error) {
	a, ok := f.rows[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &a, nil
}
func (f *fakeAddresses) Update(_ context.Context, a *model.UserAddress) error { return nil }
func (f *fakeAddresses) Delete(_ context.Context, id int) error {
	f.deleted = append(f.deleted, id)
	return nil
}
func (f *fakeAddresses) List(context.Context) ([]model.UserAddress, error) {
	out := []model.UserAddress{}
	for _, a := range f.rows {
		out = append(out, a)
	}
	return out, nil
}

func ctxFor(p *authn.Principal, params map[string]string, body any) *gofr.Context {
	ctx := context.Background()
	if p != nil {
		ctx = authn.WithPrincipal(ctx, *p)
	}
	return &gofr.Context{Context: ctx, Request: fakeRequest{ctx: ctx, params: params, body: body}}
}

func TestUserAddressOwnership(t *testing.T) {
	store := &fakeAddresses{rows: map[int]model.UserAddress{1: {ID: 1, UserID: 7}, 2: {ID: 2, UserID: 8}}}
	h := &Handler{userAddress: store}
	user := &authn.Principal{UserID: "7", Role: "user"}
	admin := &authn.Principal{UserID: "1", Role: "admin"}

	if _, err := h.ListUserAddresses(ctxFor(nil, nil, nil)); !errors.Is(err, authn.ErrUnauthenticated) {
		t.Fatalf("anonymous list err = %v", err)
	}
	got, _ := h.ListUserAddresses(ctxFor(user, nil, nil))
	if rows := got.([]model.UserAddress); len(rows) != 1 || rows[0].UserID != 7 {
		t.Fatalf("user list = %+v", rows)
	}
	got, _ = h.ListUserAddresses(ctxFor(admin, nil, nil))
	if rows := got.([]model.UserAddress); len(rows) != 2 {
		t.Fatalf("admin list = %+v", rows)
	}
	if _, err := h.GetUserAddress(ctxFor(user, map[string]string{"id": "2"}, nil)); !errors.Is(err, authn.ErrForbidden) {
		t.Fatalf("foreign get err = %v", err)
	}
	if _, err := h.GetUserAddress(ctxFor(user, map[string]string{"id": "1"}, nil)); err != nil {
		t.Fatalf("own get err = %v", err)
	}
	if _, err := h.DeleteUserAddress(ctxFor(user, map[string]string{"id": "2"}, nil)); !errors.Is(err, authn.ErrForbidden) || len(store.deleted) != 0 {
		t.Fatalf("foreign delete err = %v deleted=%v", err, store.deleted)
	}
	if _, err := h.CreateUserAddress(ctxFor(user, nil, map[string]any{"user_id": 8, "title": "home"})); err != nil || store.created.UserID != 7 {
		t.Fatalf("create forced user_id = %+v, %v", store.created, err)
	}
	upd, err := h.UpdateUserAddress(ctxFor(user, map[string]string{"id": "1"}, map[string]any{"user_id": 8}))
	if err != nil || upd.(model.UserAddress).UserID != 7 {
		t.Fatalf("update kept owner = %+v, %v", upd, err)
	}
}

func TestAdminDecorator(t *testing.T) {
	called := false
	h := admin(func(*gofr.Context) (interface{}, error) { called = true; return nil, nil })
	if _, err := h(ctxFor(&authn.Principal{UserID: "7", Role: "user"}, nil, nil)); !errors.Is(err, authn.ErrForbidden) || called {
		t.Fatalf("user reached admin route: %v", err)
	}
	if _, err := h(ctxFor(&authn.Principal{Service: true}, nil, nil)); err != nil || !called {
		t.Fatalf("service blocked: %v", err)
	}
}
