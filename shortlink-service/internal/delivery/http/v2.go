package http

import (
	"net/http"

	"github.com/JIeeiroSst/shortlink-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type HandlerV2 struct {
	usecase *usecase.Usecase
}

func NewHandlerV2(usecase *usecase.Usecase) *HandlerV2 {
	return &HandlerV2{
		usecase: usecase,
	}
}

func NewRouter(r *gin.Engine, usecase *usecase.Usecase) {
	h := NewHandlerV2(usecase)
	r.GET("/:shortCode", h.RedirectLink)
}

func (h *HandlerV2) RedirectLink(c *gin.Context) {
	shortCode := c.Param("shortCode")

	originalURL, err := h.usecase.RedirectLink(c, shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}
