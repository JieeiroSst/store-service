package v1

import "github.com/gin-gonic/gin"

func (h *Handler) initUser(api *gin.RouterGroup) {
	_ = api.Group("/user")

}
