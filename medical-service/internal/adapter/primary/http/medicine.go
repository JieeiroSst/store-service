package http

import (
	"net/http"

	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type createMedicineRequest struct {
	Code                 string   `json:"code"`
	Name                 string   `json:"name"`
	Ingredients          []string `json:"ingredients"`
	AllergenGroups       []string `json:"allergen_groups"`
	DosageForm           string   `json:"dosage_form"`
	Strength             string   `json:"strength"`
	Unit                 string   `json:"unit"`
	RequiresPrescription bool     `json:"requires_prescription"`
	ReorderLevel         int      `json:"reorder_level"`
}

func (h *Handler) CreateMedicine(c *gin.Context) {
	var req createMedicineRequest
	if !bind(c, &req) {
		return
	}
	m, err := h.medicines.CreateMedicine(c.Request.Context(), port.CreateMedicineInput(req))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetMedicine(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	m, err := h.medicines.GetMedicine(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) ListMedicines(c *gin.Context) {
	list, err := h.medicines.ListMedicines(c.Request.Context(), c.Query("q"), parseList(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}
