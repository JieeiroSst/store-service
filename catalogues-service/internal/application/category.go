package application

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type categoryService struct{ repo port.CategoryRepository }

func NewCategoryService(repo port.CategoryRepository) port.CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) prepare(ctx context.Context, c *model.Category) error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return model.Invalid("name is required")
	}
	if c.Slug = slugOr(c.Slug, c.Name); c.Slug == "" {
		return model.Invalid("slug cannot be derived from name")
	}
	c.AncestorsArePublic = true
	if c.ParentID != nil {
		if c.ID != 0 && *c.ParentID == c.ID {
			return model.Invalid("a category cannot be its own parent")
		}
		parent, err := s.repo.Get(ctx, *c.ParentID)
		if err != nil {
			return model.Invalid("parent category %d does not exist", *c.ParentID)
		}
		c.AncestorsArePublic = parent.IsPublic && parent.AncestorsArePublic
	}
	return nil
}

func (s *categoryService) Create(ctx context.Context, c model.Category) (*model.Category, error) {
	if err := s.prepare(ctx, &c); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *categoryService) Get(ctx context.Context, id int64) (*model.Category, error) {
	return s.repo.Get(ctx, id)
}

func (s *categoryService) List(ctx context.Context) ([]model.Category, error) {
	return s.repo.List(ctx)
}

func (s *categoryService) Update(ctx context.Context, c model.Category) (*model.Category, error) {
	if _, err := s.repo.Get(ctx, c.ID); err != nil {
		return nil, err
	}
	if err := s.prepare(ctx, &c); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *categoryService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
