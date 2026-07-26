package http

import (
	"strconv"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateCustomer(c *gofr.Context) (interface{}, error) {
	var req model.Customer
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.customer.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetCustomer(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.customer.Get(c, id)
}

func (h *Handler) UpdateCustomer(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.Customer
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.CustomerID = id

	if err := h.customer.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteCustomer(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.customer.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListCustomers(c *gofr.Context) (interface{}, error) {
	return h.customer.List(c)
}
