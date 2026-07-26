package http

import (
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/google/uuid"
	"gofr.dev/pkg/gofr"
	gofrhttp "gofr.dev/pkg/gofr/http"
)

type createSubscriptionRequest struct {
	UserID          uuid.UUID `json:"userId"`
	PlanID          uuid.UUID `json:"planId"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	AutoRenewal     bool      `json:"autoRenewal"`
	TrialDays       int       `json:"trialDays"`
	PaymentMethodID uuid.UUID `json:"paymentMethodId"`
}

func (h *Handler) CreateSubscription(c *gofr.Context) (interface{}, error) {
	var req createSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"body"}}
	}

	sub, err := h.subscription.CreateSubscription(c, port.CreateSubscriptionRequest{
		UserID:          req.UserID,
		PlanID:          req.PlanID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		AutoRenewal:     req.AutoRenewal,
		TrialDays:       req.TrialDays,
		PaymentMethodID: req.PaymentMethodID,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return sub, nil
}

func (h *Handler) GetSubscription(c *gofr.Context) (interface{}, error) {
	id, err := uuid.Parse(c.PathParam("id"))
	if err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"id"}}
	}

	sub, err := h.subscription.GetSubscription(c, id)
	if err != nil {
		return nil, mapError(err)
	}
	return sub, nil
}

func (h *Handler) CancelSubscription(c *gofr.Context) (interface{}, error) {
	id, err := uuid.Parse(c.PathParam("id"))
	if err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"id"}}
	}

	sub, err := h.subscription.CancelSubscription(c, id)
	if err != nil {
		return nil, mapError(err)
	}
	return sub, nil
}

func (h *Handler) ProcessRenewals(c *gofr.Context) (interface{}, error) {
	count, err := h.subscription.ProcessDueRenewals(c)
	if err != nil {
		return nil, mapError(err)
	}
	return map[string]int{"processed": count}, nil
}

func (h *Handler) ListSubscriptionTransactions(c *gofr.Context) (interface{}, error) {
	id, err := uuid.Parse(c.PathParam("id"))
	if err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"id"}}
	}

	txs, err := h.transaction.ListTransactionsBySubscription(c, id)
	if err != nil {
		return nil, mapError(err)
	}
	return txs, nil
}

func (h *Handler) ListSubscriptionInvoices(c *gofr.Context) (interface{}, error) {
	id, err := uuid.Parse(c.PathParam("id"))
	if err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"id"}}
	}

	invoices, err := h.invoice.ListInvoicesBySubscription(c, id)
	if err != nil {
		return nil, mapError(err)
	}
	return invoices, nil
}
