package http

import "gofr.dev/pkg/gofr"

func (h *Handler) GetHealth(c *gofr.Context) (interface{}, error) {
	return h.health.Get(c)
}
