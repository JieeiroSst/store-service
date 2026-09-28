package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	medicines    port.MedicineUsecase
	inventory    port.InventoryUsecase
	interactions port.InteractionUsecase
	dispenses    port.DispenseUsecase
	terminology  port.TerminologyUsecase
}

func NewHandler(
	medicines port.MedicineUsecase,
	inventory port.InventoryUsecase,
	interactions port.InteractionUsecase,
	dispenses port.DispenseUsecase,
	terminology port.TerminologyUsecase,
) *Handler {
	return &Handler{
		medicines:    medicines,
		inventory:    inventory,
		interactions: interactions,
		dispenses:    dispenses,
		terminology:  terminology,
	}
}

func writeError(c *gin.Context, err error) {
	var safety *port.SafetyError
	if errors.As(err, &safety) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "warnings": safety.Report.Warnings})
		return
	}

	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrConflict), errors.Is(err, port.ErrInsufficientStock):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrPrescriptionRequired):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func parseList(c *gin.Context) port.ListInput {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	return port.ListInput{Limit: limit, Offset: offset}
}

func bind(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}
