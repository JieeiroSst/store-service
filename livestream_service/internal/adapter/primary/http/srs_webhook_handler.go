package http

import (
	"net/http"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SRSWebhookHandler struct {
	publish port.PublishUsecase
	cfg     *config.Config
}

func NewSRSWebhookHandler(publish port.PublishUsecase, cfg *config.Config) *SRSWebhookHandler {
	return &SRSWebhookHandler{publish: publish, cfg: cfg}
}

// srsHookPayload mirrors the fields SRS sends on every http_hooks call;
// `stream` carries the RTMP stream name, which is the stream key itself
// (streamer pushes to rtmp://<node>/live/<streamKey>).
type srsHookPayload struct {
	Action string `json:"action"`
	Stream string `json:"stream"`
	App    string `json:"app"`
	IP     string `json:"ip"`
}

func (h *SRSWebhookHandler) OnPublish(c *gin.Context) {
	var payload srsHookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		reject(c)
		return
	}

	if err := h.publish.HandleOnPublish(c.Request.Context(), payload.Stream, h.cfg.Node.ID); err != nil {
		logger.WithContext(c.Request.Context()).Warn("srs on_publish rejected",
			zap.String("stream", payload.Stream), zap.Error(err))
		reject(c)
		return
	}
	allow(c)
}

func (h *SRSWebhookHandler) OnUnpublish(c *gin.Context) {
	var payload srsHookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		allow(c) // nothing useful to reject here; ack so SRS doesn't retry forever
		return
	}

	if err := h.publish.HandleOnUnpublish(c.Request.Context(), payload.Stream); err != nil {
		logger.WithContext(c.Request.Context()).Warn("srs on_unpublish failed",
			zap.String("stream", payload.Stream), zap.Error(err))
	}
	allow(c)
}

func allow(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"code": 0}) }
func reject(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"code": 1}) }
