package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/shortlink-service/internal/app"
	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/gin-gonic/gin"
)

type LinksHandler struct {
	links *app.LinkService
}

func NewLinksHandler(links *app.LinkService) *LinksHandler {
	return &LinksHandler{links}
}

func linkResponseMap(l *domain.Link, includeClickCount, includeTemplateSlug bool) gin.H {
	m := gin.H{
		"id": l.ID, "user_id": l.UserID, "template_id": l.TemplateID,
		"short_code": l.ShortCode, "original_url": l.OriginalURL, "title": l.Title, "description": l.Description,
		"ios_app_store_url": l.IOSAppStoreURL, "android_app_store_url": l.AndroidAppStoreURL,
		"web_fallback_url": l.WebFallbackURL, "app_scheme": l.AppScheme,
		"ios_universal_link": l.IOSUniversalLink, "android_app_link": l.AndroidAppLink,
		"deep_link_path": l.DeepLinkPath, "deep_link_parameters": l.DeepLinkParameters,
		"utm_parameters": l.UTMParameters, "targeting_rules": l.TargetingRules,
		"og_title": l.OGTitle, "og_description": l.OGDescription, "og_image_url": l.OGImageURL, "og_type": l.OGType,
		"attribution_window_hours": l.AttributionWindowHours, "is_active": l.IsActive, "expires_at": l.ExpiresAt,
		"append_click_id": l.AppendClickID, "created_at": l.CreatedAt, "updated_at": l.UpdatedAt,
		"organization_id": l.OrganizationID, "warn_at": l.WarnAt, "disabled_at": l.DisabledAt,
		"disabled_reason": l.DisabledReason,
		// camelCase overrides (see NOTE above)
		"utmParameters": l.UTMParameters, "targetingRules": l.TargetingRules,
		"deepLinkParameters": l.DeepLinkParameters,
	}
	if includeClickCount {
		m["clickCount"] = l.ClickCount
		m["click_count"] = l.ClickCount
	}
	if includeTemplateSlug {
		m["template_slug"] = l.TemplateSlug
	}
	return m
}

func optionalUserID(c *gin.Context) *string {
	if v := c.Query("userId"); v != "" {
		return &v
	}
	return nil
}

func (h *LinksHandler) List(c *gin.Context) {
	links, err := h.links.List(c.Request.Context(), optionalUserID(c))
	if err != nil {
		respondInternalError(c, "Failed to list links", err)
		return
	}
	out := make([]gin.H, len(links))
	for i, l := range links {
		out[i] = linkResponseMap(l, true, true)
	}
	c.JSON(http.StatusOK, out)
}

func (h *LinksHandler) Get(c *gin.Context) {
	link, err := h.links.Get(c.Request.Context(), c.Param("id"), optionalUserID(c))
	if err != nil {
		if errors.Is(err, app.ErrLinkNotFound) {
			respondNotFound(c, "Link not found")
			return
		}
		respondInternalError(c, "Failed to get link", err)
		return
	}
	c.JSON(http.StatusOK, linkResponseMap(link, true, true))
}

type utmParamsRequest struct {
	Source   *string `json:"source"`
	Medium   *string `json:"medium"`
	Campaign *string `json:"campaign"`
	Term     *string `json:"term"`
	Content  *string `json:"content"`
}

func (r *utmParamsRequest) toDomain() domain.UTMParameters {
	if r == nil {
		return domain.UTMParameters{}
	}
	return domain.UTMParameters{
		Source: derefStr(r.Source), Medium: derefStr(r.Medium), Campaign: derefStr(r.Campaign),
		Term: derefStr(r.Term), Content: derefStr(r.Content),
	}
}

type targetingRulesRequest struct {
	Countries []string `json:"countries"`
	Devices   []string `json:"devices"`
	Languages []string `json:"languages"`
}

func (r *targetingRulesRequest) toDomain() domain.TargetingRules {
	if r == nil {
		return domain.TargetingRules{}
	}
	return domain.TargetingRules{Countries: r.Countries, Devices: r.Devices, Languages: r.Languages}
}

