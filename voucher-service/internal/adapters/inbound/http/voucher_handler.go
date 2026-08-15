package http

import (
	stdhttp "net/http"

	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	"github.com/gin-gonic/gin"
)

type VoucherHandler struct {
	svc voucherapp.VoucherService
}

func NewVoucherHandler(svc voucherapp.VoucherService) *VoucherHandler {
	return &VoucherHandler{svc: svc}
}

type issueVouchersRequest struct {
	MerchantID     string `json:"merchant_id" binding:"required"`
	ProductSKU     string `json:"product_sku" binding:"required"`
	DenomAmount    int64  `json:"denomination_amount" binding:"required"`
	Currency       string `json:"currency" binding:"required"`
	Quantity       int    `json:"quantity" binding:"required"`
	ExpiresInDays  int    `json:"expires_in_days"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *VoucherHandler) Issue(c *gin.Context) {
	var req issueVouchersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	merchantID, err := shared.ParseMerchantID(req.MerchantID)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid merchant_id"}})
		return
	}

	issued, err := h.svc.IssueVouchers(c.Request.Context(), voucherapp.IssueVouchersInput{
		MerchantID:     merchantID,
		ProductSKU:     req.ProductSKU,
		Denomination:   shared.NewMoney(req.DenomAmount, req.Currency),
		Quantity:       req.Quantity,
		ExpiresInDays:  req.ExpiresInDays,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusCreated, gin.H{"vouchers": toIssuedVoucherViews(issued)})
}

type redeemVoucherRequest struct {
	PIN            string `json:"pin"`
	Amount         int64  `json:"amount" binding:"required"`
	Currency       string `json:"currency" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *VoucherHandler) Redeem(c *gin.Context) {
	voucherID, err := shared.ParseVoucherID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid voucher id"}})
		return
	}
	var req redeemVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}

	out, err := h.svc.RedeemVoucher(c.Request.Context(), voucherapp.RedeemVoucherInput{
		VoucherID:      voucherID,
		PIN:            req.PIN,
		Amount:         shared.NewMoney(req.Amount, req.Currency),
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, out)
}

type activateVoucherRequest struct {
	OwnerType string `json:"owner_type" binding:"required"`
	OwnerID   string `json:"owner_id" binding:"required"`
}

func (h *VoucherHandler) Activate(c *gin.Context) {
	voucherID, err := shared.ParseVoucherID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid voucher id"}})
		return
	}
	var req activateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}

	v, err := h.svc.ActivateVoucher(c.Request.Context(), voucherID, voucher.OwnerType(req.OwnerType), req.OwnerID)
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, toVoucherView(v))
}

func (h *VoucherHandler) Validate(c *gin.Context) {
	voucherID, err := shared.ParseVoucherID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid voucher id"}})
		return
	}
	pin := c.Query("pin")

	result, err := h.svc.ValidateVoucher(c.Request.Context(), voucherapp.ValidateVoucherInput{VoucherID: voucherID, PIN: pin})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, result)
}

func (h *VoucherHandler) Revoke(c *gin.Context) {
	voucherID, err := shared.ParseVoucherID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid voucher id"}})
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&body)

	if err := h.svc.RevokeVoucher(c.Request.Context(), voucherID, body.Reason); err != nil {
		mapError(c, err)
		return
	}
	c.Status(stdhttp.StatusNoContent)
}

func (h *VoucherHandler) Get(c *gin.Context) {
	voucherID, err := shared.ParseVoucherID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid voucher id"}})
		return
	}
	v, err := h.svc.GetVoucher(c.Request.Context(), voucherID)
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, toVoucherView(v))
}

func (h *VoucherHandler) ListMine(c *gin.Context) {
	ownerID := c.Query("owner_id")
	vouchers, err := h.svc.ListVouchers(c.Request.Context(), voucher.OwnerTypeUser, ownerID)
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, gin.H{"vouchers": toVoucherViews(vouchers)})
}

type voucherView struct {
	ID         string `json:"id"`
	MerchantID string `json:"merchant_id"`
	Code       string `json:"code"`
	Status     string `json:"status"`
	Version    int    `json:"version"`
}

func toVoucherView(v *voucher.Voucher) voucherView {
	return voucherView{ID: v.ID.String(), MerchantID: v.MerchantID.String(), Code: v.Code, Status: string(v.Status), Version: v.Version}
}

func toVoucherViews(vouchers []*voucher.Voucher) []voucherView {
	views := make([]voucherView, len(vouchers))
	for i, v := range vouchers {
		views[i] = toVoucherView(v)
	}
	return views
}

type issuedVoucherView struct {
	voucherView
	PIN string `json:"pin,omitempty"`
}

func toIssuedVoucherViews(issued []voucherapp.IssuedVoucher) []issuedVoucherView {
	views := make([]issuedVoucherView, len(issued))
	for i, iv := range issued {
		views[i] = issuedVoucherView{voucherView: toVoucherView(iv.Voucher), PIN: iv.PlaintextPIN}
	}
	return views
}
