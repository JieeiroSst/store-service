package middleware

import (
	"net/http"

	gofrHTTP "gofr.dev/pkg/gofr/http"
)

func CustomMiddleware() gofrHTTP.Middleware {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Your custom logic here
			// For example, logging, authentication, etc.

			// Call the next handler in the chain
			inner.ServeHTTP(w, r)
		})
	}
}

//  app.UseMiddleware(customMiddleware())
