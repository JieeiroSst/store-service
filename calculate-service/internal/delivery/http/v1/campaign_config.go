package v1

import "github.com/gin-gonic/gin"

func (h *Handler) initCampaignConfig(api *gin.RouterGroup) {
	_ = api.Group("/campaign-config")

}
