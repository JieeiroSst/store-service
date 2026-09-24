package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type stubUsecase struct {
	created *model.Account
	query   port.ListQuery
}

func (s *stubUsecase) Create(_ context.Context, a *model.Account) (*model.Account, error) {
	s.created = a
	return a, nil
}
func (s *stubUsecase) Get(context.Context, uint) (*model.Account, error) {
	return nil, common.ErrNotFound
}
func (s *stubUsecase) List(_ context.Context, q port.ListQuery) ([]model.Account, int64, error) {
	s.query = q
	return nil, 7, nil
}
func (s *stubUsecase) Update(context.Context, uint, *model.Account) (*model.Account, error) {
	return nil, common.ErrNotFound
}
func (s *stubUsecase) Delete(context.Context, uint) error { return nil }

func newTestRouter(uc *stubUsecase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return newRouterWith(uc, "")
}

func newRouterWith(uc *stubUsecase, apiKey string) *gin.Engine {
	h := &crudHandler[model.Account, *model.Account]{path: "/accounts", uc: uc, filters: []string{"account_phone"}}
	cfg := &config.Config{Server: config.ServerConfig{APIKey: apiKey}}
	return NewRouter(RouterParams{Config: cfg, Auth: NewAuthenticator(cfg), Resources: []Resource{h}})
}

func do(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreate_IgnoresClientID(t *testing.T) {
	uc := &stubUsecase{}
	w := do(newTestRouter(uc), "POST", "/api/v1/accounts", `{"id": 99, "account_name": "acme"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	if uc.created.ID != 0 {
		t.Fatalf("client-supplied id leaked through: %d", uc.created.ID)
	}
}

func TestCreate_ValidatesRequiredFields(t *testing.T) {
	w := do(newTestRouter(&stubUsecase{}), "POST", "/api/v1/accounts", `{}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestGet_MapsErrors(t *testing.T) {
	r := newTestRouter(&stubUsecase{})
	if w := do(r, "GET", "/api/v1/accounts/1", ""); w.Code != http.StatusNotFound {
		t.Fatalf("missing: status = %d, want 404", w.Code)
	}
	if w := do(r, "GET", "/api/v1/accounts/abc", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("bad id: status = %d, want 400", w.Code)
	}
}

func TestList_PassesWhitelistedFiltersSearchAndTotal(t *testing.T) {
	uc := &stubUsecase{}
	w := do(newTestRouter(uc), "GET", "/api/v1/accounts?q=acme&account_phone=555&secret=x&limit=9999", "")

	if w.Code != http.StatusOK || w.Header().Get("X-Total-Count") != "7" {
		t.Fatalf("status = %d, total = %q", w.Code, w.Header().Get("X-Total-Count"))
	}
	if uc.query.Search != "acme" || uc.query.Equals["account_phone"] != uint64(555) {
		t.Fatalf("query = %+v", uc.query)
	}
	if _, leaked := uc.query.Equals["secret"]; leaked {
		t.Fatal("non-whitelisted filter reached the use case")
	}
	if uc.query.Limit != defaultLimit {
		t.Fatalf("limit = %d, want clamped to %d", uc.query.Limit, defaultLimit)
	}
}

func TestAPIKey(t *testing.T) {
	r := newRouterWith(&stubUsecase{}, "s3cret")

	if w := do(r, "GET", "/health", ""); w.Code != http.StatusOK {
		t.Fatalf("health should stay open, got %d", w.Code)
	}
	if w := do(r, "GET", "/api/v1/accounts", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("no key: status = %d, want 401", w.Code)
	}

	req := httptest.NewRequest("GET", "/api/v1/accounts", nil)
	req.Header.Set("X-API-Key", "s3cret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("with key: status = %d, want 200", w.Code)
	}
}
