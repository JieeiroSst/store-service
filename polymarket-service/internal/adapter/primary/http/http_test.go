package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func do(h http.Handler, method, path, body string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// NewRouter panics on conflicting gin wildcards, so building it is a test.
func TestRouterRegisters(t *testing.T) {
	r := NewRouter(&Handler{adminToken: "secret"})
	if rec := do(r, "GET", "/health", ""); rec.Code != 200 {
		t.Fatalf("health %d", rec.Code)
	}
}

func TestAdminEndpointsFailClosed(t *testing.T) {
	for name, token := range map[string]string{"no token configured": "", "token configured": "secret"} {
		r := NewRouter(&Handler{adminToken: token})
		for _, path := range []string{"/api/v1/events", "/api/v1/markets/1/propose", "/api/v1/markets/1/resolve"} {
			for _, header := range []string{"", "wrong", ""} {
				var rec *httptest.ResponseRecorder
				if header == "" {
					rec = do(r, "POST", path, `{}`)
				} else {
					rec = do(r, "POST", path, `{}`, "X-Admin-Token", header)
				}
				if rec.Code != http.StatusUnauthorized {
					t.Errorf("%s: %s with %q got %d", name, path, header, rec.Code)
				}
			}
		}
	}
	// With the right token the request reaches the handler (which then rejects the empty body).
	r := NewRouter(&Handler{adminToken: "secret"})
	if rec := do(r, "POST", "/api/v1/events", `{}`, "X-Admin-Token", "secret"); rec.Code != http.StatusBadRequest {
		t.Fatalf("authorised request should reach validation, got %d", rec.Code)
	}
}

func TestPrivateRoutesRequireMatchingGatewayUser(t *testing.T) {
	r := NewRouter(&Handler{})
	for _, path := range []string{"/api/v1/users/ann/balance", "/api/v1/users/ann/portfolio", "/api/v1/users/ann/orders", "/api/v1/users/ann/watchlist"} {
		if rec := do(r, "GET", path, "", "X-User-Id", "eve"); rec.Code != http.StatusForbidden {
			t.Errorf("%s as eve: got %d", path, rec.Code)
		}
	}
	if rec := do(r, "POST", "/api/v1/users/ann/withdraw", `{"amount":5}`, "X-User-Id", "eve"); rec.Code != http.StatusForbidden {
		t.Errorf("withdraw as eve: got %d", rec.Code)
	}
	if rec := do(r, "PUT", "/api/v1/users/ann/profile", `{"username":"x"}`, "X-User-Id", "eve"); rec.Code != http.StatusForbidden {
		t.Errorf("profile edit as eve: got %d", rec.Code)
	}
}

func TestBodyIdentityMustAgreeWithGatewayHeader(t *testing.T) {
	r := NewRouter(&Handler{})
	body := `{"user_id":"ann","outcome":"yes","side":"buy","price":50,"size":1}`
	if rec := do(r, "POST", "/api/v1/markets/1/orders", body, "X-User-Id", "eve"); rec.Code != http.StatusForbidden {
		t.Fatalf("spoofed user_id must be refused, got %d", rec.Code)
	}
	if rec := do(r, "DELETE", "/api/v1/orders/1?user_id=ann", "", "X-User-Id", "eve"); rec.Code != http.StatusForbidden {
		t.Fatalf("spoofed cancel must be refused, got %d", rec.Code)
	}
}

func TestBadPathIDs(t *testing.T) {
	r := NewRouter(&Handler{})
	for _, path := range []string{"/api/v1/markets/abc/book", "/api/v1/markets/0/trades", "/api/v1/orders/x"} {
		if rec := do(r, "GET", path, ""); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: got %d", path, rec.Code)
		}
	}
}

func TestErrorStatusMapping(t *testing.T) {
	cases := map[error]int{
		port.ErrNotFound:            404,
		port.ErrWalletNotFound:      404,
		port.ErrForbidden:           403,
		port.ErrUnauthorized:        401,
		port.ErrAlreadyExists:       409,
		port.ErrMarketNotTradable:   409,
		port.ErrNotFilled:           409,
		port.ErrInvalidPrice:        400,
		port.ErrInvalidInput:        400,
		port.ErrInsufficientBalance: 422,
		port.ErrInsufficientShares:  422,
		port.ErrWalletUnavailable:   502,
		errors.New("secret detail"): 500,
	}
	for err, want := range cases {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		writeError(c, fmt.Errorf("wrapped: %w", err))
		if rec.Code != want {
			t.Errorf("%v: got %d want %d", err, rec.Code, want)
		}
		if want == 500 && strings.Contains(rec.Body.String(), "secret detail") {
			t.Error("internal errors must not leak their message")
		}
	}
}

