package v1

import "github.com/gin-gonic/gin"

func (h *Handler) initTickets(api *gin.RouterGroup) {
	_ = api.Group("/tickets")

}
