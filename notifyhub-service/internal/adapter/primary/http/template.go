package http

import (
	"net/http"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateTemplate(c *gin.Context) {
	var t model.Template
	if err := c.ShouldBindJSON(&t); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.templates.Create(c.Request.Context(), &t)
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusCreated, created)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	ts, err := h.templates.List(c.Request.Context(), c.Query("channel"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, ts)
}

func (h *Handler) GetTemplate(c *gin.Context) {
	t, err := h.templates.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, t)
}

func (h *Handler) UpdateTemplate(c *gin.Context) {
	t, err := h.templates.Update(c.Request.Context(), c.Param("id"), bindTo[model.Template](c))
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, t)
}

func (h *Handler) DeleteTemplate(c *gin.Context) {
	if err := h.templates.Delete(c.Request.Context(), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, gin.H{"deleted": true})
}
