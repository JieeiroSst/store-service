package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type createWalletRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

func (h *Handler) CreateWallet(c *gin.Context) {
	var req createWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet, err := h.wallets.CreateWallet(c.Request.Context(), req.UserID, req.Currency)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, wallet)
}

func (h *Handler) GetWallet(c *gin.Context) {
	wallet, err := h.wallets.GetWallet(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (h *Handler) GetWalletByUser(c *gin.Context) {
	wallet, err := h.wallets.GetWalletByUser(c.Request.Context(), c.Param("userId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, wallet)
}

type freezeWalletRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *Handler) FreezeWallet(c *gin.Context) {
	var req freezeWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet, err := h.wallets.FreezeWallet(c.Request.Context(), c.Param("id"), req.Reason)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (h *Handler) UnfreezeWallet(c *gin.Context) {
	wallet, err := h.wallets.UnfreezeWallet(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (h *Handler) CloseWallet(c *gin.Context) {
	wallet, err := h.wallets.CloseWallet(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, wallet)
}

type setLimitsRequest struct {
	DailyLimit          int64 `json:"daily_limit"`
	PerTransactionLimit int64 `json:"per_transaction_limit"`
}

func (h *Handler) SetLimits(c *gin.Context) {
	var req setLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet, err := h.wallets.SetLimits(c.Request.Context(), c.Param("id"), req.DailyLimit, req.PerTransactionLimit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, wallet)
}
