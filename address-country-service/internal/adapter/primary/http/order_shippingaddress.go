package http

import (
	"strconv"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateOrderShippingaddress(c *gofr.Context) (interface{}, error) {
	var req model.OrderShippingaddress
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.orderShippingaddress.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetOrderShippingaddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.orderShippingaddress.Get(c, id)
}

func (h *Handler) UpdateOrderShippingaddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.OrderShippingaddress
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.ID = id

	if err := h.orderShippingaddress.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteOrderShippingaddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.orderShippingaddress.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListOrderShippingaddresses(c *gofr.Context) (interface{}, error) {
	return h.orderShippingaddress.List(c)
}
