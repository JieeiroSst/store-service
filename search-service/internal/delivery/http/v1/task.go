package v1

import "github.com/gin-gonic/gin"

func (h *Handler) initSearchTaskRoutes(api *gin.RouterGroup) {
	group := api.Group("/task")

	group.POST("/search", h.middleware.AuthorizeControl(), h.SearchTask)
}

func (h *Handler) SearchTask(c *gin.Context) {

}
