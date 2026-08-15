package http

import (
	stdhttp "net/http"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(stdhttp.StatusOK, gin.H{"status": "ok"})
}
