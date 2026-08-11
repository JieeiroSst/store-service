package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/metrics"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	rooms      port.RoomUsecase
	ingest     port.IngestUsecase
	viewers    port.ViewerUsecase
	moderation port.ModerationUsecase
}

func NewHandler(rooms port.RoomUsecase, ingest port.IngestUsecase, viewers port.ViewerUsecase, moderation port.ModerationUsecase) *Handler {
	return &Handler{rooms: rooms, ingest: ingest, viewers: viewers, moderation: moderation}
}

// writeError maps a usecase error to the right HTTP status - callers of
// this API need 403 vs 404 vs 503 vs 500 to behave sensibly, not a blanket
// 500 for everything.
func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrStreamKeyNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrNoNodeAvailable), errors.Is(err, port.ErrNodeAtCapacity):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

type createRoomRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// CreateRoom takes ownerUserId from the verified JWT, never from the
// request body - otherwise any caller could create rooms "owned" by
// someone else.
func (h *Handler) CreateRoom(c *gin.Context) {
	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.rooms.CreateRoom(c.Request.Context(), model.CreateRoomInput{
		OwnerUserID: middleware.UserID(c),
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, room)
}

func (h *Handler) ListRooms(c *gin.Context) {
	live := c.Query("live") == "true"
	rooms, err := h.rooms.ListRooms(c.Request.Context(), live)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func (h *Handler) GetRoom(c *gin.Context) {
	room, err := h.rooms.GetRoom(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *Handler) RegenerateStreamKey(c *gin.Context) {
	key, err := h.rooms.RegenerateStreamKey(c.Request.Context(), c.Param("id"), middleware.UserID(c), middleware.IsAdmin(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"streamKey": key})
}

func (h *Handler) RequestIngest(c *gin.Context) {
	endpoint, err := h.ingest.RequestIngestEndpoint(c.Request.Context(), c.Param("id"), middleware.UserID(c), middleware.IsAdmin(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, endpoint)
}

func (h *Handler) GetActiveStream(c *gin.Context) {
	stream, err := h.ingest.GetActiveStream(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active stream"})
		return
	}
	c.JSON(http.StatusOK, stream)
}

func (h *Handler) ListRecordings(c *gin.Context) {
	recs, err := h.ingest.ListRecordings(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, recs)
}

// GetPlaybackURL returns a signed, time-limited URL for the room's live
// output (if live) or its most recent VOD recording otherwise - see
// application.signPlaybackToken and config.PlaybackConfig.
func (h *Handler) GetPlaybackURL(c *gin.Context) {
	info, err := h.ingest.GetPlaybackURL(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (h *Handler) GetViewerCount(c *gin.Context) {
	count, err := h.viewers.GetViewerCount(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"viewers": count})
}

type viewerHeartbeatRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
}

// ViewerHeartbeat is called periodically (recommended ~15s) by the HLS
// player - this is the real online-viewer signal, since viewers pull HLS
// from a CDN and never connect to SRS directly. See port.ViewerCounter.
func (h *Handler) ViewerHeartbeat(c *gin.Context) {
	var req viewerHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.viewers.Heartbeat(c.Request.Context(), c.Param("id"), req.SessionID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type qoeRequest struct {
	BitrateKbps     float64 `json:"bitrateKbps"`
	BufferingEvents int     `json:"bufferingEvents"`
}

// ReportQoE ingests player-side quality-of-experience signals (bitrate,
// buffering) directly into Prometheus - a pure metrics passthrough, not a
// business rule, so there's no usecase/domain modeling for it.
func (h *Handler) ReportQoE(c *gin.Context) {
	var req qoeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.BitrateKbps > 0 {
		metrics.PlayerBitrateKbps.Observe(req.BitrateKbps)
	}
	for i := 0; i < req.BufferingEvents; i++ {
		metrics.PlayerBufferingEventsTotal.Inc()
	}
	c.Status(http.StatusNoContent)
}

// EndStream force-stops a room's live stream (owner or admin only).
func (h *Handler) EndStream(c *gin.Context) {
	err := h.moderation.ForceEndStream(c.Request.Context(), c.Param("id"), middleware.UserID(c), middleware.IsAdmin(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type banRequest struct {
	TargetUserID    string `json:"targetUserId" binding:"required"`
	DurationSeconds int    `json:"durationSeconds"`
}

// BanChat mutes a user in this room's chat (owner or admin only). Only
// blocks future messages - an already-open WebSocket connection isn't
// forcibly closed, see README's "Known limitations".
func (h *Handler) BanChat(c *gin.Context) {
	var req banRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	duration := time.Duration(req.DurationSeconds) * time.Second
	if duration <= 0 {
		duration = 24 * time.Hour
	}
	err := h.moderation.BanFromChat(c.Request.Context(), c.Param("id"), req.TargetUserID, middleware.UserID(c), middleware.IsAdmin(c), duration)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type unbanRequest struct {
	TargetUserID string `json:"targetUserId" binding:"required"`
}

func (h *Handler) UnbanChat(c *gin.Context) {
	var req unbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.moderation.UnbanFromChat(c.Request.Context(), c.Param("id"), req.TargetUserID, middleware.UserID(c), middleware.IsAdmin(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AdminDeleteRoom is platform-admin only (see middleware.RequireRole in
// routes.go) - deleting rooms you don't own isn't something a regular
// owner action ever needs.
func (h *Handler) AdminDeleteRoom(c *gin.Context) {
	err := h.moderation.DeleteRoom(c.Request.Context(), c.Param("id"), middleware.UserID(c), middleware.IsAdmin(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
