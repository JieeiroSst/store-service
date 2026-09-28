package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/gin-gonic/gin"
)

type stubVending struct {
	port.VendingService
	err      error
	input    port.CheckoutInput
	dispense *bool
	reason   string
}

func (s *stubVending) Checkout(_ context.Context, id string, in port.CheckoutInput) (*domain.Checkout, error) {
	s.input = in
	if s.err != nil {
		return nil, s.err
	}
	return &domain.Checkout{
		Order:   &domain.Order{ID: "o1", OrderNo: 7, ReservationID: id, PaymentID: "p1", Status: domain.OrderProcessing},
		Payment: &domain.Payment{ID: "p1", AmountCents: 120, DiscountCents: 30, CouponCode: in.CouponCode, Method: in.Method, Status: domain.PaymentCompleted},
	}, nil
}

func (s *stubVending) RefundOrder(_ context.Context, id, reason string) (*domain.Order, error) {
	s.reason = reason
	return &domain.Order{ID: id, Status: domain.OrderRefunded}, nil
}

type stubMachines struct {
	port.MachineService
	from, to time.Time
}

func (s *stubMachines) SalesReport(_ context.Context, id string, from, to time.Time) (*domain.SalesReport, error) {
	s.from, s.to = from, to
	return &domain.SalesReport{MachineID: id, From: from, To: to, Orders: 2, RevenueCents: 300,
		Products: []domain.ProductSales{{ProductID: "p", Units: 2, RevenueCents: 300}}}, nil
}

type stubHealth struct{ err error }

func (s stubHealth) Ping(context.Context) error { return s.err }

func (s *stubVending) ReportDispense(_ context.Context, id string, dispensed bool) (*domain.Order, error) {
	s.dispense = &dispensed
	return &domain.Order{ID: id, Status: domain.OrderCompleted}, nil
}

func do(t *testing.T, h *Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	NewRouter(h).ServeHTTP(w, req)
	return w
}

