package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type templateService struct {
	templates port.TemplateRepository
	renderer  port.TemplateRenderer
	log       *zap.Logger
}

func NewTemplateService(templates port.TemplateRepository, renderer port.TemplateRenderer, log *zap.Logger) port.TemplateUsecase {
	return &templateService{templates: templates, renderer: renderer, log: log}
}

func (s *templateService) validate(t *model.Template) error {
	if t.Name == "" {
		return invalid("template name is required")
	}
	if t.Body == "" {
		return invalid("template body is required")
	}
	if !t.Channel.Valid() {
		return invalid("template channel must be email|sms|firebase")
	}
	if err := s.renderer.Compile(t); err != nil {
		return invalid("%v", err)
	}
	return nil
}

func (s *templateService) Create(ctx context.Context, t *model.Template) (*model.Template, error) {
	now := time.Now()
	t.ID = uuid.NewString()
	t.IsActive = true
	t.CreatedAt, t.UpdatedAt = now, now
	if err := s.validate(t); err != nil {
		return nil, err
	}
	if err := s.templates.Create(ctx, t); err != nil {
		s.renderer.Evict(t.ID)
		return nil, err
	}
	return t, nil
}

func (s *templateService) Get(ctx context.Context, id string) (*model.Template, error) {
	return s.templates.Get(ctx, id)
}

func (s *templateService) List(ctx context.Context, channel string) ([]*model.Template, error) {
	return s.templates.ListActive(ctx, channel)
}

func (s *templateService) Update(ctx context.Context, id string, apply func(*model.Template) error) (*model.Template, error) {
	t, err := s.templates.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	createdAt := t.CreatedAt
	if err := apply(t); err != nil {
		return nil, invalid("%v", err)
	}
	t.ID, t.CreatedAt, t.UpdatedAt = id, createdAt, time.Now()
	if err := s.validate(t); err != nil {
		return nil, err
	}
	if err := s.templates.Update(ctx, t); err != nil {
		s.renderer.Evict(id)
		return nil, err
	}
	return t, nil
}

func (s *templateService) Delete(ctx context.Context, id string) error {
	t, err := s.templates.Get(ctx, id)
	if err != nil {
		return err
	}
	t.IsActive = false
	t.UpdatedAt = time.Now()
	if err := s.templates.Update(ctx, t); err != nil {
		return err
	}
	s.renderer.Evict(id)
	return nil
}

func (s *templateService) WarmUp(ctx context.Context) (int, error) {
	templates, err := s.templates.ListActive(ctx, "")
	if err != nil {
		return 0, err
	}
	compiled := 0
	for _, t := range templates {
		if err := s.renderer.Compile(t); err != nil {
			s.log.Warn("template compile error",
				zap.String("id", t.ID),
				zap.String("name", t.Name),
				zap.Error(err),
			)
			continue
		}
		compiled++
	}
	return compiled, nil
}
