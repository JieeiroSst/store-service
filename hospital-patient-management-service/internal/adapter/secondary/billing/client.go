package billing

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/httpx"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"go.uber.org/fx"
)

// Client talks to billing-service, a gofr app: every answer is wrapped as
// {"data": ...} and timestamps are unix seconds.
type Client struct {
	http     *httpx.Client
	planName string

	mu     sync.Mutex
	planID int
}

func New(cfg *config.Config) *Client {
	return &Client{
		http:     httpx.New(cfg.Billing.BaseURL, cfg.Billing.Timeout),
		planName: cfg.Billing.PlanName,
	}
}

func (c *Client) Enabled() bool { return c.http.Enabled() }

type envelope[T any] struct {
	Data T `json:"data"`
}

type invoiceDTO struct {
	InvoiceID      int     `json:"invoice_id,omitempty"`
	SubscriptionID int     `json:"subscription_id"`
	InvoiceDate    int64   `json:"invoice_date"`
	DueDate        int64   `json:"due_date"`
	Amount         float64 `json:"amount"`
	Tax            string  `json:"tax,omitempty"`
	Status         string  `json:"status"`
}

func (c *Client) EnsureAccount(ctx context.Context, p *model.Patient, existing *model.BillingAccount) (*model.BillingAccount, error) {
	if existing != nil {
		return existing, nil
	}
	planID, err := c.ensurePlan(ctx)
	if err != nil {
		return nil, err
	}

	var cust envelope[struct {
		CustomerID int `json:"customer_id"`
	}]
	err = c.http.Do(ctx, http.MethodPost, "/customers", map[string]any{
		"name":         strings.TrimSpace(p.FirstName + " " + p.LastName),
		"email":        p.Email,
		"phone_number": p.Phone,
	}, &cust)
	if err != nil {
		return nil, wrap("create customer", err)
	}

	var sub envelope[struct {
		SubscriptionID int `json:"subscription_id"`
	}]
	err = c.http.Do(ctx, http.MethodPost, "/subscriptions", map[string]any{
		"customer_id": cust.Data.CustomerID,
		"plan_id":     planID,
		"start_date":  time.Now().Unix(),
		"status":      "active",
	}, &sub)
	if err != nil {
		return nil, wrap("create subscription", err)
	}
	return &model.BillingAccount{
		PatientID:      p.ID,
		CustomerID:     int32(cust.Data.CustomerID),
		SubscriptionID: int32(sub.Data.SubscriptionID),
	}, nil
}

// ensurePlan finds (or creates once) the plan all hospital invoices use.
func (c *Client) ensurePlan(ctx context.Context) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.planID != 0 {
		return c.planID, nil
	}

	var plans envelope[[]struct {
		PlanID int    `json:"plan_id"`
		Name   string `json:"name"`
	}]
	if err := c.http.Do(ctx, http.MethodGet, "/plans", nil, &plans); err != nil {
		return 0, wrap("list plans", err)
	}
	for _, p := range plans.Data {
		if p.Name == c.planName {
			c.planID = p.PlanID
			return c.planID, nil
		}
	}

	var created envelope[struct {
		PlanID int `json:"plan_id"`
	}]
	err := c.http.Do(ctx, http.MethodPost, "/plans", map[string]any{
		"name":          c.planName,
		"description":   "Charges from the hospital patient management service",
		"price":         0,
		"billing_cycle": "one_time",
	}, &created)
	if err != nil {
		return 0, wrap("create plan", err)
	}
	c.planID = created.Data.PlanID
	return c.planID, nil
}

func (c *Client) CreateInvoice(ctx context.Context, acct *model.BillingAccount, st port.InvoiceState) (int32, error) {
	var out envelope[invoiceDTO]
	err := c.http.Do(ctx, http.MethodPost, "/invoices", invoiceDTO{
		SubscriptionID: int(acct.SubscriptionID),
		InvoiceDate:    time.Now().Unix(),
		DueDate:        unix(st.DueDate),
		Amount:         st.Amount,
		Status:         string(st.Status),
	}, &out)
	if err != nil {
		return 0, wrap("create invoice", err)
	}
	if out.Data.InvoiceID == 0 {
		return 0, fmt.Errorf("%w: billing-service returned no invoice id", model.ErrUpstream)
	}
	return int32(out.Data.InvoiceID), nil
}

// SyncInvoice reads the invoice first because billing-service's PUT replaces
// the whole record: sending only the changed fields would zero the rest.
func (c *Client) SyncInvoice(ctx context.Context, invoiceID int32, st port.InvoiceState) error {
	path := "/invoices/" + strconv.Itoa(int(invoiceID))
	var cur envelope[invoiceDTO]
	if err := c.http.Do(ctx, http.MethodGet, path, nil, &cur); err != nil {
		return wrap("get invoice", err)
	}
	inv := cur.Data
	inv.InvoiceID = int(invoiceID)
	inv.Amount = st.Amount
	inv.Status = string(st.Status)
	if st.DueDate != nil {
		inv.DueDate = st.DueDate.Unix()
	}
	if err := c.http.Do(ctx, http.MethodPut, path, inv, nil); err != nil {
		return wrap("update invoice", err)
	}
	return nil
}

func (c *Client) RecordPayment(ctx context.Context, invoiceID int32, amount float64, method string, at time.Time) error {
	err := c.http.Do(ctx, http.MethodPost, "/transactions", map[string]any{
		"invoice_id":       invoiceID,
		"payment_method":   method,
		"transaction_date": at.Unix(),
		"amount":           amount,
		"status":           "successful",
	}, nil)
	if err != nil {
		return wrap("record transaction", err)
	}
	return nil
}

func unix(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}

func wrap(what string, err error) error {
	return fmt.Errorf("%w: billing-service %s: %v", model.ErrUpstream, what, err)
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.InvoiceGateway)))),
)
