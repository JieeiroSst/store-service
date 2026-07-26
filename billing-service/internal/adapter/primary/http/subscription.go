package http

import (
	"strconv"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateSubscription(c *gofr.Context) (interface{}, error) {
	var req model.Subscription
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	if err := h.subscription.Create(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) GetSubscription(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	return h.subscription.Get(c, id)
}

func (h *Handler) UpdateSubscription(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var req model.Subscription
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.SubscriptionID = id

	if err := h.subscription.Update(c, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Handler) DeleteSubscription(c *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(c.PathParam("id"))
	if err != nil {
		return nil, err
	}

	if err := h.subscription.Delete(c, id); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) ListSubscriptions(c *gofr.Context) (interface{}, error) {
	return h.subscription.List(c)
}
