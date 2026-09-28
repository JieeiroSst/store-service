package handler

import (
	"github.com/gin-gonic/gin"
)

// mustGetUserID returns the caller's user-service id set by the auth
// middleware (0 if the route is unauthenticated).
func mustGetUserID(c *gin.Context) int64 {
	raw, _ := c.Get("user_id")
	id, _ := raw.(int64)
	return id
}
