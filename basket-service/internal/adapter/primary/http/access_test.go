package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

// tokens: "alice" → user 1, "bob" → user 2, "root" → admin 3.
type fakeAuthn struct{}

func (fakeAuthn) Authenticate(_ context.Context, token string) (userservice.Identity, error) {
	switch token {
	case "alice":
		return userservice.Identity{UserID: "1", Role: "user"}, nil
	case "bob":
		return userservice.Identity{UserID: "2", Role: "user"}, nil
	case "root":
		return userservice.Identity{UserID: "3", Role: "super_admin"}, nil
	}
	return userservice.Identity{}, userservice.ErrUnauthenticated
}

type fakeBaskets struct {
	port.BasketUsecase
	byID map[int]*model.Basket
}

func (f *fakeBaskets) CreateBasket(_ context.Context, b *model.Basket) (*model.Basket, error) {
	b.ID = len(f.byID) + 100
	f.byID[b.ID] = b
	return b, nil
}

func (f *fakeBaskets) GetBasket(_ context.Context, id int) (*model.Basket, error) {
	b, ok := f.byID[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	return b, nil
}

func (f *fakeBaskets) ListBaskets(context.Context) ([]model.Basket, error) {
	var out []model.Basket
	for _, b := range f.byID {
		out = append(out, *b)
	}
	return out, nil
}

func (f *fakeBaskets) ListBasketsByUser(_ context.Context, userID int) ([]model.Basket, error) {
	var out []model.Basket
	for _, b := range f.byID {
		if b.UserID == userID {
			out = append(out, *b)
		}
	}
	return out, nil
}

func (f *fakeBaskets) UpdateBasket(_ context.Context, b *model.Basket) (*model.Basket, error) {
	f.byID[b.ID] = b
	return b, nil
}

func (f *fakeBaskets) DeleteBasket(_ context.Context, id int) error {
	delete(f.byID, id)
	return nil
}

func newTestRouter() (*gin.Engine, *fakeBaskets) {
	gin.SetMode(gin.TestMode)
	baskets := &fakeBaskets{byID: map[int]*model.Basket{
		10: {ID: 10, UserID: 1, Status: "Open"},
		20: {ID: 20, UserID: 2, Status: "Open"},
	}}
	h := NewHandler(baskets, nil, nil, nil, nil)
	return NewRouter(h, fakeAuthn{}), baskets
}

func call(r *gin.Engine, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestBasketOwnership(t *testing.T) {
	r, baskets := newTestRouter()

	cases := []struct {
		name, method, path, token, body string
		want                            int
	}{
		{"no token", http.MethodGet, "/api/v1/baskets/10", "", "", http.StatusUnauthorized},
		{"owner reads", http.MethodGet, "/api/v1/baskets/10", "alice", "", http.StatusOK},
		{"other user reads", http.MethodGet, "/api/v1/baskets/10", "bob", "", http.StatusNotFound},
		{"admin reads", http.MethodGet, "/api/v1/baskets/10", "root", "", http.StatusOK},
		{"other user updates", http.MethodPut, "/api/v1/baskets/10", "bob", `{"status":"Frozen"}`, http.StatusNotFound},
		{"other user deletes", http.MethodDelete, "/api/v1/baskets/10", "bob", "", http.StatusNotFound},
		{"other user reads profile", http.MethodGet, "/api/v1/users/1", "bob", "", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if w := call(r, tc.method, tc.path, tc.token, tc.body); w.Code != tc.want {
				t.Errorf("status = %d, want %d (%s)", w.Code, tc.want, w.Body.String())
			}
		})
	}
	if _, ok := baskets.byID[10]; !ok {
		t.Fatal("basket 10 was deleted by a non-owner")
	}

	// Create ignores a spoofed user_id and uses the caller.
	w := call(r, http.MethodPost, "/api/v1/baskets", "bob", `{"user_id":1,"status":"Open"}`)
	var created model.Basket
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if w.Code != http.StatusCreated || created.UserID != 2 {
		t.Errorf("create: status %d, owner %d; want 201 owned by 2", w.Code, created.UserID)
	}

	// Owner update can't transfer the basket.
	w = call(r, http.MethodPut, "/api/v1/baskets/10", "alice", `{"user_id":2,"status":"Frozen"}`)
	if w.Code != http.StatusOK || baskets.byID[10].UserID != 1 {
		t.Errorf("update: status %d, owner %d; want 200 owned by 1", w.Code, baskets.byID[10].UserID)
	}

	// Listing is scoped to the caller.
	w = call(r, http.MethodGet, "/api/v1/baskets", "alice", "")
	var mine []model.Basket
	_ = json.Unmarshal(w.Body.Bytes(), &mine)
	for _, b := range mine {
		if b.UserID != 1 {
			t.Errorf("alice sees basket %d of user %d", b.ID, b.UserID)
		}
	}
}
