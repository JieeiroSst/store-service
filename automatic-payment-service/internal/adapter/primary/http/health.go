package http

import "gofr.dev/pkg/gofr"

func (h *Handler) GetHealth(c *gofr.Context) (interface{}, error) {
	return map[string]string{"status": "ok"}, nil
}
