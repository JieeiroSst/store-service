package http

import (
	"net/http"

	"github.com/JIeeiroSst/admanagement-service/pkg/authn"
)

func admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := authn.RequireAdmin(r.Context()); err != nil {
			authn.WriteError(w, err)
			return
		}
		next(w, r)
	}
}
