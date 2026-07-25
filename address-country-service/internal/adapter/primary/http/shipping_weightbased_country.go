package http

import (
	"strconv"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateShippingWeightbasedCountry(c *gofr.Context) (interface{}, error) {
	var req model.ShippingWeightbasedCountry
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.shippingWeightbasedCountry.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetShippingWeightbasedCountry(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.shippingWeightbasedCountry.Get(c, id)
}

func (h *Handler) UpdateShippingWeightbasedCountry(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.ShippingWeightbasedCountry
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.ID = id

	if err := h.shippingWeightbasedCountry.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteShippingWeightbasedCountry(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.shippingWeightbasedCountry.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListShippingWeightbasedCountries(c *gofr.Context) (interface{}, error) {
	return h.shippingWeightbasedCountry.List(c)
}
