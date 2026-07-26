package http

import (
	"strconv"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateAddress(c *gofr.Context) (interface{}, error) {
	var req model.Address
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.address.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetAddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.address.Get(c, id)
}

func (h *Handler) UpdateAddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.Address
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.AddressID = id

	if err := h.address.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteAddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.address.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListAddresses(c *gofr.Context) (interface{}, error) {
	return h.address.List(c)
}