type createLinkRequest struct {
	UserID                 *string                `json:"userId" binding:"omitempty,uuid"`
	TemplateID             *string                `json:"templateId" binding:"omitempty,uuid"`
	OriginalURL            string                 `json:"originalUrl" binding:"required,url"`
	Title                  *string                `json:"title"`
	Description            *string                `json:"description"`
	IOSAppStoreURL         *string                `json:"iosAppStoreUrl" binding:"omitempty,url"`
	AndroidAppStoreURL     *string                `json:"androidAppStoreUrl" binding:"omitempty,url"`
	WebFallbackURL         *string                `json:"webFallbackUrl" binding:"omitempty,url"`
	AppScheme              *string                `json:"appScheme"`
	IOSUniversalLink       *string                `json:"iosUniversalLink" binding:"omitempty,url"`
	AndroidAppLink         *string                `json:"androidAppLink" binding:"omitempty,url"`
	DeepLinkPath           *string                `json:"deepLinkPath"`
	DeepLinkParameters     map[string]interface{} `json:"deepLinkParameters"`
	CustomCode             *string                `json:"customCode"`
	UTMParameters          *utmParamsRequest      `json:"utmParameters"`
	TargetingRules         *targetingRulesRequest `json:"targetingRules"`
	OGTitle                *string                `json:"ogTitle"`
	OGDescription          *string                `json:"ogDescription"`
	OGImageURL             *string                `json:"ogImageUrl" binding:"omitempty,url"`
	OGType                 *string                `json:"ogType"`
	AttributionWindowHours *int                   `json:"attributionWindowHours" binding:"omitempty,min=1,max=2160"`
	ExpiresAt              *string                `json:"expiresAt"`
}

