package http

import (
	"net/http"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	rooms   port.RoomUsecase
	ingest  port.IngestUsecase
	viewers port.ViewerUsecase
}

func NewHandler(rooms port.RoomUsecase, ingest port.IngestUsecase, viewers port.ViewerUsecase) *Handler {
	return &Handler{rooms: rooms, ingest: ingest, viewers: viewers}
}

type createRoomRequest struct {
	OwnerUserID string `json:"ownerUserId" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.rooms.CreateRoom(c.Request.Context(), model.CreateRoomInput{
		OwnerUserID: req.OwnerUserID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, room)
}

func (h *Handler) ListRooms(c *gin.Context) {
	live := c.Query("live") == "true"
	rooms, err := h.rooms.ListRooms(c.Request.Context(), live)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	key, err := h.rooms.RegenerateStreamKey(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"streamKey": key})
}

func (h *Handler) RequestIngest(c *gin.Context) {
	endpoint, err := h.ingest.RequestIngestEndpoint(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recs)
}

func (h *Handler) GetViewerCount(c *gin.Context) {
	count, err := h.viewers.GetViewerCount(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
