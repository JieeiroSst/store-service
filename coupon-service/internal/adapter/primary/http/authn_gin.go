package http

import (
	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/coupon-service/pkg/authn"
)

func authenticate(a *authn.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, ok, err := a.Optional(c.Request)
		if err != nil {
			abortAuth(c, err)
			return
		}
		if ok {
			c.Request = c.Request.WithContext(authn.WithPrincipal(c.Request.Context(), p))
		}
		c.Next()
	}
}

func requireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := authn.Require(c.Request.Context()); err != nil {
			abortAuth(c, err)
			return
		}
		c.Next()
	}
}

func requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := authn.RequireAdmin(c.Request.Context()); err != nil {
			abortAuth(c, err)
			return
		}
		c.Next()
	}
}

func principal(c *gin.Context) authn.Principal {
	p, _ := authn.FromContext(c.Request.Context())
	return p
}

func abortAuth(c *gin.Context, err error) {
	c.AbortWithStatusJSON(authn.Status(err), gin.H{"error": authn.Message(err)})
}
