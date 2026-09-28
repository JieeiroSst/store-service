package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

func parseUint(c *gin.Context, name string) (uint, bool) {
	raw := c.Query(name)
	if raw == "" {
		return 0, true
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " must be a positive integer"})
		return 0, false
	}
	return uint(v), true
}

func parseTime(c *gin.Context, name string) (time.Time, bool) {
	raw := c.Query(name)
	if raw == "" {
		return time.Time{}, true
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " must be RFC3339, e.g. 2026-09-01T00:00:00+07:00"})
		return time.Time{}, false
	}
	return t, true
}

func (h *Handler) ListAuditDeliveries(c *gin.Context) {
	q := port.AuditQuery{
		Email:       c.Query("email"),
		Phone:       c.Query("phone"),
		Channel:     c.Query("channel"),
		Status:      c.Query("status"),
		SourceType:  c.Query("source_type"),
		RequestedBy: c.Query("requested_by"),
	}
	var ok bool
	if q.UserID, ok = parseUint(c, "user_id"); !ok {
		return
	}
	if q.SourceID, ok = parseUint(c, "source_id"); !ok {
		return
	}
	if q.BeforeID, ok = parseUint(c, "before_id"); !ok {
		return
	}
	limit, ok := parseUint(c, "limit")
	if !ok {
		return
	}
	q.Limit = int(limit)
	if q.From, ok = parseTime(c, "from"); !ok {
		return
	}
	if q.To, ok = parseTime(c, "to"); !ok {
		return
	}

	page, err := h.audit.ListDeliveries(c.Request.Context(), q)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

func (h *Handler) GetAuditContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	content, err := h.audit.GetContent(c.Request.Context(), uint(id))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, content)
}
