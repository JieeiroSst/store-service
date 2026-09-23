package http

import (
	"net/http"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

type outcomeRequest struct {
	Outcome string `json:"outcome" binding:"required"`
}

type disputeRequest struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

func (h *Handler) ProposeResolution(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req outcomeRequest
	if !bind(c, &req) {
		return
	}
	m, err := h.resolution.Propose(c.Request.Context(), id, model.Outcome(req.Outcome))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMarketView(m))
}

func (h *Handler) DisputeResolution(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req disputeRequest
	if !bind(c, &req) {
		return
	}
	user, ok := h.caller(c, req.UserID)
	if !ok {
		return
	}
	m, err := h.resolution.Dispute(c.Request.Context(), id, user, req.Reason)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMarketView(m))
}

func (h *Handler) FinalizeResolution(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	m, err := h.resolution.Finalize(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMarketView(m))
}

func (h *Handler) ResolveMarket(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req outcomeRequest
	if !bind(c, &req) {
		return
	}
	m, err := h.resolution.Resolve(c.Request.Context(), id, model.Outcome(req.Outcome))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMarketView(m))
}
