package http

import (
	stdhttp "net/http"

	distributionapp "github.com/JIeeiroSst/voucher-service/internal/application/distribution"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/gin-gonic/gin"
)

type DistributionHandler struct {
	svc distributionapp.DistributionService
}

func NewDistributionHandler(svc distributionapp.DistributionService) *DistributionHandler {
	return &DistributionHandler{svc: svc}
}

type createDistributionJobRequest struct {
	CorporateID  string   `json:"corporate_id" binding:"required"`
	MerchantID   string   `json:"merchant_id" binding:"required"`
	ProductSKU   string   `json:"product_sku" binding:"required"`
	DenomAmount  int64    `json:"denomination_amount" binding:"required"`
	Currency     string   `json:"currency" binding:"required"`
	Recipients   []string `json:"recipients" binding:"required"`
}

func (h *DistributionHandler) CreateJob(c *gin.Context) {
	var req createDistributionJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	corporateID, err := shared.ParseCorporateID(req.CorporateID)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid corporate_id"}})
		return
	}
	merchantID, err := shared.ParseMerchantID(req.MerchantID)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid merchant_id"}})
		return
	}

	job, err := h.svc.CreateJob(c.Request.Context(), distributionapp.CreateJobInput{
		CorporateID:  corporateID,
		MerchantID:   merchantID,
		ProductSKU:   req.ProductSKU,
		Denomination: shared.NewMoney(req.DenomAmount, req.Currency),
		Recipients:   req.Recipients,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusCreated, gin.H{"id": job.ID, "status": job.Status, "total_recipients": job.TotalRecipients})
}

func (h *DistributionHandler) Claim(c *gin.Context) {
	token := c.Param("token")
	voucherID, err := h.svc.ClaimVoucher(c.Request.Context(), token)
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, gin.H{"voucher_id": voucherID.String()})
}
