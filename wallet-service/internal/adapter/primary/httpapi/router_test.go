package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type stubWallets struct {
	port.WalletService
	err error
	got uuid.UUID
}

func (s *stubWallets) Get(_ context.Context, id uuid.UUID) (*model.Wallet, error) {
	s.got = id
	if s.err != nil {
		return nil, s.err
	}
	return &model.Wallet{ID: id, Currency: model.CurrencyUSD}, nil
}

type stubTransfers struct {
	cmd port.TransferCommand
	err error
}

func (s *stubTransfers) Transfer(_ context.Context, cmd port.TransferCommand) (*model.Transaction, error) {
	s.cmd = cmd
	if s.err != nil {
		return nil, s.err
	}
	return &model.Transaction{ID: uuid.New(), Type: model.TransactionTypeTransferOut}, nil
}

func newTestRouter(svc Services) http.Handler {
	return NewRouter(svc, RouterOptions{}, zap.NewNop())
}

func do(t *testing.T, h http.Handler, method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) errorResponse {
	t.Helper()
	var e errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("decode error body %q: %v", rec.Body.String(), err)
	}
	return e
}

func TestDomainErrorsMapToStatusAndCode(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{model.ErrWalletNotFound, http.StatusNotFound, "WALLET_NOT_FOUND"},
		{model.ErrInsufficientFunds, http.StatusUnprocessableEntity, "INSUFFICIENT_FUNDS"},
		{model.ErrPinLocked, http.StatusLocked, "PIN_LOCKED"},
		{model.ErrNotWalletOwner, http.StatusForbidden, "FORBIDDEN"},
		{model.ErrIdempotencyKeyReused, http.StatusConflict, "IDEMPOTENCY_KEY_REUSED"},
		{model.ErrBankDeclined, http.StatusBadGateway, "BANK_DECLINED"},
		{context.DeadlineExceeded, http.StatusInternalServerError, "INTERNAL"},
	}
	for _, tc := range cases {
		wallets := &stubWallets{err: tc.err}
		rec := do(t, newTestRouter(Services{Wallets: wallets}), http.MethodGet, "/api/v1/wallets/"+uuid.NewString(), nil, nil)
		if rec.Code != tc.status {
			t.Errorf("%v: expected status %d, got %d", tc.err, tc.status, rec.Code)
		}
		if got := decodeError(t, rec).Code; got != tc.code {
			t.Errorf("%v: expected code %s, got %s", tc.err, tc.code, got)
		}
	}
}

