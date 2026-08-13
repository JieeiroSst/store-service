package app

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/shortlink-service/internal/adapters/repo"
	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
)

var ErrNoUpdatesProvided = errors.New("no updates provided")
var ErrShortCodeExhausted = errors.New("unable to generate unique short code")

// LinkService mirrors src/routes/links.ts.
type LinkService struct {
	links ports.LinkRepository
	cache ports.Cache
}

func NewLinkService(links ports.LinkRepository, cache ports.Cache) *LinkService {
	return &LinkService{links, cache}
}

func (s *LinkService) List(ctx context.Context, userID *string) ([]*domain.Link, error) {
	return s.links.List(ctx, ports.LinkFilter{UserID: userID})
}

func (s *LinkService) Get(ctx context.Context, id string, userID *string) (*domain.Link, error) {
	link, err := s.links.GetByID(ctx, id, ports.LinkFilter{UserID: userID})
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrLinkNotFound
	}
	return link, err
}

// generateUniqueShortCode mirrors the "generate + retry up to 10 times"
// loop shared by POST /api/links and POST /api/links/:id/duplicate.
func (s *LinkService) generateUniqueShortCode(ctx context.Context, customCode string) (string, error) {
	shortCode := customCode
	if shortCode == "" {
		code, err := domain.GenerateShortCode(8)
		if err != nil {
			return "", err
		}
		shortCode = code
	}

	for attempts := 0; attempts < 10; attempts++ {
		exists, err := s.links.ExistsByShortCode(ctx, shortCode)
		if err != nil {
			return "", err
		}
		if !exists {
			return shortCode, nil
		}
		code, err := domain.GenerateShortCode(8)
		if err != nil {
			return "", err
		}
		shortCode = code
	}
	return "", ErrShortCodeExhausted
}

// CreateLinkInput mirrors createLinkSchema in links.ts.
type CreateLinkInput struct {
	UserID                 *string
	TemplateID             *string
	OriginalURL            string
	Title                  *string
	Description            *string
	IOSAppStoreURL         *string
	AndroidAppStoreURL     *string
	WebFallbackURL         *string
	AppScheme              *string
	IOSUniversalLink       *string
	AndroidAppLink         *string
	DeepLinkPath           *string
	DeepLinkParameters     map[string]interface{}
	CustomCode             string
	UTMParameters          domain.UTMParameters
	TargetingRules         domain.TargetingRules
	OGTitle                *string
	OGDescription          *string
	OGImageURL             *string
	OGType                 string
	AttributionWindowHours int
	ExpiresAt              *string
}

func (s *LinkService) Create(ctx context.Context, in CreateLinkInput) (*domain.Link, error) {
	shortCode, err := s.generateUniqueShortCode(ctx, in.CustomCode)
	if err != nil {
		return nil, err
	}

	ogType := in.OGType
	if ogType == "" {
		ogType = "website"
	}
	attrWindow := in.AttributionWindowHours
	if attrWindow == 0 {
		attrWindow = domain.DefaultAttributionWindowHours
	}
	deepLinkParams := in.DeepLinkParameters
	if deepLinkParams == nil {
		deepLinkParams = map[string]interface{}{}
	}

	link := &domain.Link{
		UserID: in.UserID, TemplateID: in.TemplateID, ShortCode: shortCode, OriginalURL: in.OriginalURL,
		Title: in.Title, Description: in.Description,
		IOSAppStoreURL: in.IOSAppStoreURL, AndroidAppStoreURL: in.AndroidAppStoreURL, WebFallbackURL: in.WebFallbackURL,
		AppScheme: in.AppScheme, IOSUniversalLink: in.IOSUniversalLink, AndroidAppLink: in.AndroidAppLink,
		DeepLinkPath: in.DeepLinkPath, DeepLinkParameters: deepLinkParams,
		UTMParameters: in.UTMParameters, TargetingRules: in.TargetingRules,
		OGTitle: in.OGTitle, OGDescription: in.OGDescription, OGImageURL: in.OGImageURL, OGType: ogType,
		AttributionWindowHours: attrWindow, IsActive: true,
	}
	if in.ExpiresAt != nil {
		if t, err := parseISOTime(*in.ExpiresAt); err == nil {
			link.ExpiresAt = &t
		}
	}

	if err := s.links.Create(ctx, link); err != nil {
		return nil, err
	}
	return link, nil
}

