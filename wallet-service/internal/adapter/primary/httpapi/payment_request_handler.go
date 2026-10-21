package httpapi

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

func (a *API) createPaymentRequest(c *gin.Context) {
	var req struct {
		UserID      string          `json:"user_id"`
		WalletID    string          `json:"wallet_id"`
		PayerUserID string          `json:"payer_user_id"`
		Amount      decimal.Decimal `json:"amount"`
		Note        string          `json:"note"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	cmd := port.CreatePaymentRequestCommand{
		RequesterUserID:   p.required("user_id", req.UserID),
		RequesterWalletID: p.required("wallet_id", req.WalletID),
		PayerUserID:       p.required("payer_user_id", req.PayerUserID),
		Amount:            req.Amount,
		Note:              req.Note,
	}
	if !p.ok(c) {
		return
	}
	pr, err := a.PaymentRequests.Create(c.Request.Context(), cmd)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, pr)
}

func (a *API) getPaymentRequest(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	pr, err := a.PaymentRequests.Get(c.Request.Context(), id)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, pr)
}

func (a *API) listPaymentRequests(c *gin.Context) {
	userID, ok := pathID(c, "user_id")
	if !ok {
		return
	}
	pg, size := pagination(c)
	items, total, err := a.PaymentRequests.List(c.Request.Context(), model.PaymentRequestFilter{
		UserID:   userID,
		Role:     model.PaymentRequestRole(c.Query("role")),
		Status:   model.PaymentRequestStatus(c.Query("status")),
		Page:     pg,
		PageSize: size,
	})
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, page[*model.PaymentRequest]{Items: items, Total: total, Page: pg, PageSize: size})
}

func (a *API) payPaymentRequest(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req struct {
		UserID   string `json:"user_id"`
		WalletID string `json:"wallet_id"`
		Pin      string `json:"pin"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	cmd := port.PayPaymentRequestCommand{
		RequestID: id,
		UserID:    p.required("user_id", req.UserID),
		WalletID:  p.required("wallet_id", req.WalletID),
		Pin:       req.Pin,
	}
	if !p.ok(c) {
		return
	}
	pr, err := a.PaymentRequests.Pay(c.Request.Context(), cmd)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, pr)
}

func (a *API) declinePaymentRequest(c *gin.Context) {
	a.paymentRequestAction(c, a.PaymentRequests.Decline)
}

func (a *API) cancelPaymentRequest(c *gin.Context) {
	a.paymentRequestAction(c, a.PaymentRequests.Cancel)
}

func (a *API) paymentRequestAction(c *gin.Context, fn func(context.Context, uuid.UUID, uuid.UUID) (*model.PaymentRequest, error)) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req struct {
		UserID string `json:"user_id"`
	}
	if !bind(c, &req) {
		return
	}
	var p ids
	userID := p.required("user_id", req.UserID)
	if !p.ok(c) {
		return
	}
	pr, err := fn(c.Request.Context(), id, userID)
	if err != nil {
		a.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, pr)
}
