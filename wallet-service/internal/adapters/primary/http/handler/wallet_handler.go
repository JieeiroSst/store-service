package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
	"github.com/JIeeiroSst/wallet-service/internal/core/ports"
)

type WalletHandler struct {
	walletUC ports.WalletUseCase
}

func NewWalletHandler(walletUC ports.WalletUseCase) *WalletHandler {
	return &WalletHandler{walletUC: walletUC}
}

// POST /wallets
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id"  binding:"required,uuid"`
		Currency string `json:"currency" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := uuid.Parse(req.UserID)
	wallet, err := h.walletUC.CreateWallet(c.Request.Context(), userID, domain.Currency(req.Currency))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, wallet)
}

// GET /wallets/:id
func (h *WalletHandler) GetWallet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id"})
		return
	}
	wallet, err := h.walletUC.GetWallet(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrWalletNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, wallet)
}

// POST /wallets/:id/credit
func (h *WalletHandler) Credit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id"})
		return
	}
	var req struct {
		Amount      string `json:"amount"      binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.IsNegative() || amount.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}
	tx, err := h.walletUC.Credit(c.Request.Context(), id, amount, req.Description)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tx)
}

// POST /wallets/:id/freeze
func (h *WalletHandler) FreezeWallet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id"})
		return
	}
	if err := h.walletUC.FreezeWallet(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "frozen"})
}
