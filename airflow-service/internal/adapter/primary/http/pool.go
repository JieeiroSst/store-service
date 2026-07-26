package http

import (
	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) CreatePool(c *gofr.Context) (interface{}, error) {
	var req model.Pool
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	return h.pool.Create(c, &req)
}

func (h *Handler) ListPools(c *gofr.Context) (interface{}, error) {
	limit, offset := pagination(c)
	return h.pool.List(c, limit, offset)
}

func (h *Handler) GetPool(c *gofr.Context) (interface{}, error) {
	name := c.PathParam("name")
	return h.pool.Get(c, name)
}

func (h *Handler) UpdatePool(c *gofr.Context) (interface{}, error) {
	name := c.PathParam("name")

	var req model.Pool
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	req.Name = name

	return h.pool.Update(c, &req)
}

func (h *Handler) DeletePool(c *gofr.Context) (interface{}, error) {
	name := c.PathParam("name")

	if err := h.pool.Delete(c, name); err != nil {
		return nil, err
	}

	return nil, nil
}
