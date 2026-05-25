package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func mustGetUserID(c *gin.Context) uuid.UUID {
	raw, _ := c.Get("user_id")
	if id, ok := raw.(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}
