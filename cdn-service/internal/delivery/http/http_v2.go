package http

import (
	"net/http"

	"path/filepath"

	"github.com/JIeeiroSst/cdn-service/config"
	"github.com/JIeeiroSst/cdn-service/internal/usecase"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HandlerV2 struct {
	usecase    *usecase.Usecase
	baseUpload config.BaseHostConfig
}

func NewHandlerV2(usecase *usecase.Usecase,
	baseUpload config.BaseHostConfig) *HandlerV2 {
	return &HandlerV2{
		usecase: usecase,
	}
}

func NewRouter(r *gin.Engine,
	usecase *usecase.Usecase,
	baseUpload config.BaseHostConfig) {
	h := NewHandlerV2(usecase, baseUpload)
	r.StaticFS("/media", http.Dir(baseUpload.BaseDirUpload))
	r.GET("/v1/files/:file_id/content", h.GetFile)
}

func (h *HandlerV2) GetFile(c *gin.Context) {
	lg := logger.WithContext(c)
	fileID := c.Param("file_id")
	if fileID == "" {
		c.JSON(400, gin.H{"error": "file_id is required"})
		return
	}

	res, err := h.usecase.GetFile(c.Request.Context(), fileID)
	if err != nil {
		lg.Error("error get file: %v", zap.Error(err))
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	filePath := filepath.Join(h.baseUpload.BaseDirUpload, res.StoragePath)

	c.File(filePath)
}
