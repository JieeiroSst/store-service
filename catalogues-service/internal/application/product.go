package application

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

const (
	defaultPageSize = 50
	maxPageSize     = 200
)

type productService struct{ repo port.ProductRepository }

func NewProductService(repo port.ProductRepository) port.ProductService {
	return &productService{repo: repo}
}

func prepareProduct(p *model.Product) error {
	p.Title = strings.TrimSpace(p.Title)
	if p.Title == "" {
		return model.Invalid("title is required")
	}
	if p.Slug = slugOr(p.Slug, p.Title); p.Slug == "" {
		return model.Invalid("slug cannot be derived from title")
	}
	if p.Structure == "" {
		p.Structure = model.StructureStandalone
	}
	if !p.Structure.Valid() {
		return model.Invalid("structure must be one of standalone, parent, child")
	}
	if p.Structure == model.StructureChild && p.ParentID == nil {
		return model.Invalid("a child product needs a parent_id")
	}
	if p.Structure != model.StructureChild && p.ParentID != nil {
		return model.Invalid("only a child product can have a parent_id")
	}
	if p.ParentID != nil && p.ID != 0 && *p.ParentID == p.ID {
		return model.Invalid("a product cannot be its own parent")
	}
	if p.UPC != nil {
		if u := strings.TrimSpace(*p.UPC); u == "" {
			p.UPC = nil
		} else {
			p.UPC = &u
		}
	}
	if p.Rating != nil && (*p.Rating < 0 || *p.Rating > 5) {
		return model.Invalid("rating must be between 0 and 5")
	}
	p.CategoryIDs = uniqueIDs(p.CategoryIDs)
	p.OptionIDs = uniqueIDs(p.OptionIDs)
	return nil
}

func (s *productService) Create(ctx context.Context, p model.Product) (*model.Product, error) {
	if err := prepareProduct(&p); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, &p); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, p.ID)
}

func (s *productService) Get(ctx context.Context, id int64) (*model.Product, error) {
	return s.repo.Get(ctx, id)
}

func (s *productService) List(ctx context.Context, f model.ProductFilter) ([]model.Product, error) {
	if f.Limit <= 0 {
		f.Limit = defaultPageSize
	}
	if f.Limit > maxPageSize {
		f.Limit = maxPageSize
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	f.Query = strings.TrimSpace(f.Query)
	return s.repo.List(ctx, f)
}

func (s *productService) Update(ctx context.Context, p model.Product) (*model.Product, error) {
	if err := prepareProduct(&p); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, &p); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, p.ID)
}

func (s *productService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *productService) AddImage(ctx context.Context, img model.ProductImage) (*model.ProductImage, error) {
	img.Original = strings.TrimSpace(img.Original)
	if img.Original == "" {
		return nil, model.Invalid("original is required")
	}
	if err := s.repo.AddImage(ctx, &img); err != nil {
		return nil, err
	}
	return &img, nil
}

func (s *productService) DeleteImage(ctx context.Context, productID, imageID int64) error {
	return s.repo.DeleteImage(ctx, productID, imageID)
}

func (s *productService) SetRecommendations(ctx context.Context, productID int64, recs []model.Recommendation) (*model.Product, error) {
	seen := make(map[int64]struct{}, len(recs))
	for _, r := range recs {
		if r.RecommendedID == productID {
			return nil, model.Invalid("a product cannot recommend itself")
		}
		if _, dup := seen[r.RecommendedID]; dup {
			return nil, model.Invalid("product %d is recommended twice", r.RecommendedID)
		}
		seen[r.RecommendedID] = struct{}{}
	}
	if err := s.repo.SetRecommendations(ctx, productID, recs); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, productID)
}
