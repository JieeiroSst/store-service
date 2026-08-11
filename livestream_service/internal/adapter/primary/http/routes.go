package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

// NewNodeRouter serves only the SRS http_hooks contract - it runs on a
// transcode node, reached exclusively by that node's own local SRS
// sidecar over 127.0.0.1. Never exposed outside the pod.
func NewNodeRouter(srs *SRSWebhookHandler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	srsHooks := engine.Group("/api/srs")
	{
		srsHooks.POST("/on_publish", srs.OnPublish)
		srsHooks.POST("/on_unpublish", srs.OnUnpublish)
	}

	return engine
}

// NewEdgeRouter serves every viewer/client-facing route: rooms, ingest
// requests, viewer heartbeat/count, VOD listing, and chat. Stateless -
// meant to run behind a load-balanced Deployment scaled on connection
// count, independent of transcode node capacity.
func NewEdgeRouter(h *Handler, ws *WSHandler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", getHealth)

	rooms := engine.Group("/api/v1/rooms")
	{
		rooms.POST("", h.CreateRoom)
		rooms.GET("", h.ListRooms)
		rooms.GET("/:id", h.GetRoom)
		rooms.POST("/:id/stream-key/regenerate", h.RegenerateStreamKey)
		rooms.POST("/:id/ingest", h.RequestIngest)
		rooms.GET("/:id/stream", h.GetActiveStream)
		rooms.GET("/:id/recordings", h.ListRecordings)
		rooms.GET("/:id/viewers", h.GetViewerCount)
		rooms.POST("/:id/viewers/heartbeat", h.ViewerHeartbeat)
	}

	engine.GET("/ws/rooms/:id/chat", ws.ChatRoom)

	return engine
}

// NewAllInOneRouter mounts every route (edge + node) on a single engine -
// used only by the monolithic dev/docker-compose entrypoint (cmd/main.go),
// never by the split node/edge deployables the chart runs in production.
func NewAllInOneRouter(h *Handler, srs *SRSWebhookHandler, ws *WSHandler) *gin.Engine {
	engine := NewEdgeRouter(h, ws)

	srsHooks := engine.Group("/api/srs")
	{
		srsHooks.POST("/on_publish", srs.OnPublish)
		srsHooks.POST("/on_unpublish", srs.OnUnpublish)
	}

	return engine
}
