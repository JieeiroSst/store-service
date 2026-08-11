// Package metrics holds process-wide Prometheus collectors. Metrics are a
// cross-cutting infrastructure concern, not business logic, so - unlike
// the hexagonal ports elsewhere in this service - these are plain
// package-level vars referenced directly from whichever adapter/usecase
// observes them, the conventional Go/Prometheus client pattern.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "livestream_http_requests_total",
		Help: "HTTP requests by route, method, and status code.",
	}, []string{"route", "method", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "livestream_http_request_duration_seconds",
		Help:    "HTTP request latency by route and method.",
		Buckets: prometheus.DefBuckets,
	}, []string{"route", "method"})

	// Node role.
	ActiveStreams = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "livestream_node_active_streams",
		Help: "Number of ffmpeg jobs currently running on this node.",
	})
	NodeCapacity = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "livestream_node_capacity",
		Help: "Remaining stream capacity on this node (max_streams - active).",
	})
	FFmpegRestartsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "livestream_ffmpeg_restarts_total",
		Help: "Total ffmpeg process restarts across every stream on this node.",
	})

	// Edge role.
	ChatMessagesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "livestream_chat_messages_total",
		Help: "Total chat messages successfully published.",
	})
	ViewerHeartbeatsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "livestream_viewer_heartbeats_total",
		Help: "Total viewer heartbeat pings received.",
	})
	WebsocketConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "livestream_websocket_connections",
		Help: "Currently open chat WebSocket connections on this pod.",
	})

	// Player-reported QoE (see POST /api/v1/rooms/:id/qoe).
	PlayerBitrateKbps = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "livestream_player_bitrate_kbps",
		Help:    "Player-reported playback bitrate in kbps.",
		Buckets: []float64{200, 500, 1000, 2000, 4000, 6000, 8000, 12000},
	})
	PlayerBufferingEventsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "livestream_player_buffering_events_total",
		Help: "Total player-reported buffering/stall events.",
	})
)

// Handler serves the /metrics scrape endpoint.
func Handler() http.Handler { return promhttp.Handler() }

// GinMiddleware records HTTPRequestsTotal/HTTPRequestDuration for every
// request. Uses the matched route template (c.FullPath()), not the raw
// path, so per-room/per-stream IDs don't explode the metric's cardinality.
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		HTTPRequestsTotal.WithLabelValues(route, c.Request.Method, status).Inc()
		HTTPRequestDuration.WithLabelValues(route, c.Request.Method).Observe(time.Since(start).Seconds())
	}
}
