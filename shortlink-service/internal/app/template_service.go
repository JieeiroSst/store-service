package app

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/shortlink-service/internal/adapters/repo"
	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
)

var ErrTemplateNotFound = errors.New("template not found")
var ErrTemplateHasLinks = errors.New("cannot delete template that has links assigned to it")
var ErrSlugExhausted = errors.New("unable to generate unique template slug")

type TemplateService struct {
	templates ports.TemplateRepository
}

func NewTemplateService(templates ports.TemplateRepository) *TemplateService {
	return &TemplateService{templates}
}

func (s *TemplateService) List(ctx context.Context, userID *string) ([]*domain.LinkTemplate, error) {
	return s.templates.List(ctx, userID)
}

func (s *TemplateService) Get(ctx context.Context, id string, userID *string) (*domain.LinkTemplate, error) {
	t, err := s.templates.GetByID(ctx, id, userID)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrTemplateNotFound
	}
	return t, err
}

type CreateTemplateInput struct {
	UserID      *string
	Name        string
	Description *string
	Settings    domain.LinkTemplateSettings
	IsDefault   bool
}

func (s *TemplateService) Create(ctx context.Context, in CreateTemplateInput) (*domain.LinkTemplate, error) {
	slug := ""
	for attempts := 0; attempts < 10; attempts++ {
		candidate, err := domain.GenerateTemplateSlug()
		if err != nil {
			return nil, err
		}
		exists, err := s.templates.ExistsBySlug(ctx, candidate)
		if err != nil {
			return nil, err
		}
		if !exists {
			slug = candidate
			break
		}
	}
	if slug == "" {
		return nil, ErrSlugExhausted
	}

	if in.IsDefault {
		if err := s.templates.UnsetDefaults(ctx, in.UserID, ""); err != nil {
			return nil, err
		}
	}

	tpl := &domain.LinkTemplate{
		UserID: in.UserID, Name: in.Name, Slug: slug, Description: in.Description,
		Settings: in.Settings, IsDefault: in.IsDefault,
	}
	if err := s.templates.Create(ctx, tpl); err != nil {
		return nil, err
	}
	return tpl, nil
}

type UpdateTemplateInput struct {
	Name        *string
	Description *string
	Settings    *domain.LinkTemplateSettings
	IsDefault   *bool
}

func (s *TemplateService) Update(ctx context.Context, id string, userID *string, in UpdateTemplateInput) (*domain.LinkTemplate, error) {
	if _, err := s.Get(ctx, id, userID); err != nil {
		return nil, err
	}

	if in.IsDefault != nil && *in.IsDefault {
		if err := s.templates.UnsetDefaults(ctx, userID, id); err != nil {
			return nil, err
		}
	}

	patch := map[string]interface{}{}
	if in.Name != nil {
		patch["name"] = *in.Name
	}
	if in.Description != nil {
		patch["description"] = *in.Description
	}
	if in.Settings != nil {
		patch["settings"] = mustSettingsJSON(*in.Settings)
	}
	if in.IsDefault != nil {
		patch["is_default"] = *in.IsDefault
	}
	if len(patch) == 0 {
		return nil, ErrNoUpdatesProvided
	}

	tpl, err := s.templates.Update(ctx, id, userID, patch)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	return tpl, nil
}

func (s *TemplateService) Delete(ctx context.Context, id string, userID *string) error {
	count, err := s.templates.LinkCountByTemplate(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrTemplateHasLinks
	}
	err = s.templates.Delete(ctx, id, userID)
	if errors.Is(err, repo.ErrNotFound) {
		return ErrTemplateNotFound
	}
	return err
}

func (s *TemplateService) SetDefault(ctx context.Context, id string, userID *string) (*domain.LinkTemplate, error) {
	if _, err := s.Get(ctx, id, userID); err != nil {
		return nil, err
	}
	if err := s.templates.UnsetDefaults(ctx, userID, ""); err != nil {
		return nil, err
	}
	return s.templates.SetDefault(ctx, id)
}

func mustSettingsJSON(s domain.LinkTemplateSettings) interface{} {
	return linkTemplateSettingsToJSONForPatch(s)
}
