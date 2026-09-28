package http

import (
	"net/http"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/gin-gonic/gin"
)

func (h *Handler) StartSession(c *gin.Context) {
	var req sessionRequest
	if !bind(c, &req) {
		return
	}
	s, err := h.vending.StartSession(c.Request.Context(), req.MachineID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toSession(s))
}

func (h *Handler) GetSession(c *gin.Context) {
	s, err := h.vending.GetSession(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toSession(s))
}

func (h *Handler) EndSession(c *gin.Context) {
	s, err := h.vending.EndSession(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toSession(s))
}

func (h *Handler) Reserve(c *gin.Context) {
	var req reserveRequest
	if !bind(c, &req) {
		return
	}
	r, err := h.vending.Reserve(c.Request.Context(), c.Param("id"), req.Slot)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toReservation(r))
}

func (h *Handler) CancelReservation(c *gin.Context) {
	r, err := h.vending.CancelReservation(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservation(r))
}

func (h *Handler) Checkout(c *gin.Context) {
	var req checkoutRequest
	if !bind(c, &req) {
		return
	}
	method, err := domain.ParsePaymentMethod(req.PaymentMethod)
	if err != nil {
		writeError(c, err)
		return
	}
	out, err := h.vending.Checkout(c.Request.Context(), c.Param("id"), port.CheckoutInput{
		Method:     method,
		WalletID:   req.WalletID,
		CouponCode: req.CouponCode,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toCheckout(out))
}

func (h *Handler) GetOrder(c *gin.Context) {
	o, err := h.vending.GetOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOrder(o))
}

func (h *Handler) ReportDispense(c *gin.Context) {
	var req dispenseRequest
	if !bind(c, &req) {
		return
	}
	if req.Dispensed == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispensed is required"})
		return
	}
	o, err := h.vending.ReportDispense(c.Request.Context(), c.Param("id"), *req.Dispensed)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOrder(o))
}

func (h *Handler) ListSessionOrders(c *gin.Context) {
	orders, err := h.vending.ListSessionOrders(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(orders, toOrder))
}

func (h *Handler) RefundOrder(c *gin.Context) {
	var req refundRequest
	if !bind(c, &req) {
		return
	}
	o, err := h.vending.RefundOrder(c.Request.Context(), c.Param("id"), req.Reason)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOrder(o))
}

func (h *Handler) GetPayment(c *gin.Context) {
	out, err := h.vending.GetPayment(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPaymentStatus(out))
}
