package http

import (
	stdhttp "net/http"
	"time"

	reportingapp "github.com/JIeeiroSst/voucher-service/internal/application/reporting"
	"github.com/gin-gonic/gin"
)

type ReportingHandler struct {
	svc reportingapp.ReportingService
}

func NewReportingHandler(svc reportingapp.ReportingService) *ReportingHandler {
	return &ReportingHandler{svc: svc}
}

func (h *ReportingHandler) RedemptionRate(c *gin.Context) {
	from, to, err := parseRange(c)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	reports, err := h.svc.RedemptionRateByMerchant(c.Request.Context(), from, to)
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, gin.H{"reports": reports})
}

func (h *ReportingHandler) CorporateSpend(c *gin.Context) {
	corporateID := c.Param("id")
	from, to, err := parseRange(c)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	report, err := h.svc.CorporateSpend(c.Request.Context(), corporateID, from, to)
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, report)
}

func parseRange(c *gin.Context) (time.Time, time.Time, error) {
	toStr := c.DefaultQuery("to", time.Now().UTC().Format(time.RFC3339))
	fromStr := c.DefaultQuery("from", time.Now().UTC().AddDate(0, -1, 0).Format(time.RFC3339))
	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to, nil
}
