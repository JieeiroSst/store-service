package http

import (
	"github.com/JIeeiroSst/gateway-service/pkg/authn"
	"github.com/gin-gonic/gin"
)

func Authenticate(a *authn.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Header.Del("X-User-Id")
		p, err := a.Request(c.Request)
		if err != nil {
			authn.WriteError(c.Writer, err)
			c.Abort()
			return
		}
		if p.UserID != "" {
			c.Request.Header.Set("X-User-Id", p.UserID)
		}
		c.Request = c.Request.WithContext(authn.WithPrincipal(c.Request.Context(), p))
		c.Next()
	}
}
