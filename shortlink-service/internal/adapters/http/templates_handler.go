package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/shortlink-service/internal/app"
	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/gin-gonic/gin"
)

type TemplatesHandler struct {
	templates *app.TemplateService
}

func NewTemplatesHandler(templates *app.TemplateService) *TemplatesHandler {
	return &TemplatesHandler{templates}
}

func templateJSON(t *domain.LinkTemplate) gin.H {
	return gin.H{
		"id": t.ID, "user_id": t.UserID, "name": t.Name, "slug": t.Slug, "description": t.Description,
		"settings": t.Settings, "is_default": t.IsDefault, "created_at": t.CreatedAt, "updated_at": t.UpdatedAt,
	}
}

func (h *TemplatesHandler) List(c *gin.Context) {
	templates, err := h.templates.List(c.Request.Context(), optionalUserID(c))
	if err != nil {
		respondInternalError(c, "Failed to list templates", err)
		return
	}
	out := make([]gin.H, len(templates))
	for i, t := range templates {
		out[i] = templateJSON(t)
	}
	c.JSON(http.StatusOK, out)
}

func (h *TemplatesHandler) Get(c *gin.Context) {
	t, err := h.templates.Get(c.Request.Context(), c.Param("id"), optionalUserID(c))
	if err != nil {
		if errors.Is(err, app.ErrTemplateNotFound) {
			respondNotFound(c, "Template not found")
			return
		}
		respondInternalError(c, "Failed to get template", err)
		return
	}
	c.JSON(http.StatusOK, templateJSON(t))
}

type templateSettingsRequest struct {
	DefaultIosURL                 *string                `json:"defaultIosUrl" binding:"omitempty,url"`
	DefaultAndroidURL             *string                `json:"defaultAndroidUrl" binding:"omitempty,url"`
	DefaultWebFallbackURL         *string                `json:"defaultWebFallbackUrl" binding:"omitempty,url"`
	DefaultAttributionWindowHours *int                   `json:"defaultAttributionWindowHours" binding:"omitempty,min=1,max=2160"`
	UTMParameters                 *utmParamsRequest      `json:"utmParameters"`
	TargetingRules                *targetingRulesRequest `json:"targetingRules"`
	ExpiresAfterDays              *int                   `json:"expiresAfterDays"`
}

func (r *templateSettingsRequest) toDomain() domain.LinkTemplateSettings {
	if r == nil {
		return domain.LinkTemplateSettings{}
	}
	s := domain.LinkTemplateSettings{
		DefaultIosURL: derefStr(r.DefaultIosURL), DefaultAndroidURL: derefStr(r.DefaultAndroidURL),
		DefaultWebFallbackURL: derefStr(r.DefaultWebFallbackURL), DefaultAttributionWindowHours: r.DefaultAttributionWindowHours,
		ExpiresAfterDays: r.ExpiresAfterDays,
	}
	if r.UTMParameters != nil {
		utm := r.UTMParameters.toDomain()
		s.UTMParameters = &utm
	}
	if r.TargetingRules != nil {
		tr := r.TargetingRules.toDomain()
		s.TargetingRules = &tr
	}
	return s
}

type createTemplateRequest struct {
	UserID      *string                  `json:"userId" binding:"omitempty,uuid"`
	Name        string                   `json:"name" binding:"required,min=1,max=255"`
	Description *string                  `json:"description"`
	Settings    *templateSettingsRequest `json:"settings"`
	IsDefault   bool                     `json:"isDefault"`
}

func (h *TemplatesHandler) Create(c *gin.Context) {
	var req createTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}
	t, err := h.templates.Create(c.Request.Context(), app.CreateTemplateInput{
		UserID: req.UserID, Name: req.Name, Description: req.Description,
		Settings: req.Settings.toDomain(), IsDefault: req.IsDefault,
	})
	if err != nil {
		respondInternalError(c, "Failed to create template", err)
		return
	}
	c.JSON(http.StatusOK, templateJSON(t))
}

type updateTemplateRequest struct {
	Name        *string                  `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string                  `json:"description"`
	Settings    *templateSettingsRequest `json:"settings"`
	IsDefault   *bool                    `json:"isDefault"`
}

func (h *TemplatesHandler) Update(c *gin.Context) {
	var req updateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	in := app.UpdateTemplateInput{Name: req.Name, Description: req.Description, IsDefault: req.IsDefault}
	if req.Settings != nil {
		settings := req.Settings.toDomain()
		in.Settings = &settings
	}

	t, err := h.templates.Update(c.Request.Context(), c.Param("id"), optionalUserID(c), in)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrTemplateNotFound):
			respondNotFound(c, "Template not found")
		case errors.Is(err, app.ErrNoUpdatesProvided):
			respondValidationError(c, err)
		default:
			respondInternalError(c, "Failed to update template", err)
		}
		return
	}
	c.JSON(http.StatusOK, templateJSON(t))
}

func (h *TemplatesHandler) Delete(c *gin.Context) {
	err := h.templates.Delete(c.Request.Context(), c.Param("id"), optionalUserID(c))
	if err != nil {
		switch {
		case errors.Is(err, app.ErrTemplateNotFound):
			respondNotFound(c, "Template not found")
		case errors.Is(err, app.ErrTemplateHasLinks):
			respondValidationError(c, err)
		default:
			respondInternalError(c, "Failed to delete template", err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TemplatesHandler) SetDefault(c *gin.Context) {
	t, err := h.templates.SetDefault(c.Request.Context(), c.Param("id"), optionalUserID(c))
	if err != nil {
		if errors.Is(err, app.ErrTemplateNotFound) {
			respondNotFound(c, "Template not found")
			return
		}
		respondInternalError(c, "Failed to set default template", err)
		return
	}
	c.JSON(http.StatusOK, templateJSON(t))
}
