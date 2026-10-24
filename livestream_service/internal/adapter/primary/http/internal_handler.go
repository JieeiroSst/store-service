package http

import (
	"net/http"

	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

// InternalHandler serves node-only, cluster-internal routes - reached
// exclusively by the edge role via port.NodeCaller, authenticated with a
// shared secret (see middleware.RequireInternalSecret), never a user JWT.
type InternalHandler struct {
	publish port.PublishUsecase
}

func NewInternalHandler(publish port.PublishUsecase) *InternalHandler {
	return &InternalHandler{publish: publish}
}

// ForceUnpublish stops this node's ffmpeg job for a stream key - the
// same effect as SRS calling on_unpublish, but triggered by an
// owner/admin moderation action (ModerationUsecase.ForceEndStream)
// instead of the streamer disconnecting.
func (h *InternalHandler) ForceUnpublish(c *gin.Context) {
	if err := h.publish.HandleOnUnpublish(c.Request.Context(), c.Param("streamKey")); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
