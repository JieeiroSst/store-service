package http

import (
	"strconv"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateShippingOrderanditemchangesCountry(c *gofr.Context) (interface{}, error) {
	var req model.ShippingOrderanditemchangesCountry
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.shippingOrderanditemchangesCountry.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetShippingOrderanditemchangesCountry(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.shippingOrderanditemchangesCountry.Get(c, id)
}

func (h *Handler) UpdateShippingOrderanditemchangesCountry(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.ShippingOrderanditemchangesCountry
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.ID = id

	if err := h.shippingOrderanditemchangesCountry.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteShippingOrderanditemchangesCountry(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.shippingOrderanditemchangesCountry.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListShippingOrderanditemchangesCountries(c *gofr.Context) (interface{}, error) {
	return h.shippingOrderanditemchangesCountry.List(c)
}
