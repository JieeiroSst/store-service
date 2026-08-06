package http

import (
	"net/http"
	"path/filepath"

	"github.com/JIeeiroSst/cdn-service/config"
	"github.com/JIeeiroSst/cdn-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	usecase    port.CDNUsecase
	baseUpload config.BaseHostConfig
}

func NewHandler(usecase port.CDNUsecase, cfg *config.Config) *Handler {
	return &Handler{usecase: usecase, baseUpload: cfg.BaseHost}
}

func (h *Handler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) GetFile(c *gin.Context) {
	lg := logger.WithContext(c.Request.Context())
	fileID := c.Param("file_id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id is required"})
		return
	}

	res, err := h.usecase.GetFile(c.Request.Context(), fileID)
	if err != nil {
		lg.Error("failed to get file", zap.String("file_id", fileID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.File(filepath.Join(h.baseUpload.BaseDirUpload, res.StoragePath))
}