// UpdateLinkInput mirrors updateLinkSchema (createLinkSchema.partial(),
// minus userId, plus isActive) in links.ts. Nil fields are "not provided"
// and excluded from the dynamic UPDATE, matching the JS `value !== undefined`
// filter.
type UpdateLinkInput struct {
	TemplateID             **string
	OriginalURL            *string
	Title                  **string
	Description            **string
	IOSAppStoreURL         **string
	AndroidAppStoreURL     **string
	WebFallbackURL         **string
	AppScheme              **string
	IOSUniversalLink       **string
	AndroidAppLink         **string
	DeepLinkPath           **string
	DeepLinkParameters     *map[string]interface{}
	UTMParameters          *domain.UTMParameters
	TargetingRules         *domain.TargetingRules
	OGTitle                **string
	OGDescription          **string
	OGImageURL             **string
	OGType                 *string
	AttributionWindowHours *int
	ExpiresAt              **string
	IsActive               *bool
}

func (s *LinkService) Update(ctx context.Context, id string, userID *string, in UpdateLinkInput) (*domain.Link, error) {
	oldLink, err := s.links.GetByID(ctx, id, ports.LinkFilter{})
	var oldShortCode, oldTemplateSlug string
	if err == nil {
		oldShortCode = oldLink.ShortCode
		if oldLink.TemplateSlug != nil {
			oldTemplateSlug = *oldLink.TemplateSlug
		}
	}

	patch := map[string]interface{}{}
	if in.TemplateID != nil {
		patch["template_id"] = derefDerefStr(in.TemplateID)
	}
	if in.OriginalURL != nil {
		patch["original_url"] = *in.OriginalURL
	}
	if in.Title != nil {
		patch["title"] = derefDerefStr(in.Title)
	}
	if in.Description != nil {
		patch["description"] = derefDerefStr(in.Description)
	}
	if in.IOSAppStoreURL != nil {
		patch["ios_app_store_url"] = derefDerefStr(in.IOSAppStoreURL)
	}
	if in.AndroidAppStoreURL != nil {
		patch["android_app_store_url"] = derefDerefStr(in.AndroidAppStoreURL)
	}
	if in.WebFallbackURL != nil {
		patch["web_fallback_url"] = derefDerefStr(in.WebFallbackURL)
	}
	if in.AppScheme != nil {
		patch["app_scheme"] = derefDerefStr(in.AppScheme)
	}
	if in.IOSUniversalLink != nil {
		patch["ios_universal_link"] = derefDerefStr(in.IOSUniversalLink)
	}
	if in.AndroidAppLink != nil {
		patch["android_app_link"] = derefDerefStr(in.AndroidAppLink)
	}
	if in.DeepLinkPath != nil {
		patch["deep_link_path"] = derefDerefStr(in.DeepLinkPath)
	}
	if in.DeepLinkParameters != nil {
		patch["deep_link_parameters"] = mustMapJSON(*in.DeepLinkParameters)
	}
	if in.UTMParameters != nil {
		patch["utm_parameters"] = mustUTMJSON(*in.UTMParameters)
	}
	if in.TargetingRules != nil {
		patch["targeting_rules"] = mustTargetingJSON(*in.TargetingRules)
	}
	if in.OGTitle != nil {
		patch["og_title"] = derefDerefStr(in.OGTitle)
	}
	if in.OGDescription != nil {
		patch["og_description"] = derefDerefStr(in.OGDescription)
	}
	if in.OGImageURL != nil {
		patch["og_image_url"] = derefDerefStr(in.OGImageURL)
	}
	if in.OGType != nil {
		patch["og_type"] = *in.OGType
	}
	if in.AttributionWindowHours != nil {
		patch["attribution_window_hours"] = *in.AttributionWindowHours
	}
	if in.ExpiresAt != nil {
		if *in.ExpiresAt == nil {
			patch["expires_at"] = nil
		} else if t, err := parseISOTime(**in.ExpiresAt); err == nil {
			patch["expires_at"] = t
		}
	}
	if in.IsActive != nil {
		patch["is_active"] = *in.IsActive
	}

	if len(patch) == 0 {
		return nil, ErrNoUpdatesProvided
	}

	link, err := s.links.Update(ctx, id, ports.LinkFilter{UserID: userID}, patch)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	newTemplateSlug := ""
	if link.TemplateSlug != nil {
		newTemplateSlug = *link.TemplateSlug
	}
	s.cache.Del(ctx, ports.LinkResolutionCacheKey(link.ShortCode, oldTemplateSlug))
	if oldShortCode != link.ShortCode {
		s.cache.Del(ctx, ports.LinkResolutionCacheKey(oldShortCode, oldTemplateSlug))
	}
	if newTemplateSlug != oldTemplateSlug {
		s.cache.Del(ctx, ports.LinkResolutionCacheKey(link.ShortCode, newTemplateSlug))
	}

	return link, nil
}

