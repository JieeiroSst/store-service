package billing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

// fakeBilling mimics billing-service: gofr envelopes and full-record PUTs.
type fakeBilling struct {
	mu       sync.Mutex
	plans    []map[string]any
	requests []string
	invoices map[int]map[string]any
	put      map[string]any
}

func (f *fakeBilling) handler(t *testing.T) http.Handler {
	reply := func(w http.ResponseWriter, v any) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": v})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.requests = append(f.requests, r.Method+" "+r.URL.Path)
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)

		switch r.Method + " " + r.URL.Path {
		case "GET /plans":
			reply(w, f.plans)
		case "POST /plans":
			body["plan_id"] = 7
			f.plans = append(f.plans, body)
			reply(w, body)
		case "POST /customers":
			reply(w, map[string]any{"customer_id": 11})
		case "POST /subscriptions":
			reply(w, map[string]any{"subscription_id": 22})
		case "POST /invoices":
			body["invoice_id"] = 33
			f.invoices[33] = body
			reply(w, body)
		case "GET /invoices/33":
			reply(w, f.invoices[33])
		case "PUT /invoices/33":
			f.put = body
			reply(w, body)
		case "POST /transactions":
			reply(w, body)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	})
}

func newClient(t *testing.T) (*Client, *fakeBilling) {
	f := &fakeBilling{invoices: map[int]map[string]any{}}
	srv := httptest.NewServer(f.handler(t))
	t.Cleanup(srv.Close)
	cfg := &config.Config{Billing: config.BillingConfig{
		UpstreamConfig: config.UpstreamConfig{BaseURL: srv.URL, Timeout: 2 * time.Second},
		PlanName:       "Hospital services",
	}}
	return New(cfg), f
}

func TestEnsureAccountCreatesCustomerAndSubscriptionAndReusesThePlan(t *testing.T) {
	c, f := newClient(t)
	ctx := context.Background()
	p := &model.Patient{ID: 5, FirstName: "Binh", LastName: "Tran"}

	acct, err := c.EnsureAccount(ctx, p, nil)
	if err != nil {
		t.Fatal(err)
	}
	if acct.PatientID != 5 || acct.CustomerID != 11 || acct.SubscriptionID != 22 {
		t.Errorf("account = %+v", acct)
	}
	if _, err := c.EnsureAccount(ctx, &model.Patient{ID: 6}, nil); err != nil {
		t.Fatal(err)
	}
	creates := 0
	for _, r := range f.requests {
		if r == "POST /plans" {
			creates++
		}
	}
	if creates != 1 {
		t.Errorf("plan created %d times, want once", creates)
	}

	existing := &model.BillingAccount{PatientID: 5, CustomerID: 1, SubscriptionID: 2}
	if got, _ := c.EnsureAccount(ctx, p, existing); got != existing {
		t.Errorf("an existing account must be returned untouched")
	}
}

func TestEnsureAccountFindsAnExistingPlanByName(t *testing.T) {
	c, f := newClient(t)
	f.plans = []map[string]any{{"plan_id": 3, "name": "Other"}, {"plan_id": 9, "name": "Hospital services"}}
	if _, err := c.EnsureAccount(context.Background(), &model.Patient{ID: 1}, nil); err != nil {
		t.Fatal(err)
	}
	for _, r := range f.requests {
		if r == "POST /plans" {
			t.Error("plan created although one with that name exists")
		}
	}
}

func TestInvoiceLifecycle(t *testing.T) {
	c, f := newClient(t)
	ctx := context.Background()
	acct := &model.BillingAccount{PatientID: 5, CustomerID: 11, SubscriptionID: 22}
	due := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	id, err := c.CreateInvoice(ctx, acct, port.InvoiceState{Amount: 60.5, DueDate: &due, Status: port.InvoiceIssued})
	if err != nil || id != 33 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	in := f.invoices[33]
	if in["subscription_id"] != float64(22) || in["amount"] != 60.5 || in["status"] != "issued" || in["due_date"] != float64(due.Unix()) {
		t.Errorf("invoice sent = %v", in)
	}

	if err := c.SyncInvoice(ctx, 33, port.InvoiceState{Amount: 60.5, Status: port.InvoicePaid}); err != nil {
		t.Fatal(err)
	}
	// PUT replaces the record, so the untouched fields must be sent back.
	if f.put["status"] != "paid" || f.put["subscription_id"] != float64(22) || f.put["due_date"] != float64(due.Unix()) || f.put["invoice_id"] != float64(33) {
		t.Errorf("PUT body = %v", f.put)
	}

	if err := c.RecordPayment(ctx, 33, 60.5, "wallet", time.Now()); err != nil {
		t.Errorf("record payment: %v", err)
	}
}

func TestOutageIsAnUpstreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":{"message":"db down"}}`, http.StatusInternalServerError)
	}))
	defer srv.Close()
	c := New(&config.Config{Billing: config.BillingConfig{UpstreamConfig: config.UpstreamConfig{BaseURL: srv.URL, Timeout: time.Second}}})

	_, err := c.CreateInvoice(context.Background(), &model.BillingAccount{}, port.InvoiceState{})
	if err == nil || !isUpstream(err) {
		t.Errorf("err = %v, want ErrUpstream", err)
	}
	if New(&config.Config{}).Enabled() {
		t.Error("client with no base URL must be disabled")
	}
}

func isUpstream(err error) bool { return errors.Is(err, model.ErrUpstream) }
