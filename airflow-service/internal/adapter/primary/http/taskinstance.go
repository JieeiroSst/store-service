package http

import "gofr.dev/pkg/gofr"

func (h *Handler) ListTaskInstances(c *gofr.Context) (interface{}, error) {
	dagId := c.PathParam("dagId")
	dagRunId := c.PathParam("dagRunId")

	return h.taskInstance.List(c, dagId, dagRunId)
}

func (h *Handler) GetTaskInstance(c *gofr.Context) (interface{}, error) {
	dagId := c.PathParam("dagId")
	dagRunId := c.PathParam("dagRunId")
	taskId := c.PathParam("taskId")

	return h.taskInstance.Get(c, dagId, dagRunId, taskId)
}
