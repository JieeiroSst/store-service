package http

import (
	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreateVariable(c *gofr.Context) (interface{}, error) {
	var req model.Variable
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	return h.variable.Create(c, &req)
}

func (h *Handler) ListVariables(c *gofr.Context) (interface{}, error) {
	limit, offset := pagination(c)
	return h.variable.List(c, limit, offset)
}

func (h *Handler) GetVariable(c *gofr.Context) (interface{}, error) {
	key := c.PathParam("key")
	return h.variable.Get(c, key)
}

func (h *Handler) UpdateVariable(c *gofr.Context) (interface{}, error) {
	key := c.PathParam("key")

	var req model.Variable
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.Key = key

	return h.variable.Update(c, &req)
}

func (h *Handler) DeleteVariable(c *gofr.Context) (interface{}, error) {
	key := c.PathParam("key")

	if err := h.variable.Delete(c, key); err != nil {
		return nil, err
	}

	return nil, nil
}
