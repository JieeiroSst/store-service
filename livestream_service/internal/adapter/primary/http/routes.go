package http

import (
	"net/http"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/metrics"
	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

// NewNodeRouter serves the SRS http_hooks contract plus the internal
// force-unpublish route. SRS hooks are reached exclusively by this node's
// own local SRS sidecar over 127.0.0.1; the internal route is reached
// exclusively by the edge role, guarded by a shared secret (never a user
// JWT - see middleware.RequireInternalSecret). Neither is meant to be
// exposed outside the pod/cluster.
func NewNodeRouter(srs *SRSWebhookHandler, internal *InternalHandler, cfg *config.Config) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), metrics.GinMiddleware())

	engine.GET("/health", getHealth)
	engine.GET("/metrics", gin.WrapH(metrics.Handler()))

	srsHooks := engine.Group("/api/srs")
	{
		srsHooks.POST("/on_publish", srs.OnPublish)
		srsHooks.POST("/on_unpublish", srs.OnUnpublish)
	}

	internalRoutes := engine.Group("/internal", middleware.RequireInternalSecret(cfg.Internal.SharedSecret))
	{
		internalRoutes.POST("/streams/:streamKey/force-unpublish", internal.ForceUnpublish)
	}

	return engine
}

// NewEdgeRouter serves every viewer/client-facing route: rooms, ingest
// requests, viewer heartbeat/count, playback, QoE, chat, and moderation.
// Stateless - meant to run behind a load-balanced Deployment scaled on
// connection count, independent of transcode node capacity.
func NewEdgeRouter(h *Handler, ws *WSHandler, cfg *config.Config) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), metrics.GinMiddleware())

	engine.GET("/health", getHealth)
	engine.GET("/metrics", gin.WrapH(metrics.Handler()))

	auth := middleware.RequireAuth(cfg.Auth.JWTSecret)

	rooms := engine.Group("/api/v1/rooms")
	{
		// Public reads - watching shouldn't require an account.
		rooms.GET("", h.ListRooms)
		rooms.GET("/:id", h.GetRoom)
		rooms.GET("/:id/stream", h.GetActiveStream)
		rooms.GET("/:id/recordings", h.ListRecordings)
		rooms.GET("/:id/viewers", h.GetViewerCount)
		rooms.GET("/:id/playback", h.GetPlaybackURL)
		rooms.POST("/:id/viewers/heartbeat", h.ViewerHeartbeat)
		rooms.POST("/:id/qoe", h.ReportQoE)

		// Authenticated writes - ownership of the specific room is
		// enforced inside the usecase, not just here (RequireAuth only
		// proves identity, not permission on *this* room).
		rooms.POST("", auth, h.CreateRoom)
		rooms.POST("/:id/stream-key/regenerate", auth, h.RegenerateStreamKey)
		rooms.POST("/:id/ingest", auth, h.RequestIngest)
		rooms.POST("/:id/end", auth, h.EndStream)
		rooms.POST("/:id/chat/ban", auth, h.BanChat)
		rooms.POST("/:id/chat/unban", auth, h.UnbanChat)
	}

	admin := engine.Group("/api/v1/admin", auth, middleware.RequireRole("admin"))
	{
		admin.DELETE("/rooms/:id", h.AdminDeleteRoom)
	}

	engine.GET("/ws/rooms/:id/chat", ws.ChatRoom)

	return engine
}

// NewAllInOneRouter mounts every route (edge + node) on a single engine -
// used only by the monolithic dev/docker-compose entrypoint (cmd/main.go),
// never by the split node/edge deployables the chart runs in production.
func NewAllInOneRouter(h *Handler, srs *SRSWebhookHandler, internal *InternalHandler, ws *WSHandler, cfg *config.Config) *gin.Engine {
	engine := NewEdgeRouter(h, ws, cfg)

	srsHooks := engine.Group("/api/srs")
	{
		srsHooks.POST("/on_publish", srs.OnPublish)
		srsHooks.POST("/on_unpublish", srs.OnUnpublish)
	}
	internalRoutes := engine.Group("/internal", middleware.RequireInternalSecret(cfg.Internal.SharedSecret))
	{
		internalRoutes.POST("/streams/:streamKey/force-unpublish", internal.ForceUnpublish)
	}

	return engine
}
