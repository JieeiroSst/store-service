package http

import (
	stdhttp "net/http"

	merchantapp "github.com/JIeeiroSst/voucher-service/internal/application/merchant"
	"github.com/JIeeiroSst/voucher-service/internal/domain/merchant"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/gin-gonic/gin"
)

type MerchantHandler struct {
	svc merchantapp.MerchantService
}

func NewMerchantHandler(svc merchantapp.MerchantService) *MerchantHandler {
	return &MerchantHandler{svc: svc}
}

type registerMerchantRequest struct {
	Name         string         `json:"name" binding:"required"`
	ProviderType string         `json:"provider_type" binding:"required"`
	Config       map[string]any `json:"config"`
}

func (h *MerchantHandler) Register(c *gin.Context) {
	var req registerMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	m, err := h.svc.RegisterMerchant(c.Request.Context(), merchantapp.RegisterMerchantInput{
		Name:         req.Name,
		ProviderType: shared.ProviderType(req.ProviderType),
		Config:       req.Config,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusCreated, toMerchantView(m))
}

func (h *MerchantHandler) Get(c *gin.Context) {
	id, err := shared.ParseMerchantID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid merchant id"}})
		return
	}
	m, err := h.svc.GetMerchant(c.Request.Context(), id)
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, toMerchantView(m))
}

func (h *MerchantHandler) List(c *gin.Context) {
	merchants, err := h.svc.ListMerchants(c.Request.Context())
	if err != nil {
		mapError(c, err)
		return
	}
	views := make([]merchantView, len(merchants))
	for i, m := range merchants {
		views[i] = toMerchantView(m)
	}
	c.JSON(stdhttp.StatusOK, gin.H{"merchants": views})
}

type merchantView struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	Status       string `json:"status"`
}

func toMerchantView(m *merchant.Merchant) merchantView {
	return merchantView{ID: m.ID.String(), Name: m.Name, ProviderType: string(m.ProviderType), Status: string(m.Status)}
}
