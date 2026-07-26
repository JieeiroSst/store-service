package http

import (
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/google/uuid"
	"gofr.dev/pkg/gofr"
	gofrhttp "gofr.dev/pkg/gofr/http"
)

type addPaymentMethodRequest struct {
	UserID         uuid.UUID `json:"userId"`
	Provider       string    `json:"provider"`
	TokenID        string    `json:"tokenId"`
	LastFourDigits string    `json:"lastFourDigits"`
	ExpiryDate     string    `json:"expiryDate"`
	IsDefault      bool      `json:"isDefault"`
}

func (h *Handler) AddPaymentMethod(c *gofr.Context) (interface{}, error) {
	var req addPaymentMethodRequest
	if err := c.Bind(&req); err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"body"}}
	}

	pm, err := h.paymentMethod.AddPaymentMethod(c, port.AddPaymentMethodRequest{
		UserID:         req.UserID,
		Provider:       req.Provider,
		TokenID:        req.TokenID,
		LastFourDigits: req.LastFourDigits,
		ExpiryDate:     req.ExpiryDate,
		IsDefault:      req.IsDefault,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return pm, nil
}

func (h *Handler) ListPaymentMethods(c *gofr.Context) (interface{}, error) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"userId"}}
	}

	pms, err := h.paymentMethod.ListPaymentMethods(c, userID)
	if err != nil {
		return nil, mapError(err)
	}
	return pms, nil
}

func (h *Handler) DeletePaymentMethod(c *gofr.Context) (interface{}, error) {
	id, err := uuid.Parse(c.PathParam("id"))
	if err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"id"}}
	}

	if err := h.paymentMethod.DeletePaymentMethod(c, id); err != nil {
		return nil, mapError(err)
	}
	return map[string]bool{"success": true}, nil
}

type setDefaultPaymentMethodRequest struct {
	UserID uuid.UUID `json:"userId"`
}

func (h *Handler) SetDefaultPaymentMethod(c *gofr.Context) (interface{}, error) {
	id, err := uuid.Parse(c.PathParam("id"))
	if err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"id"}}
	}

	var req setDefaultPaymentMethodRequest
	if err := c.Bind(&req); err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"body"}}
	}

	if err := h.paymentMethod.SetDefaultPaymentMethod(c, req.UserID, id); err != nil {
		return nil, mapError(err)
	}
	return map[string]bool{"success": true}, nil
}
