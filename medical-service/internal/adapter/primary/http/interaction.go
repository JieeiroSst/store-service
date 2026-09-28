package http

import (
	"net/http"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type createInteractionRequest struct {
	IngredientA string `json:"ingredient_a"`
	IngredientB string `json:"ingredient_b"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

func (h *Handler) CreateInteraction(c *gin.Context) {
	var req createInteractionRequest
	if !bind(c, &req) {
		return
	}
	rule, err := h.interactions.CreateRule(c.Request.Context(), port.CreateInteractionInput{
		IngredientA: req.IngredientA,
		IngredientB: req.IngredientB,
		Severity:    model.Severity(req.Severity),
		Description: req.Description,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *Handler) ListInteractions(c *gin.Context) {
	rules, err := h.interactions.ListRules(c.Request.Context(), parseList(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rules})
}

type safetyCheckRequest struct {
	MedicineIDs        []int64  `json:"medicine_ids"`
	CurrentMedicineIDs []int64  `json:"current_medicine_ids"`
	Allergies          []string `json:"allergies"`
}

func (h *Handler) CheckSafety(c *gin.Context) {
	var req safetyCheckRequest
	if !bind(c, &req) {
		return
	}
	report, err := h.interactions.Check(c.Request.Context(), port.SafetyCheckInput(req))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"warnings":       report.Warnings,
		"blocked":        report.Blocked(),
		"needs_override": report.NeedsOverride(),
	})
}
