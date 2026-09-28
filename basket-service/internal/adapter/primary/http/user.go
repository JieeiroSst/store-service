package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if !isOwner(c, id) {
		writeError(c, common.ErrNotFound)
		return
	}
	result, err := h.user.GetUser(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
