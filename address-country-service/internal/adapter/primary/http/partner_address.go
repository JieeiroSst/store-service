package http

import (
	"strconv"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreatePartneraddress(c *gofr.Context) (interface{}, error) {
	var req model.Partneraddress
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.partneraddress.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetPartneraddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.partneraddress.Get(c, id)
}

func (h *Handler) UpdatePartneraddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.Partneraddress
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.ID = id

	if err := h.partneraddress.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeletePartneraddress(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.partneraddress.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListPartneraddresses(c *gofr.Context) (interface{}, error) {
	return h.partneraddress.List(c)
}