func (h *LinksHandler) Create(c *gin.Context) {
	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	link, err := h.links.Create(c.Request.Context(), app.CreateLinkInput{
		UserID: req.UserID, TemplateID: req.TemplateID, OriginalURL: req.OriginalURL,
		Title: req.Title, Description: req.Description,
		IOSAppStoreURL: req.IOSAppStoreURL, AndroidAppStoreURL: req.AndroidAppStoreURL, WebFallbackURL: req.WebFallbackURL,
		AppScheme: req.AppScheme, IOSUniversalLink: req.IOSUniversalLink, AndroidAppLink: req.AndroidAppLink,
		DeepLinkPath: req.DeepLinkPath, DeepLinkParameters: req.DeepLinkParameters,
		CustomCode: derefStr(req.CustomCode), UTMParameters: req.UTMParameters.toDomain(), TargetingRules: req.TargetingRules.toDomain(),
		OGTitle: req.OGTitle, OGDescription: req.OGDescription, OGImageURL: req.OGImageURL, OGType: derefStr(req.OGType),
		AttributionWindowHours: derefInt(req.AttributionWindowHours), ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		respondInternalError(c, "Unable to generate unique short code", err)
		return
	}

	c.JSON(http.StatusOK, linkResponseMap(link, true, false))
}

type updateLinkRequest struct {
	TemplateID             *string                `json:"templateId"`
	OriginalURL            *string                `json:"originalUrl" binding:"omitempty,url"`
	Title                  *string                `json:"title"`
	Description            *string                `json:"description"`
	IOSAppStoreURL         *string                `json:"iosAppStoreUrl" binding:"omitempty,url"`
	AndroidAppStoreURL     *string                `json:"androidAppStoreUrl" binding:"omitempty,url"`
	WebFallbackURL         *string                `json:"webFallbackUrl" binding:"omitempty,url"`
	AppScheme              *string                `json:"appScheme"`
	IOSUniversalLink       *string                `json:"iosUniversalLink" binding:"omitempty,url"`
	AndroidAppLink         *string                `json:"androidAppLink" binding:"omitempty,url"`
	DeepLinkPath           *string                `json:"deepLinkPath"`
	DeepLinkParameters     map[string]interface{} `json:"deepLinkParameters"`
	UTMParameters          *utmParamsRequest      `json:"utmParameters"`
	TargetingRules         *targetingRulesRequest `json:"targetingRules"`
	OGTitle                *string                `json:"ogTitle"`
	OGDescription          *string                `json:"ogDescription"`
	OGImageURL             *string                `json:"ogImageUrl" binding:"omitempty,url"`
	OGType                 *string                `json:"ogType"`
	AttributionWindowHours *int                   `json:"attributionWindowHours" binding:"omitempty,min=1,max=2160"`
	ExpiresAt              *string                `json:"expiresAt"`
	IsActive               *bool                  `json:"isActive"`
}

func (h *LinksHandler) Update(c *gin.Context) {
	var req updateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	in := app.UpdateLinkInput{
		OriginalURL: req.OriginalURL, DeepLinkParameters: nilableMap(req.DeepLinkParameters),
		OGType: req.OGType, AttributionWindowHours: req.AttributionWindowHours, IsActive: req.IsActive,
	}
	if req.TemplateID != nil {
		in.TemplateID = ptrPtr(req.TemplateID)
	}
	if req.Title != nil {
		in.Title = ptrPtr(req.Title)
	}
	if req.Description != nil {
		in.Description = ptrPtr(req.Description)
	}
	if req.IOSAppStoreURL != nil {
		in.IOSAppStoreURL = ptrPtr(req.IOSAppStoreURL)
	}
	if req.AndroidAppStoreURL != nil {
		in.AndroidAppStoreURL = ptrPtr(req.AndroidAppStoreURL)
	}
	if req.WebFallbackURL != nil {
		in.WebFallbackURL = ptrPtr(req.WebFallbackURL)
	}
	if req.AppScheme != nil {
		in.AppScheme = ptrPtr(req.AppScheme)
	}
	if req.IOSUniversalLink != nil {
		in.IOSUniversalLink = ptrPtr(req.IOSUniversalLink)
	}
	if req.AndroidAppLink != nil {
		in.AndroidAppLink = ptrPtr(req.AndroidAppLink)
	}
	if req.DeepLinkPath != nil {
		in.DeepLinkPath = ptrPtr(req.DeepLinkPath)
	}
	if req.UTMParameters != nil {
		utm := req.UTMParameters.toDomain()
		in.UTMParameters = &utm
	}
	if req.TargetingRules != nil {
		tr := req.TargetingRules.toDomain()
		in.TargetingRules = &tr
	}
	if req.OGTitle != nil {
		in.OGTitle = ptrPtr(req.OGTitle)
	}
	if req.OGDescription != nil {
		in.OGDescription = ptrPtr(req.OGDescription)
	}
	if req.OGImageURL != nil {
		in.OGImageURL = ptrPtr(req.OGImageURL)
	}
	if req.ExpiresAt != nil {
		in.ExpiresAt = ptrPtr(req.ExpiresAt)
	}

	link, err := h.links.Update(c.Request.Context(), c.Param("id"), optionalUserID(c), in)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrLinkNotFound):
			respondNotFound(c, "Link not found")
		case errors.Is(err, app.ErrNoUpdatesProvided):
			respondValidationError(c, err)
		default:
			respondInternalError(c, "Failed to update link", err)
		}
		return
	}
	c.JSON(http.StatusOK, linkResponseMap(link, false, true))
}

func (h *LinksHandler) Duplicate(c *gin.Context) {
	link, err := h.links.Duplicate(c.Request.Context(), c.Param("id"), optionalUserID(c))
	if err != nil {
		if errors.Is(err, app.ErrLinkNotFound) {
			respondNotFound(c, "Link not found")
			return
		}
		respondInternalError(c, "Failed to duplicate link", err)
		return
	}
	c.JSON(http.StatusOK, linkResponseMap(link, true, false))
}

func (h *LinksHandler) Delete(c *gin.Context) {
	err := h.links.Delete(c.Request.Context(), c.Param("id"), optionalUserID(c))
	if err != nil {
		if errors.Is(err, app.ErrLinkNotFound) {
			respondNotFound(c, "Link not found")
			return
		}
		respondInternalError(c, "Failed to delete link", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ptrPtr(p *string) **string { return &p }

func nilableMap(m map[string]interface{}) *map[string]interface{} {
	if m == nil {
		return nil
	}
	return &m
}
