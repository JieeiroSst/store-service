package http

import (
	"strconv"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateInvoice(c *gofr.Context) (interface{}, error) {
	var req model.Invoice
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.invoice.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetInvoice(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.invoice.Get(c, id)
}

func (h *Handler) UpdateInvoice(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.Invoice
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.InvoiceID = id

	if err := h.invoice.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteInvoice(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.invoice.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListInvoices(c *gofr.Context) (interface{}, error) {
	return h.invoice.List(c)
}
