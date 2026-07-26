package http

import (
	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"gofr.dev/pkg/gofr"
)

func (h *Handler) TriggerDagRun(c *gofr.Context) (interface{}, error) {
	dagId := c.PathParam("dagId")

	var req model.TriggerDAGRunRequest
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	return h.dagRun.Trigger(c, dagId, &req)
}

func (h *Handler) ListDagRuns(c *gofr.Context) (interface{}, error) {
	dagId := c.PathParam("dagId")
	limit, offset := pagination(c)

	return h.dagRun.List(c, dagId, limit, offset)
}

func (h *Handler) GetDagRun(c *gofr.Context) (interface{}, error) {
	dagId := c.PathParam("dagId")
	dagRunId := c.PathParam("dagRunId")

	return h.dagRun.Get(c, dagId, dagRunId)
}

func (h *Handler) DeleteDagRun(c *gofr.Context) (interface{}, error) {
	dagId := c.PathParam("dagId")
	dagRunId := c.PathParam("dagRunId")

	if err := h.dagRun.Delete(c, dagId, dagRunId); err != nil {
		return nil, err
	}

	return nil, nil
}