func TestCheckout(t *testing.T) {
	v := &stubVending{}
	w := do(t, &Handler{vending: v}, http.MethodPost, "/api/v1/reservations/r1/checkout",
		`{"payment_method":"wallet","wallet_id":"w1","coupon_code":"SAVE"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	var resp checkoutResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	want := port.CheckoutInput{Method: domain.PaymentWallet, WalletID: "w1", CouponCode: "SAVE"}
	if v.input != want {
		t.Fatalf("input %+v", v.input)
	}
	if resp.Order.OrderNo != 7 || resp.Payment.DiscountCents != 30 || resp.Payment.CouponCode != "SAVE" {
		t.Fatalf("response %+v", resp)
	}
}

func TestCheckoutRejectsUnknownMethod(t *testing.T) {
	w := do(t, &Handler{vending: &stubVending{}}, http.MethodPost, "/api/v1/reservations/r1/checkout", `{"payment_method":"iou"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestDispenseRequiresFlag(t *testing.T) {
	v := &stubVending{}
	h := &Handler{vending: v}
	if w := do(t, h, http.MethodPost, "/api/v1/orders/o1/dispense", `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
	if w := do(t, h, http.MethodPost, "/api/v1/orders/o1/dispense", `{"dispensed":false}`); w.Code != http.StatusOK || v.dispense == nil || *v.dispense {
		t.Fatalf("status %d dispense=%v", w.Code, v.dispense)
	}
}

func TestErrorStatus(t *testing.T) {
	cases := map[error]int{
		domain.ErrNotFound:                         http.StatusNotFound,
		domain.ErrInvalidInput:                     http.StatusBadRequest,
		domain.ErrOutOfStock:                       http.StatusConflict,
		domain.ErrReservationClosed:                http.StatusConflict,
		errors.Join(domain.ErrPaymentFailed):       http.StatusPaymentRequired,
		errors.New("connection reset by postgres"): http.StatusInternalServerError,
	}
	for err, want := range cases {
		w := do(t, &Handler{vending: &stubVending{err: err}}, http.MethodPost, "/api/v1/reservations/r1/checkout", `{"payment_method":"cash"}`)
		if w.Code != want {
			t.Errorf("%v: status %d, want %d", err, w.Code, want)
		}
		if want == http.StatusInternalServerError && strings.Contains(w.Body.String(), "postgres") {
			t.Errorf("internal error leaked: %s", w.Body)
		}
	}
}

func TestHealth(t *testing.T) {
	if w := do(t, &Handler{}, http.MethodGet, "/health", ""); w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if w := do(t, &Handler{health: stubHealth{}}, http.MethodGet, "/health/ready", ""); w.Code != http.StatusOK {
		t.Fatalf("ready status %d", w.Code)
	}
	if w := do(t, &Handler{health: stubHealth{errors.New("db down")}}, http.MethodGet, "/health/ready", ""); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("not ready status %d", w.Code)
	}
}

func TestRefundOrder(t *testing.T) {
	v := &stubVending{}
	w := do(t, &Handler{vending: v}, http.MethodPost, "/api/v1/orders/o1/refund", `{"reason":"damaged"}`)
	if w.Code != http.StatusOK || v.reason != "damaged" {
		t.Fatalf("status %d reason %q", w.Code, v.reason)
	}
}

func TestSalesReportQuery(t *testing.T) {
	m := &stubMachines{}
	h := &Handler{machines: m}
	w := do(t, h, http.MethodGet, "/api/v1/machines/m1/sales?from=2026-01-01&to=2026-02-01T00:00:00Z", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	if !m.from.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) || !m.to.Equal(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("range %v - %v", m.from, m.to)
	}
	var resp salesReportResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.RevenueCents != 300 || len(resp.Products) != 1 {
		t.Fatalf("response %+v", resp)
	}
	if w := do(t, h, http.MethodGet, "/api/v1/machines/m1/sales?from=yesterday", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("bad date status %d", w.Code)
	}
	if w := do(t, h, http.MethodGet, "/api/v1/machines/m1/sales", ""); w.Code != http.StatusOK || m.to.Sub(m.from) != 30*24*time.Hour {
		t.Fatalf("default range %v", m.to.Sub(m.from))
	}
}

type stubCatalog struct {
	port.CatalogService
	page domain.PageRequest
}

func (s *stubCatalog) ListProducts(_ context.Context, _ string, page domain.PageRequest) (domain.Page[domain.Product], error) {
	s.page = page
	return domain.Page[domain.Product]{Items: []domain.Product{{ID: "p1"}}, NextCursor: "abc"}, nil
}

func TestListProductsPaging(t *testing.T) {
	cat := &stubCatalog{}
	h := &Handler{catalog: cat}
	w := do(t, h, http.MethodGet, "/api/v1/products?limit=5&cursor=xyz", "")
	if w.Code != http.StatusOK || cat.page.Limit != 5 || cat.page.Cursor != "xyz" {
		t.Fatalf("status %d page %+v", w.Code, cat.page)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["next_cursor"] != "abc" || resp["is_last_page"] != false || resp["limit"] != float64(5) || len(resp["items"].([]any)) != 1 {
		t.Fatalf("response %v", resp)
	}
	for _, bad := range []string{"0", "101", "ten"} {
		if w := do(t, h, http.MethodGet, "/api/v1/products?limit="+bad, ""); w.Code != http.StatusBadRequest {
			t.Errorf("limit=%s: status %d", bad, w.Code)
		}
	}
}

func TestPendingPaymentAnswers202(t *testing.T) {
	v := &stubVending{err: &domain.PendingPaymentError{PaymentID: "pay-1"}}
	w := do(t, &Handler{vending: v}, http.MethodPost, "/api/v1/reservations/r1/checkout", `{"payment_method":"cash"}`)
	if w.Code != http.StatusAccepted || !strings.Contains(w.Body.String(), `"payment_id":"pay-1"`) {
		t.Fatalf("status %d body %s", w.Code, w.Body)
	}
}
