package http

import (
	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateAddressCountry(c *gofr.Context) (interface{}, error) {
	var req model.AddressCountry
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.addressCountry.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetAddressCountry(c *gofr.Context) (interface{}, error) {
	id := c.PathParam("id")

	return h.addressCountry.Get(c, id)
}

func (h *Handler) UpdateAddressCountry(c *gofr.Context) (interface{}, error) {
	id := c.PathParam("id")

	var req model.AddressCountry
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.Id = id

	if err := h.addressCountry.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteAddressCountry(c *gofr.Context) (interface{}, error) {
	id := c.PathParam("id")

	if err := h.addressCountry.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListAddressCountries(c *gofr.Context) (interface{}, error) {
	return h.addressCountry.List(c)
}
