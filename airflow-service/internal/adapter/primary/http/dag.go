package http

import "gofr.dev/pkg/gofr"

func (h *Handler) ListDags(c *gofr.Context) (interface{}, error) {
	limit, offset := pagination(c)
	return h.dag.List(c, limit, offset)
}

func (h *Handler) GetDag(c *gofr.Context) (interface{}, error) {
	dagId := c.PathParam("dagId")
	return h.dag.Get(c, dagId)
}

type setPausedRequest struct {
	IsPaused bool `json:"is_paused"`
}

func (h *Handler) SetDagPaused(c *gofr.Context) (interface{}, error) {
	dagId := c.PathParam("dagId")

	var req setPausedRequest
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	return h.dag.SetPaused(c, dagId, req.IsPaused)
}
