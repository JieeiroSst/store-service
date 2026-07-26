package http

import (
	"strconv"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateTransaction(c *gofr.Context) (interface{}, error) {
	var req model.Transaction
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.transaction.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetTransaction(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.transaction.Get(c, id)
}

func (h *Handler) UpdateTransaction(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.Transaction
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.TransactionID = id

	if err := h.transaction.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteTransaction(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.transaction.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListTransactions(c *gofr.Context) (interface{}, error) {
	return h.transaction.List(c)
}
