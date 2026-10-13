package http

import (
	"github.com/JIeeiroSst/billing-service/pkg/authn"
	"gofr.dev/pkg/gofr"
)

func admin(next gofr.Handler) gofr.Handler {
	return func(c *gofr.Context) (interface{}, error) {
		if _, err := authn.RequireAdmin(c); err != nil {
			return nil, err
		}
		return next(c)
	}
}
