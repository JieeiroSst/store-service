package http

import (
	stdhttp "net/http"

	corporateapp "github.com/JIeeiroSst/voucher-service/internal/application/corporate"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/gin-gonic/gin"
)

type CorporateHandler struct {
	svc corporateapp.CorporateService
}

func NewCorporateHandler(svc corporateapp.CorporateService) *CorporateHandler {
	return &CorporateHandler{svc: svc}
}

type registerCorporateRequest struct {
	Name         string `json:"name" binding:"required"`
	TaxCode      string `json:"tax_code"`
	ContactEmail string `json:"contact_email"`
}

func (h *CorporateHandler) Register(c *gin.Context) {
	var req registerCorporateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	corp, err := h.svc.RegisterCorporate(c.Request.Context(), corporateapp.RegisterCorporateInput{
		Name: req.Name, TaxCode: req.TaxCode, ContactEmail: req.ContactEmail,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusCreated, gin.H{"id": corp.ID.String(), "name": corp.Name})
}

type setBudgetRequest struct {
	Amount   int64  `json:"amount" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

func (h *CorporateHandler) SetBudget(c *gin.Context) {
	id, err := shared.ParseCorporateID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid corporate id"}})
		return
	}
	var req setBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	corp, err := h.svc.SetBudget(c.Request.Context(), id, shared.NewMoney(req.Amount, req.Currency))
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, gin.H{"id": corp.ID.String(), "budget_limit": corp.BudgetLimit})
}
