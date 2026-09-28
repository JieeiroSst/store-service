package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type aliasRequest struct {
	Alias     string `json:"alias"`
	Canonical string `json:"canonical"`
}

func (h *Handler) AddAlias(c *gin.Context) {
	var req aliasRequest
	if !bind(c, &req) {
		return
	}
	a, err := h.terminology.AddAlias(c.Request.Context(), req.Alias, req.Canonical)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, a)
}

type allergenGroupRequest struct {
	Ingredient    string `json:"ingredient"`
	AllergenGroup string `json:"allergen_group"`
}

func (h *Handler) AddAllergenGroup(c *gin.Context) {
	var req allergenGroupRequest
	if !bind(c, &req) {
		return
	}
	g, err := h.terminology.AddAllergenGroup(c.Request.Context(), req.Ingredient, req.AllergenGroup)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, g)
}

func (h *Handler) ListTerminology(c *gin.Context) {
	list, err := h.terminology.List(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}
