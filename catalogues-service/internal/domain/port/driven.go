package port

import (
	"context"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
)

type CategoryRepository interface {
	Create(ctx context.Context, c *model.Category) error
	Get(ctx context.Context, id int64) (*model.Category, error)
	List(ctx context.Context) ([]model.Category, error)
	Update(ctx context.Context, c *model.Category) error
	Delete(ctx context.Context, id int64) error
}

type ProductClassRepository interface {
	Create(ctx context.Context, c *model.ProductClass) error
	Get(ctx context.Context, id int64) (*model.ProductClass, error)
	List(ctx context.Context) ([]model.ProductClass, error)
	Update(ctx context.Context, c *model.ProductClass) error
	Delete(ctx context.Context, id int64) error
}

type OptionRepository interface {
	Create(ctx context.Context, o *model.Option) error
	Get(ctx context.Context, id int64) (*model.Option, error)
	List(ctx context.Context) ([]model.Option, error)
	Update(ctx context.Context, o *model.Option) error
	Delete(ctx context.Context, id int64) error
}

type StockRecordRepository interface {
	Create(ctx context.Context, s *model.StockRecord) error
	Get(ctx context.Context, id int64) (*model.StockRecord, error)
	ListByProducts(ctx context.Context, productIDs []int64) ([]model.StockRecord, error)
	Update(ctx context.Context, s *model.StockRecord) error
	Delete(ctx context.Context, id int64) error
	Allocate(ctx context.Context, id int64, qty int) error
	Release(ctx context.Context, id int64, qty int) error
}

type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) error
	Get(ctx context.Context, id int64) (*model.Product, error)
	List(ctx context.Context, f model.ProductFilter) ([]model.Product, error)
	Update(ctx context.Context, p *model.Product) error
	Delete(ctx context.Context, id int64) error

	AddImage(ctx context.Context, img *model.ProductImage) error
	DeleteImage(ctx context.Context, productID, imageID int64) error
	SetRecommendations(ctx context.Context, productID int64, recs []model.Recommendation) error
}
