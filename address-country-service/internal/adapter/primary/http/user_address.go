package http

import (
	"strconv"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateUserAddress(c *gofr.Context) (interface{}, error) {
	var req model.UserAddress
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.userAddress.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetUserAddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.userAddress.Get(c, id)
}

func (h *Handler) UpdateUserAddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.UserAddress
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.ID = id

	if err := h.userAddress.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteUserAddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.userAddress.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListUserAddresses(c *gofr.Context) (interface{}, error) {
	return h.userAddress.List(c)
}
