package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const version = "1.0.0"

type Handler struct {
	channels  port.ChannelUsecase
	sources   port.DataSourceUsecase
	templates port.TemplateUsecase
	jobs      port.JobUsecase
	history   port.HistoryUsecase
	log       *zap.Logger
}

func NewHandler(
	channels port.ChannelUsecase,
	sources port.DataSourceUsecase,
	templates port.TemplateUsecase,
	jobs port.JobUsecase,
	history port.HistoryUsecase,
	log *zap.Logger,
) *Handler {
	return &Handler{
		channels:  channels,
		sources:   sources,
		templates: templates,
		jobs:      jobs,
		history:   history,
		log:       log,
	}
}

func respond(c *gin.Context, code int, data interface{}) {
	c.JSON(code, gin.H{
		"data":       data,
		"ts":         time.Now().Unix(),
		"request_id": c.GetString(requestIDKey),
	})
}

func respondPage(c *gin.Context, data interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"total":      total,
		"page":       page,
		"page_size":  size,
		"ts":         time.Now().Unix(),
		"request_id": c.GetString(requestIDKey),
	})
}

func respondErr(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"error":      msg,
		"ts":         time.Now().Unix(),
		"request_id": c.GetString(requestIDKey),
	})
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		respondErr(c, http.StatusNotFound, err.Error())
	case errors.Is(err, port.ErrInvalidInput):
		respondErr(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, port.ErrConflict):
		respondErr(c, http.StatusConflict, err.Error())
	case errors.Is(err, port.ErrQueueFull), errors.Is(err, port.ErrQueueClosed):
		c.Header("Retry-After", "1")
		respondErr(c, http.StatusServiceUnavailable, err.Error())
	default:
		_ = c.Error(err)
		h.log.Error("request failed", zap.String("path", c.FullPath()), zap.Error(err))
		respondErr(c, http.StatusInternalServerError, "internal server error")
	}
}

// bindTo returns an apply func that decodes the JSON body onto an entity.
func bindTo[T any](c *gin.Context) func(*T) error {
	return func(v *T) error { return c.ShouldBindJSON(v) }
}

func paginate(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"ts":      time.Now().Unix(),
		"version": version,
	})
}

func (h *Handler) SchedulerStatus(c *gin.Context) {
	ids := h.jobs.ScheduledJobs()
	c.JSON(http.StatusOK, gin.H{
		"scheduled_jobs": ids,
		"count":          len(ids),
		"ts":             time.Now().Unix(),
	})
}
