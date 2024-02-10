package v1

import (
	"net/http"

	"github.com/JIeeiroSst/search-service/internal"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initSearchTaskRoutes(api *gin.RouterGroup) {
	group := api.Group("/task")

	group.POST("/search", h.middleware.AuthorizeControl(), h.SearchTask)
}

func (h *Handler) SearchTask(c *gin.Context) {
	var args internal.SearchParams
	if err := c.ShouldBindJSON(&args); err != nil {
		Response(c, http.StatusBadRequest, Message{Message: err.Error()})
		return
	}
	tasks, err := h.service.Tasks.Search(c, args)
	if err != nil {
		Response(c, http.StatusInternalServerError, Message{Message: err.Error()})
		return
	}

	Response(c, http.StatusOK, Message{
		Message: "Success",
		Data:    tasks,
	})
}
