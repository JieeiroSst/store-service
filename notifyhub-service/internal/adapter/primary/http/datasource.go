package http

import (
	"net/http"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateDataSource(c *gin.Context) {
	var ds model.DataSource
	if err := c.ShouldBindJSON(&ds); err != nil {
		respondErr(c, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.sources.Create(c.Request.Context(), &ds)
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusCreated, created)
}

func (h *Handler) ListDataSources(c *gin.Context) {
	dss, err := h.sources.List(c.Request.Context())
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, dss)
}

func (h *Handler) GetDataSource(c *gin.Context) {
	ds, err := h.sources.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, ds)
}

func (h *Handler) UpdateDataSource(c *gin.Context) {
	ds, err := h.sources.Update(c.Request.Context(), c.Param("id"), bindTo[model.DataSource](c))
	if err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, ds)
}

func (h *Handler) DeleteDataSource(c *gin.Context) {
	if err := h.sources.Delete(c.Request.Context(), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	respond(c, http.StatusOK, gin.H{"deleted": true})
}
