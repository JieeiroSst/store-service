package http

import (
	stdhttp "net/http"

	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/gin-gonic/gin"
)

type PartnerHandler struct {
	voucherSvc voucherapp.VoucherService
}

func NewPartnerHandler(voucherSvc voucherapp.VoucherService) *PartnerHandler {
	return &PartnerHandler{voucherSvc: voucherSvc}
}

type partnerRedeemRequest struct {
	VoucherID      string `json:"voucher_id" binding:"required"`
	PIN            string `json:"pin"`
	Amount         int64  `json:"amount" binding:"required"`
	Currency       string `json:"currency" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *PartnerHandler) Redeem(c *gin.Context) {
	var req partnerRedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	voucherID, err := shared.ParseVoucherID(req.VoucherID)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid voucher_id"}})
		return
	}

	out, err := h.voucherSvc.RedeemVoucher(c.Request.Context(), voucherapp.RedeemVoucherInput{
		VoucherID: voucherID, PIN: req.PIN, Amount: shared.NewMoney(req.Amount, req.Currency), IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, out)
}
