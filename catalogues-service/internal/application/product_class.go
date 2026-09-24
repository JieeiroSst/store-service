package application

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type productClassService struct{ repo port.ProductClassRepository }

func NewProductClassService(repo port.ProductClassRepository) port.ProductClassService {
	return &productClassService{repo: repo}
}

func prepareProductClass(c *model.ProductClass) error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return model.Invalid("name is required")
	}
	if c.Slug = slugOr(c.Slug, c.Name); c.Slug == "" {
		return model.Invalid("slug cannot be derived from name")
	}
	c.OptionIDs = uniqueIDs(c.OptionIDs)
	return nil
}

func (s *productClassService) Create(ctx context.Context, c model.ProductClass) (*model.ProductClass, error) {
	if err := prepareProductClass(&c); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *productClassService) Get(ctx context.Context, id int64) (*model.ProductClass, error) {
	return s.repo.Get(ctx, id)
}

func (s *productClassService) List(ctx context.Context) ([]model.ProductClass, error) {
	return s.repo.List(ctx)
}

func (s *productClassService) Update(ctx context.Context, c model.ProductClass) (*model.ProductClass, error) {
	if err := prepareProductClass(&c); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *productClassService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}
