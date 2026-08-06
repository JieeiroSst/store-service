package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreatePublisher(c *gin.Context) {
	var publisher model.Publisher
	if err := c.ShouldBindJSON(&publisher); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.publisher.CreatePublisher(c.Request.Context(), &publisher)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetPublisher(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.publisher.GetPublisher(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListPublishers(c *gin.Context) {
	result, err := h.publisher.ListPublishers(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) UpdatePublisher(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var publisher model.Publisher
	if err := c.ShouldBindJSON(&publisher); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	publisher.ID = id
	result, err := h.publisher.UpdatePublisher(c.Request.Context(), &publisher)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeletePublisher(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.publisher.DeletePublisher(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