func TestInvalidPathIDIsRejected(t *testing.T) {
	wallets := &stubWallets{}
	rec := do(t, newTestRouter(Services{Wallets: wallets}), http.MethodGet, "/api/v1/wallets/not-a-uuid", nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if wallets.got != uuid.Nil {
		t.Fatalf("service must not be called with an invalid id")
	}
}

func TestTransferBindsRequestAndIdempotencyHeader(t *testing.T) {
	transfers := &stubTransfers{}
	h := newTestRouter(Services{Transfers: transfers})
	user, from, to := uuid.New(), uuid.New(), uuid.New()

	rec := do(t, h, http.MethodPost, "/api/v1/transfers", map[string]any{
		"user_id": user.String(), "from_wallet_id": from.String(), "to_wallet_id": to.String(),
		"amount": "12.50", "pin": "123456", "idempotency_key": "body-key", "unknown_field": true,
	}, map[string]string{"Idempotency-Key": "header-key"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if transfers.cmd.IdempotencyKey != "header-key" || transfers.cmd.UserID != user || transfers.cmd.ToWalletID != to {
		t.Fatalf("unexpected command: %+v", transfers.cmd)
	}
	if !transfers.cmd.Amount.Equal(decimal.RequireFromString("12.50")) {
		t.Fatalf("amount not parsed: %s", transfers.cmd.Amount)
	}

	rec = do(t, h, http.MethodPost, "/api/v1/transfers", map[string]any{
		"user_id": user.String(), "from_wallet_id": from.String(), "amount": 3, "pin": "123456", "idempotency_key": "body-only",
	}, nil)
	if rec.Code != http.StatusCreated || transfers.cmd.IdempotencyKey != "body-only" || transfers.cmd.ToWalletID != uuid.Nil {
		t.Fatalf("body idempotency key fallback failed: %d %+v", rec.Code, transfers.cmd)
	}

	rec = do(t, h, http.MethodPost, "/api/v1/transfers", map[string]any{"user_id": "nope", "from_wallet_id": from.String()}, nil)
	if rec.Code != http.StatusBadRequest || !strings.Contains(decodeError(t, rec).Error, "user_id") {
		t.Fatalf("expected user_id validation error, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestLegacyGatewayRoutesStillRegistered(t *testing.T) {
	engine := NewRouter(Services{}, RouterOptions{}, zap.NewNop())
	registered := map[string]bool{}
	for _, r := range engine.Routes() {
		registered[r.Method+" "+r.Path] = true
	}
	for _, route := range []string{
		"POST /api/v1/wallets",
		"GET /api/v1/wallets/:id",
		"POST /api/v1/wallets/:id/credit",
		"POST /api/v1/wallets/:id/freeze",
		"POST /api/v1/transactions/authorize",
		"POST /api/v1/transactions/:id/capture",
		"POST /api/v1/transactions/:id/void",
		"GET /api/v1/transactions/:id",
		"POST /api/v1/settlements/batches",
		"POST /api/v1/settlements/batches/:batch_id/clear",
		"POST /api/v1/settlements/batches/:batch_id/settle",
	} {
		if !registered[route] {
			t.Errorf("route used by gateway-api is missing: %s", route)
		}
	}
}

func TestReadinessReflectsDependency(t *testing.T) {
	healthy := NewRouter(Services{}, RouterOptions{Ready: func(context.Context) error { return nil }}, zap.NewNop())
	if rec := do(t, healthy, http.MethodGet, "/health/ready", nil, nil); rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	down := NewRouter(Services{}, RouterOptions{Ready: func(context.Context) error { return context.DeadlineExceeded }}, zap.NewNop())
	if rec := do(t, down, http.MethodGet, "/health/ready", nil, nil); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if rec := do(t, down, http.MethodGet, "/health", nil, nil); rec.Code != http.StatusOK {
		t.Fatalf("liveness must not depend on the database, got %d", rec.Code)
	}
}

type stubUserWallets struct {
	port.WalletService
	user uuid.UUID
}

func (s *stubUserWallets) ListByUser(_ context.Context, userID uuid.UUID) ([]*model.Wallet, error) {
	s.user = userID
	return []*model.Wallet{{ID: uuid.New(), UserID: userID, Currency: model.CurrencyUSD}}, nil
}

func TestNumericUserIDsMapToStableUUIDs(t *testing.T) {
	a, ok := UserUUID("7")
	b, _ := UserUUID(" 7 ")
	other, _ := UserUUID("8")
	if !ok || a != b || a == other || a.Version() != 5 {
		t.Fatalf("numeric ids must map to a stable v5 uuid: %v %v %v", a, b, other)
	}
	existing := uuid.New()
	if got, ok := UserUUID(existing.String()); !ok || got != existing {
		t.Fatalf("uuid ids must pass through unchanged, got %v", got)
	}
	for _, bad := range []string{"nope", "-3", "0", "1.5"} {
		if _, ok := UserUUID(bad); ok {
			t.Fatalf("%q must be rejected", bad)
		}
	}

	wallets := &stubUserWallets{}
	rec := do(t, newTestRouter(Services{Wallets: wallets}), http.MethodGet, "/api/v1/users/7/wallets", nil, nil)
	if rec.Code != http.StatusOK || wallets.user != a {
		t.Fatalf("expected numeric path id to resolve to %v, got %d %v: %s", a, rec.Code, wallets.user, rec.Body.String())
	}
}
