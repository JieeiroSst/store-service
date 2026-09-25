package http

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct{ usecase port.RecommendationUsecase }

func NewHandler(usecase port.RecommendationUsecase) *Handler { return &Handler{usecase: usecase} }

func pageQuery(c *gin.Context) port.PageQuery {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("page_size"))
	return port.PageQuery{Page: page, PageSize: size, Snapshot: c.Query("snapshot")}
}

func (h *Handler) ForUser(c *gin.Context) {
	page, err := h.usecase.ForUser(c.Request.Context(), c.Param("user_id"), pageQuery(c))
	respond(c, page, err)
}

func (h *Handler) Similar(c *gin.Context) {
	page, err := h.usecase.Similar(c.Request.Context(), c.Param("id"), pageQuery(c))
	respond(c, page, err)
}

func (h *Handler) Trending(c *gin.Context) {
	page, err := h.usecase.Trending(c.Request.Context(), pageQuery(c))
	respond(c, page, err)
}

func (h *Handler) NewReleases(c *gin.Context) {
	page, err := h.usecase.NewReleases(c.Request.Context(), pageQuery(c))
	respond(c, page, err)
}

func (h *Handler) ContinueWatching(c *gin.Context) {
	page, err := h.usecase.ContinueWatching(c.Request.Context(), c.Param("user_id"), pageQuery(c))
	respond(c, page, err)
}

func (h *Handler) History(c *gin.Context) {
	page, err := h.usecase.History(c.Request.Context(), c.Param("user_id"), pageQuery(c))
	respond(c, page, err)
}

func (h *Handler) RemoveFromHistory(c *gin.Context) {
	if err := h.usecase.RemoveFromHistory(c.Request.Context(), c.Param("user_id"), c.Param("video_id")); err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Home(c *gin.Context) {
	home, err := h.usecase.Home(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, home)
}

func respond(c *gin.Context, page *port.Page, err error) {
	if err != nil {
		fail(c, err)
		return
	}
	if page.Items == nil {
		page.Items = []model.Recommendation{}
	}
	c.JSON(http.StatusOK, page)
}

type eventRequest struct {
	UserID  string                `json:"user_id"`
	VideoID string                `json:"video_id"`
	Type    model.InteractionType `json:"type"`
	Value   float64               `json:"value"`
}

func (h *Handler) RecordEvent(c *gin.Context) {
	var req eventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, port.ErrInvalid)
		return
	}
	err := h.usecase.RecordEvent(c.Request.Context(), model.Interaction{
		UserID: req.UserID, VideoID: req.VideoID, Type: req.Type, Value: req.Value,
	})
	if err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusAccepted)
}

type videoRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Status      string   `json:"status"`
	Duration    float64  `json:"duration"`
}

func (h *Handler) UpsertVideo(c *gin.Context) {
	var req videoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, port.ErrInvalid)
		return
	}
	err := h.usecase.UpsertVideo(c.Request.Context(), model.Video{
		ID: c.Param("id"), Title: req.Title, Description: req.Description, Tags: req.Tags,
		Status: req.Status, Duration: req.Duration, CreatedAt: time.Now(),
	})
	if err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteVideo(c *gin.Context) {
	if err := h.usecase.DeleteVideo(c.Request.Context(), c.Param("id")); err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Readyz(c *gin.Context) {
	if err := h.usecase.Ready(c.Request.Context()); err != nil {
		log.Printf("readiness: %v", err)
		c.String(http.StatusServiceUnavailable, "not ready")
		return
	}
	c.String(http.StatusOK, "ok")
}

func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, port.ErrInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		log.Printf("%s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
