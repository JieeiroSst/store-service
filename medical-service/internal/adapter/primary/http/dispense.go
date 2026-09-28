package http

import (
	"net/http"

	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type dispenseItemRequest struct {
	MedicineID int64 `json:"medicine_id"`
	Quantity   int   `json:"quantity"`
}

type dispenseRequest struct {
	PatientRef         string                `json:"patient_ref"`
	PrescriptionRef    string                `json:"prescription_ref"`
	DispensedBy        string                `json:"dispensed_by"`
	Items              []dispenseItemRequest `json:"items"`
	CurrentMedicineIDs []int64               `json:"current_medicine_ids"`
	Allergies          []string              `json:"allergies"`
	OverrideReason     string                `json:"override_reason"`
}

func (h *Handler) CreateDispense(c *gin.Context) {
	var req dispenseRequest
	if !bind(c, &req) {
		return
	}
	items := make([]port.DispenseItemInput, len(req.Items))
	for i, it := range req.Items {
		items[i] = port.DispenseItemInput(it)
	}
	d, err := h.dispenses.Dispense(c.Request.Context(), port.DispenseInput{
		PatientRef:         req.PatientRef,
		PrescriptionRef:    req.PrescriptionRef,
		DispensedBy:        req.DispensedBy,
		Items:              items,
		CurrentMedicineIDs: req.CurrentMedicineIDs,
		Allergies:          req.Allergies,
		OverrideReason:     req.OverrideReason,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *Handler) GetDispense(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	d, err := h.dispenses.GetDispense(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *Handler) ListDispenses(c *gin.Context) {
	items, err := h.dispenses.ListByPatient(c.Request.Context(), c.Query("patient_ref"), parseList(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