func (s *LinkService) Duplicate(ctx context.Context, id string, userID *string) (*domain.Link, error) {
	original, err := s.links.GetByID(ctx, id, ports.LinkFilter{UserID: userID})
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	shortCode, err := s.generateUniqueShortCode(ctx, "")
	if err != nil {
		return nil, err
	}

	var title *string
	if original.Title != nil {
		t := *original.Title + " (Copy)"
		title = &t
	}

	dup := &domain.Link{
		UserID: original.UserID, TemplateID: original.TemplateID, ShortCode: shortCode,
		OriginalURL: original.OriginalURL, Title: title, Description: original.Description,
		IOSAppStoreURL: original.IOSAppStoreURL, AndroidAppStoreURL: original.AndroidAppStoreURL,
		WebFallbackURL: original.WebFallbackURL, AppScheme: original.AppScheme,
		IOSUniversalLink: original.IOSUniversalLink, AndroidAppLink: original.AndroidAppLink,
		DeepLinkPath: original.DeepLinkPath, DeepLinkParameters: original.DeepLinkParameters,
		UTMParameters: original.UTMParameters, TargetingRules: original.TargetingRules,
		OGTitle: original.OGTitle, OGDescription: original.OGDescription, OGImageURL: original.OGImageURL,
		OGType: original.OGType, AttributionWindowHours: original.AttributionWindowHours,
		IsActive: true, ExpiresAt: original.ExpiresAt,
	}
	if err := s.links.Create(ctx, dup); err != nil {
		return nil, err
	}
	return dup, nil
}

func (s *LinkService) Delete(ctx context.Context, id string, userID *string) error {
	deleted, err := s.links.Delete(ctx, id, ports.LinkFilter{UserID: userID})
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrLinkNotFound
		}
		return err
	}
	templateSlug := ""
	if deleted.TemplateSlug != nil {
		templateSlug = *deleted.TemplateSlug
	} else if deleted.TemplateID != nil {
		slug, _ := s.links.TemplateSlugByID(ctx, *deleted.TemplateID)
		templateSlug = slug
	}
	s.cache.Del(ctx, ports.LinkResolutionCacheKey(deleted.ShortCode, templateSlug))
	return nil
}

func derefDerefStr(pp **string) interface{} {
	if pp == nil || *pp == nil {
		return nil
	}
	return **pp
}
