package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func respondValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "Invalid request",
		"message": err.Error(),
	})
}

func respondNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}

func respondInternalError(c *gin.Context, publicMessage string, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   publicMessage,
		"message": err.Error(),
	})
}
