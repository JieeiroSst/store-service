package http

import (
	stdhttp "net/http"

	walletapp "github.com/JIeeiroSst/voucher-service/internal/application/wallet"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/wallet"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	svc walletapp.WalletService
}

func NewWalletHandler(svc walletapp.WalletService) *WalletHandler {
	return &WalletHandler{svc: svc}
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	ownerType := c.Param("ownerType")
	ownerID := c.Param("ownerId")

	balance, err := h.svc.GetBalance(c.Request.Context(), wallet.OwnerType(ownerType), ownerID)
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, gin.H{"amount": balance.Amount, "currency": balance.Currency})
}

type creditWalletRequest struct {
	Amount         int64  `json:"amount" binding:"required"`
	Currency       string `json:"currency" binding:"required"`
	RefType        string `json:"ref_type"`
	RefID          string `json:"ref_id"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *WalletHandler) Credit(c *gin.Context) {
	ownerType := c.Param("ownerType")
	ownerID := c.Param("ownerId")
	var req creditWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	if _, err := h.svc.GetOrCreateWallet(c.Request.Context(), wallet.OwnerType(ownerType), ownerID, req.Currency); err != nil {
		mapError(c, err)
		return
	}
	if err := h.svc.Credit(c.Request.Context(), wallet.OwnerType(ownerType), ownerID,
		shared.NewMoney(req.Amount, req.Currency), req.RefType, req.RefID, req.IdempotencyKey); err != nil {
		mapError(c, err)
		return
	}
	balance, err := h.svc.GetBalance(c.Request.Context(), wallet.OwnerType(ownerType), ownerID)
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, gin.H{"amount": balance.Amount, "currency": balance.Currency})
}
