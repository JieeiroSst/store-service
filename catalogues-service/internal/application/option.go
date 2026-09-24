package application

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type optionService struct{ repo port.OptionRepository }

func NewOptionService(repo port.OptionRepository) port.OptionService {
	return &optionService{repo: repo}
}

func prepareOption(o *model.Option) error {
	o.Name = strings.TrimSpace(o.Name)
	if o.Name == "" {
		return model.Invalid("name is required")
	}
	if o.Code = slugOr(o.Code, o.Name); o.Code == "" {
		return model.Invalid("code cannot be derived from name")
	}
	if o.Type == "" {
		o.Type = model.OptionText
	}
	if !o.Type.Valid() {
		return model.Invalid("type must be one of text, integer, float, boolean, date")
	}
	return nil
}

func (s *optionService) Create(ctx context.Context, o model.Option) (*model.Option, error) {
	if err := prepareOption(&o); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *optionService) Get(ctx context.Context, id int64) (*model.Option, error) {
	return s.repo.Get(ctx, id)
}

func (s *optionService) List(ctx context.Context) ([]model.Option, error) {
	return s.repo.List(ctx)
}

func (s *optionService) Update(ctx context.Context, o model.Option) (*model.Option, error) {
	if err := prepareOption(&o); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *optionService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
