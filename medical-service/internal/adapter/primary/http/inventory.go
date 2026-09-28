package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type receiveBatchRequest struct {
	BatchNumber string `json:"batch_number"`
	ExpiryDate  string `json:"expiry_date"`
	Quantity    int    `json:"quantity"`
	UnitCost    int64  `json:"unit_cost"`
	Supplier    string `json:"supplier"`
	ReceivedBy  string `json:"received_by"`
}

func (h *Handler) ReceiveBatch(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req receiveBatchRequest
	if !bind(c, &req) {
		return
	}
	expiry, err := model.ParseDate(req.ExpiryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expiry_date: " + err.Error()})
		return
	}
	batch, err := h.inventory.ReceiveBatch(c.Request.Context(), port.ReceiveBatchInput{
		MedicineID:  id,
		BatchNumber: req.BatchNumber,
		ExpiryDate:  expiry,
		Quantity:    req.Quantity,
		UnitCost:    req.UnitCost,
		Supplier:    req.Supplier,
		ReceivedBy:  req.ReceivedBy,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, batch)
}

type adjustBatchRequest struct {
	Quantity   *int   `json:"quantity"`
	Reason     string `json:"reason"`
	AdjustedBy string `json:"adjusted_by"`
}

func (h *Handler) AdjustBatch(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req adjustBatchRequest
	if !bind(c, &req) {
		return
	}
	if req.Quantity == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity is required"})
		return
	}
	batch, err := h.inventory.AdjustBatch(c.Request.Context(), port.AdjustBatchInput{
		BatchID:     id,
		NewQuantity: *req.Quantity,
		Reason:      req.Reason,
		AdjustedBy:  req.AdjustedBy,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, batch)
}

func (h *Handler) GetStock(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	stock, err := h.inventory.GetStock(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, stock)
}

func (h *Handler) ListMovements(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	movements, err := h.inventory.ListMovements(c.Request.Context(), id, parseList(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": movements})
}

func (h *Handler) ListExpiring(c *gin.Context) {
	days := 30
	if raw := c.Query("days"); raw != "" {
		var err error
		if days, err = strconv.Atoi(raw); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "days must be a number"})
			return
		}
	}
	batches, err := h.inventory.ListExpiring(c.Request.Context(), days)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": batches})
}

func (h *Handler) ListLowStock(c *gin.Context) {
	items, err := h.inventory.ListLowStock(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

type disposeRequest struct {
	DisposedBy string `json:"disposed_by"`
}

func (h *Handler) DisposeExpired(c *gin.Context) {
	var req disposeRequest
	if !bind(c, &req) {
		return
	}
	result, err := h.inventory.DisposeExpired(c.Request.Context(), req.DisposedBy)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
