package http

import (
	"strconv"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/pkg/authn"
	"github.com/gin-gonic/gin"
)

func (h *Handler) orderAccess(allowDriver bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := principal(c)
		if p.IsAdmin() {
			c.Next()
			return
		}
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			abortAuth(c, authn.ErrForbidden)
			return
		}
		order, err := h.orders.GetOrder(c.Request.Context(), id)
		if err != nil {
			writeError(c, err)
			c.Abort()
			return
		}
		if p.Owns(order.CustomerID) || (allowDriver && isDriver(order, p)) {
			c.Next()
			return
		}
		abortAuth(c, authn.ErrForbidden)
	}
}

func isDriver(order *model.Order, p authn.Principal) bool {
	return order.DriverID != nil && p.UserID != "" && *order.DriverID == p.UserID
}

func (h *Handler) driverOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := principal(c)
		if p.IsAdmin() {
			c.Next()
			return
		}
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			abortAuth(c, authn.ErrForbidden)
			return
		}
		a, err := h.assignments.GetAssignment(c.Request.Context(), id)
		if err != nil {
			writeError(c, err)
			c.Abort()
			return
		}
		if !p.Owns(a.DriverID) {
			abortAuth(c, authn.ErrForbidden)
			return
		}
		c.Next()
	}
}

func (h *Handler) orderDriverOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := principal(c)
		if p.IsAdmin() {
			c.Next()
			return
		}
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			abortAuth(c, authn.ErrForbidden)
			return
		}
		order, err := h.orders.GetOrder(c.Request.Context(), id)
		if err != nil {
			writeError(c, err)
			c.Abort()
			return
		}
		if !isDriver(order, p) {
			abortAuth(c, authn.ErrForbidden)
			return
		}
		c.Next()
	}
}

func listOrdersAccess(c *gin.Context) {
	p := principal(c)
	if p.IsAdmin() {
		c.Next()
		return
	}
	if c.Query("restaurant_id") != "" || !p.Owns(c.Query("customer_id")) {
		abortAuth(c, authn.ErrForbidden)
		return
	}
	c.Next()
}
