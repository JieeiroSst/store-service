package http

import (
	"strconv"

	"github.com/JIeeiroSst/address-country-service/pkg/authn"
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

func principalUserID(p authn.Principal) (int, error) {
	id, err := strconv.Atoi(p.UserID)
	if err != nil {
		return 0, authn.ErrForbidden
	}
	return id, nil
}
