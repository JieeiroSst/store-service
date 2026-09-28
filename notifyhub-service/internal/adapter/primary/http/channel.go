package http

import (
	"net/http"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateChannel(c *gin.Context) {
	var ch model.Channel
	if err := c.ShouldBindJSON(&ch); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.channels.Create(c.Request.Context(), &ch)
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusCreated, created)
}

func (h *Handler) ListChannels(c *gin.Context) {
	f := port.ChannelFilter{Type: c.Query("type")}
	if v := c.Query("active"); v != "" {
		b := v == "true"
		f.Active = &b
	}
	channels, err := h.channels.List(c.Request.Context(), f)
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, channels)
}

func (h *Handler) GetChannel(c *gin.Context) {
	ch, err := h.channels.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, ch)
}

func (h *Handler) UpdateChannel(c *gin.Context) {
	ch, err := h.channels.Update(c.Request.Context(), c.Param("id"), bindTo[model.Channel](c))
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, ch)
}

func (h *Handler) DeleteChannel(c *gin.Context) {
	if err := h.channels.Delete(c.Request.Context(), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, gin.H{"deleted": true})
}
