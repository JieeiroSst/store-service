package repo

import (
	"encoding/json"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"gorm.io/datatypes"
)

func toJSON(v interface{}) datatypes.JSON {
	if v == nil {
		return datatypes.JSON([]byte("{}"))
	}
	b, err := json.Marshal(v)
	if err != nil || len(b) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(b)
}

func utmToJSON(u domain.UTMParameters) datatypes.JSON { return toJSON(u) }

func jsonToUTM(j datatypes.JSON) domain.UTMParameters {
	var u domain.UTMParameters
	if len(j) == 0 {
		return u
	}
	_ = json.Unmarshal(j, &u)
	return u
}

func targetingToJSON(t domain.TargetingRules) datatypes.JSON { return toJSON(t) }

func jsonToTargeting(j datatypes.JSON) domain.TargetingRules {
	var t domain.TargetingRules
	if len(j) == 0 {
		return t
	}
	_ = json.Unmarshal(j, &t)
	return t
}

func mapToJSON(m map[string]interface{}) datatypes.JSON { return toJSON(m) }

func jsonToMap(j datatypes.JSON) map[string]interface{} {
	m := map[string]interface{}{}
	if len(j) == 0 {
		return m
	}
	_ = json.Unmarshal(j, &m)
	return m
}

func linkTemplateSettingsToJSON(s domain.LinkTemplateSettings) datatypes.JSON { return toJSON(s) }

func jsonToLinkTemplateSettings(j datatypes.JSON) domain.LinkTemplateSettings {
	var s domain.LinkTemplateSettings
	if len(j) == 0 {
		return s
	}
	_ = json.Unmarshal(j, &s)
	return s
}

func jsonToOrgSettings(j datatypes.JSON) domain.OrganizationSettings {
	var s domain.OrganizationSettings
	if len(j) == 0 {
		return s
	}
	_ = json.Unmarshal(j, &s)
	return s
}

func modelToLink(m *LinkModel) *domain.Link {
	ogType := "website"
	if m.OGType != nil {
		ogType = *m.OGType
	}
	return &domain.Link{
		ID:                     m.ID,
		UserID:                 m.UserID,
		OrganizationID:         m.OrganizationID,
		TemplateID:             m.TemplateID,
		ShortCode:              m.ShortCode,
		OriginalURL:            m.OriginalURL,
		Title:                  m.Title,
		Description:            m.Description,
		IOSAppStoreURL:         m.IOSAppStoreURL,
		AndroidAppStoreURL:     m.AndroidAppStoreURL,
		WebFallbackURL:         m.WebFallbackURL,
		AppScheme:              m.AppScheme,
		IOSUniversalLink:       m.IOSUniversalLink,
		AndroidAppLink:         m.AndroidAppLink,
		DeepLinkPath:           m.DeepLinkPath,
		DeepLinkParameters:     jsonToMap(m.DeepLinkParameters),
		UTMParameters:          jsonToUTM(m.UTMParameters),
		TargetingRules:         jsonToTargeting(m.TargetingRules),
		OGTitle:                m.OGTitle,
		OGDescription:          m.OGDescription,
		OGImageURL:             m.OGImageURL,
		OGType:                 ogType,
		AttributionWindowHours: m.AttributionWindowHours,
		IsActive:               m.IsActive,
		ExpiresAt:              m.ExpiresAt,
		WarnAt:                 m.WarnAt,
		DisabledAt:             m.DisabledAt,
		DisabledReason:         m.DisabledReason,
		AppendClickID:          m.AppendClickID,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}
}

func linkToModel(l *domain.Link) *LinkModel {
	ogType := l.OGType
	return &LinkModel{
		ID:                     l.ID,
		UserID:                 l.UserID,
		OrganizationID:         l.OrganizationID,
		TemplateID:             l.TemplateID,
		ShortCode:              l.ShortCode,
		OriginalURL:            l.OriginalURL,
		Title:                  l.Title,
		Description:            l.Description,
		IOSAppStoreURL:         l.IOSAppStoreURL,
		AndroidAppStoreURL:     l.AndroidAppStoreURL,
		WebFallbackURL:         l.WebFallbackURL,
		AppScheme:              l.AppScheme,
		IOSUniversalLink:       l.IOSUniversalLink,
		AndroidAppLink:         l.AndroidAppLink,
		DeepLinkPath:           l.DeepLinkPath,
		DeepLinkParameters:     mapToJSON(l.DeepLinkParameters),
		UTMParameters:          utmToJSON(l.UTMParameters),
		TargetingRules:         targetingToJSON(l.TargetingRules),
		OGTitle:                l.OGTitle,
		OGDescription:          l.OGDescription,
		OGImageURL:             l.OGImageURL,
		OGType:                 &ogType,
		AttributionWindowHours: l.AttributionWindowHours,
		IsActive:               l.IsActive,
		ExpiresAt:              l.ExpiresAt,
		WarnAt:                 l.WarnAt,
		DisabledAt:             l.DisabledAt,
		DisabledReason:         l.DisabledReason,
		AppendClickID:          l.AppendClickID,
	}
}
