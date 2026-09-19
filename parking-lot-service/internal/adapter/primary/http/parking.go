package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/parking-lot-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

type checkInRequest struct {
	LicensePlate string            `json:"license_plate" binding:"required"`
	VehicleType  model.VehicleType `json:"vehicle_type" binding:"required"`
	EntryGateID  *string           `json:"entry_gate_id"`
}

func (h *Handler) CheckIn(c *gin.Context) {
	var req checkInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.parking.CheckIn(c.Request.Context(), req.LicensePlate, req.VehicleType, req.EntryGateID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ticket)
}

type checkOutRequest struct {
	TicketID   string              `json:"ticket_id" binding:"required"`
	Method     model.PaymentMethod `json:"method" binding:"required"`
	ExitGateID *string             `json:"exit_gate_id"`
}

type checkOutResponse struct {
	Ticket  *model.Ticket  `json:"ticket"`
	Payment *model.Payment `json:"payment"`
}

func (h *Handler) CheckOut(c *gin.Context) {
	var req checkOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, payment, err := h.parking.CheckOut(c.Request.Context(), req.TicketID, req.Method, req.ExitGateID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, checkOutResponse{Ticket: ticket, Payment: payment})
}

func (h *Handler) GetTicket(c *gin.Context) {
	ticket, err := h.parking.GetTicket(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, ticket)
}

func (h *Handler) ListHistory(c *gin.Context) {
	plate := c.Query("plate")
	if plate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plate is required"})
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	tickets, err := h.parking.ListHistory(c.Request.Context(), plate, limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, tickets)
}

func (h *Handler) AvailableSpots(c *gin.Context) {
	availability, err := h.parking.AvailableSpots(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, availability)
}

func (h *Handler) ListRates(c *gin.Context) {
	rates, err := h.rates.ListRates(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, rates)
}
