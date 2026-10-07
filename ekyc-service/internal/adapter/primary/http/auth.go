package http

import (
	"crypto/subtle"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

const callerKey = "ekyc.caller"

func (h *Handler) Authorize(c *gin.Context) {
	token := bearer(c.Request)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": port.ErrUnauthenticated.Error()})
		return
	}
	if h.isInternal(token) {
		c.Set(callerKey, "service")
		c.Next()
		return
	}
	userID, err := h.sessions.ValidateSession(c.Request.Context(), token)
	switch {
	case errors.Is(err, port.ErrUnauthenticated):
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	case err != nil:
		log.Printf("validate session: %v", err)
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "user-service is unavailable"})
		return
	case userID != c.Param("user_id"):
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": port.ErrForbidden.Error()})
		return
	}
	c.Set(callerKey, "user:"+userID)
	c.Next()
}

func (h *Handler) isInternal(token string) bool {
	match := 0
	for _, want := range h.internal {
		match |= subtle.ConstantTimeCompare([]byte(token), want)
	}
	return match == 1
}

func bearer(r *http.Request) string {
	if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return strings.TrimSpace(r.Header.Get("X-Session-Token"))
}
