package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/adapter/primary/http/middleware"
	"github.com/gin-gonic/gin"
)

// callerID is the authenticated user-service id of the caller.
func callerID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(middleware.UserID(c))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in session"})
		return 0, false
	}
	return id, true
}

// isOwner reports whether the caller may act on a resource owned by userID.
// Admins may act on everything.
func isOwner(c *gin.Context, userID int) bool {
	if middleware.IsAdmin(c) {
		return true
	}
	id, err := strconv.Atoi(middleware.UserID(c))
	return err == nil && id == userID
}

// requireAdmin guards listings that span every user.
func requireAdmin(c *gin.Context) bool {
	if !middleware.IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return false
	}
	return true
}

// authorizeBasket loads a basket and checks the caller owns it. A basket
// the caller doesn't own is reported as not found so ids can't be probed.
func (h *Handler) authorizeBasket(c *gin.Context, basketID int) bool {
	basket, err := h.basket.GetBasket(c.Request.Context(), basketID)
	if err != nil {
		writeError(c, err)
		return false
	}
	if !isOwner(c, basket.UserID) {
		writeError(c, common.ErrNotFound)
		return false
	}
	return true
}
