package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type createPocketRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) CreatePocket(c *gin.Context) {
	var req createPocketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pocket, err := h.pockets.CreatePocket(c.Request.Context(), c.Param("id"), req.Name)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, pocket)
}

func (h *Handler) ListPockets(c *gin.Context) {
	pockets, err := h.pockets.ListPockets(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, pockets)
}

type pocketAmountRequest struct {
	Amount int64 `json:"amount" binding:"required"`
}

func (h *Handler) DepositToPocket(c *gin.Context) {
	var req pocketAmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pocket, err := h.pockets.DepositToPocket(c.Request.Context(), c.Param("pocketId"), req.Amount)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, pocket)
}

func (h *Handler) WithdrawFromPocket(c *gin.Context) {
	var req pocketAmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pocket, err := h.pockets.WithdrawFromPocket(c.Request.Context(), c.Param("pocketId"), req.Amount)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, pocket)
}

func (h *Handler) ClosePocket(c *gin.Context) {
	wallet, err := h.pockets.ClosePocket(c.Request.Context(), c.Param("pocketId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, wallet)
}
