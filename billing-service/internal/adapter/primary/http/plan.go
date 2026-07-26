package http

import (
	"strconv"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreatePlan(c *gofr.Context) (interface{}, error) {
	var req model.Plan
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.plan.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetPlan(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.plan.Get(c, id)
}

func (h *Handler) UpdatePlan(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.Plan
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.PlanID = id

	if err := h.plan.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeletePlan(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.plan.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListPlans(c *gofr.Context) (interface{}, error) {
	return h.plan.List(c)
}
