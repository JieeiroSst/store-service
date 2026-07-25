package http

import (
	"strconv"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateOrderBillingaddress(c *gofr.Context) (interface{}, error) {
	var req model.OrderBillingaddress
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.orderBillingaddress.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetOrderBillingaddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.orderBillingaddress.Get(c, id)
}

func (h *Handler) UpdateOrderBillingaddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.OrderBillingaddress
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.ID = id

	if err := h.orderBillingaddress.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteOrderBillingaddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.orderBillingaddress.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListOrderBillingaddresses(c *gofr.Context) (interface{}, error) {
	return h.orderBillingaddress.List(c)
}
