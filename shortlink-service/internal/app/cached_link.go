package app

import (
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
)

type cachedLink struct {
	ID                     string                       `json:"id"`
	UserID                 *string                      `json:"userId,omitempty"`
	OrganizationID         *string                      `json:"organizationId,omitempty"`
	TemplateID             *string                      `json:"templateId,omitempty"`
	ShortCode              string                       `json:"shortCode"`
	OriginalURL            string                       `json:"originalUrl"`
	Title                  *string                      `json:"title,omitempty"`
	Description            *string                      `json:"description,omitempty"`
	IOSAppStoreURL         *string                      `json:"iosAppStoreUrl,omitempty"`
	AndroidAppStoreURL     *string                      `json:"androidAppStoreUrl,omitempty"`
	WebFallbackURL         *string                      `json:"webFallbackUrl,omitempty"`
	AppScheme              *string                      `json:"appScheme,omitempty"`
	IOSUniversalLink       *string                      `json:"iosUniversalLink,omitempty"`
	AndroidAppLink         *string                      `json:"androidAppLink,omitempty"`
	DeepLinkPath           *string                      `json:"deepLinkPath,omitempty"`
	DeepLinkParameters     map[string]interface{}       `json:"deepLinkParameters,omitempty"`
	UTMParameters          domain.UTMParameters         `json:"utmParameters"`
	TargetingRules         domain.TargetingRules        `json:"targetingRules"`
	OGTitle                *string                      `json:"ogTitle,omitempty"`
	OGDescription          *string                      `json:"ogDescription,omitempty"`
	OGImageURL             *string                      `json:"ogImageUrl,omitempty"`
	OGType                 string                       `json:"ogType"`
	AttributionWindowHours int                          `json:"attributionWindowHours"`
	IsActive               bool                         `json:"isActive"`
	ExpiresAt              *time.Time                   `json:"expiresAt,omitempty"`
	WarnAt                 *time.Time                   `json:"warnAt,omitempty"`
	DisabledAt             *time.Time                   `json:"disabledAt,omitempty"`
	DisabledReason         *string                      `json:"disabledReason,omitempty"`
	AppendClickID          bool                         `json:"appendClickId"`
	TemplateSlug           *string                      `json:"templateSlug,omitempty"`
	TemplateSettings       *domain.LinkTemplateSettings `json:"templateSettings,omitempty"`
	OrgSettings            *domain.OrganizationSettings `json:"orgSettings,omitempty"`
	OwnerSuspendedAt       *time.Time                   `json:"ownerSuspendedAt,omitempty"`
}

func newCachedLink(l *domain.Link) cachedLink {
	return cachedLink{
		ID: l.ID, UserID: l.UserID, OrganizationID: l.OrganizationID, TemplateID: l.TemplateID,
		ShortCode: l.ShortCode, OriginalURL: l.OriginalURL, Title: l.Title, Description: l.Description,
		IOSAppStoreURL: l.IOSAppStoreURL, AndroidAppStoreURL: l.AndroidAppStoreURL, WebFallbackURL: l.WebFallbackURL,
		AppScheme: l.AppScheme, IOSUniversalLink: l.IOSUniversalLink, AndroidAppLink: l.AndroidAppLink,
		DeepLinkPath: l.DeepLinkPath, DeepLinkParameters: l.DeepLinkParameters,
		UTMParameters: l.UTMParameters, TargetingRules: l.TargetingRules,
		OGTitle: l.OGTitle, OGDescription: l.OGDescription, OGImageURL: l.OGImageURL, OGType: l.OGType,
		AttributionWindowHours: l.AttributionWindowHours, IsActive: l.IsActive, ExpiresAt: l.ExpiresAt,
		WarnAt: l.WarnAt, DisabledAt: l.DisabledAt, DisabledReason: l.DisabledReason, AppendClickID: l.AppendClickID,
		TemplateSlug: l.TemplateSlug, TemplateSettings: l.TemplateSettings, OrgSettings: l.OrgSettings,
		OwnerSuspendedAt: l.OwnerSuspendedAt,
	}
}

func (c cachedLink) toDomain() *domain.Link {
	return &domain.Link{
		ID: c.ID, UserID: c.UserID, OrganizationID: c.OrganizationID, TemplateID: c.TemplateID,
		ShortCode: c.ShortCode, OriginalURL: c.OriginalURL, Title: c.Title, Description: c.Description,
		IOSAppStoreURL: c.IOSAppStoreURL, AndroidAppStoreURL: c.AndroidAppStoreURL, WebFallbackURL: c.WebFallbackURL,
		AppScheme: c.AppScheme, IOSUniversalLink: c.IOSUniversalLink, AndroidAppLink: c.AndroidAppLink,
		DeepLinkPath: c.DeepLinkPath, DeepLinkParameters: c.DeepLinkParameters,
		UTMParameters: c.UTMParameters, TargetingRules: c.TargetingRules,
		OGTitle: c.OGTitle, OGDescription: c.OGDescription, OGImageURL: c.OGImageURL, OGType: c.OGType,
		AttributionWindowHours: c.AttributionWindowHours, IsActive: c.IsActive, ExpiresAt: c.ExpiresAt,
		WarnAt: c.WarnAt, DisabledAt: c.DisabledAt, DisabledReason: c.DisabledReason, AppendClickID: c.AppendClickID,
		TemplateSlug: c.TemplateSlug, TemplateSettings: c.TemplateSettings, OrgSettings: c.OrgSettings,
		OwnerSuspendedAt: c.OwnerSuspendedAt,
	}
}
