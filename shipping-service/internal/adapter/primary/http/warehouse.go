package http

import (
	"net/http"

	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type createWarehouseRequest struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Street        string `json:"street"`
	WardCode      string `json:"ward_code"`
	DistrictID    int    `json:"district_id"`
	ProvinceID    int    `json:"province_id"`
	CarrierShopID int64  `json:"carrier_shop_id"`
}

func (h *Handler) CreateWarehouse(c *gin.Context) {
	var req createWarehouseRequest
	if !bind(c, &req) {
		return
	}
	w, err := h.warehouses.Create(c.Request.Context(), port.CreateWarehouseInput(req))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *Handler) ListWarehouses(c *gin.Context) {
	ws, err := h.warehouses.List(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": ws})
}

type setActiveRequest struct {
	Active *bool `json:"active"`
}

func (h *Handler) SetWarehouseActive(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req setActiveRequest
	if !bind(c, &req) {
		return
	}
	if req.Active == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "active is required"})
		return
	}
	w, err := h.warehouses.SetActive(c.Request.Context(), id, *req.Active)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *Handler) Provinces(c *gin.Context) {
	items, err := h.locations.Provinces(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) Districts(c *gin.Context) {
	items, err := h.locations.Districts(c.Request.Context(), queryInt(c, "province_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) Wards(c *gin.Context) {
	items, err := h.locations.Wards(c.Request.Context(), queryInt(c, "district_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