type stubIdentity map[string]string

func (s stubIdentity) Authenticate(_ context.Context, token string) (string, error) {
	if id, ok := s[token]; ok {
		return id, nil
	}
	return "", port.ErrUnauthenticated
}

func tokenRouter() http.Handler {
	return NewRouter(&Handler{tokenAuth: true, adminToken: "secret", identity: stubIdentity{"ann-token": "ann", "eve-token": "eve"}})
}

func bearer(token string) string { return "Bearer " + token }

func TestTokenModeRequiresAValidBearerToken(t *testing.T) {
	r := tokenRouter()
	for name, headers := range map[string][]string{
		"no credentials":         nil,
		"gateway header ignored": {"X-User-Id", "ann"},
		"unknown token":          {"Authorization", bearer("nope")},
		"not a bearer scheme":    {"Authorization", "Basic abc"},
	} {
		if rec := do(r, "GET", "/api/v1/users/ann/balance", "", headers...); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: got %d", name, rec.Code)
		}
	}
}

func TestTokenModeRefusesABadTokenEvenOnPublicRoutes(t *testing.T) {
	if rec := do(tokenRouter(), "GET", "/api/v1/categories", "", "Authorization", bearer("nope")); rec.Code != http.StatusUnauthorized {
		t.Fatalf("a presented-but-invalid token is an error, not anonymous access: %d", rec.Code)
	}
}

func TestTokenModeOnlyLetsUsersIntoTheirOwnAccount(t *testing.T) {
	r := tokenRouter()
	if rec := do(r, "GET", "/api/v1/users/ann/balance", "", "Authorization", bearer("eve-token")); rec.Code != http.StatusForbidden {
		t.Fatalf("eve reading ann: %d", rec.Code)
	}
	// ann passes the ownership check and reaches the handler, which rejects the empty body.
	if rec := do(r, "POST", "/api/v1/users/ann/referral/redeem", `{}`, "Authorization", bearer("ann-token")); rec.Code != http.StatusBadRequest {
		t.Fatalf("ann reaching her own endpoint: %d", rec.Code)
	}
}

func TestTokenModeBindsBodyIdentityToTheToken(t *testing.T) {
	r := tokenRouter()
	spoof := `{"user_id":"ann","outcome":"yes","side":"buy","price":50,"size":1}`
	if rec := do(r, "POST", "/api/v1/markets/1/orders", spoof, "Authorization", bearer("eve-token")); rec.Code != http.StatusForbidden {
		t.Fatalf("eve trading as ann: %d", rec.Code)
	}
	if rec := do(r, "POST", "/api/v1/markets/1/orders", spoof); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous trading: %d", rec.Code)
	}
	if rec := do(r, "DELETE", "/api/v1/orders/1?user_id=ann", "", "Authorization", bearer("eve-token")); rec.Code != http.StatusForbidden {
		t.Fatalf("eve cancelling ann's order: %d", rec.Code)
	}
}

func TestNewAdminRoutesFailClosed(t *testing.T) {
	r := NewRouter(&Handler{adminToken: "secret"})
	for _, path := range []string{
		"/api/v1/events/1/propose", "/api/v1/events/1/resolve", "/api/v1/markets/1/rewards", "/api/v1/admin/exchange/fund",
	} {
		if rec := do(r, "POST", path, `{}`); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: %d", path, rec.Code)
		}
	}
	if rec := do(r, "GET", "/api/v1/admin/exchange", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("GET /admin/exchange: %d", rec.Code)
	}
}

func TestNewErrorMappings(t *testing.T) {
	cases := map[error]int{
		port.ErrUnauthenticated:    401,
		port.ErrReferralNotAllowed: 409,
		port.ErrNegRisk:            409,
		port.ErrNotNegRisk:         400,
		port.ErrUpstream:           502,
	}
	for err, want := range cases {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		writeError(c, fmt.Errorf("wrapped: %w", err))
		if rec.Code != want {
			t.Errorf("%v: got %d want %d", err, rec.Code, want)
		}
	}
}
