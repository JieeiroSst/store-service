package port

import (
	"context"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
)

type CategoryService interface {
	Create(ctx context.Context, c model.Category) (*model.Category, error)
	Get(ctx context.Context, id int64) (*model.Category, error)
	List(ctx context.Context) ([]model.Category, error)
	Update(ctx context.Context, c model.Category) (*model.Category, error)
	Delete(ctx context.Context, id int64) error
}

type ProductClassService interface {
	Create(ctx context.Context, c model.ProductClass) (*model.ProductClass, error)
	Get(ctx context.Context, id int64) (*model.ProductClass, error)
	List(ctx context.Context) ([]model.ProductClass, error)
	Update(ctx context.Context, c model.ProductClass) (*model.ProductClass, error)
	Delete(ctx context.Context, id int64) error
}

type OptionService interface {
	Create(ctx context.Context, o model.Option) (*model.Option, error)
	Get(ctx context.Context, id int64) (*model.Option, error)
	List(ctx context.Context) ([]model.Option, error)
	Update(ctx context.Context, o model.Option) (*model.Option, error)
	Delete(ctx context.Context, id int64) error
}

type StockRecordService interface {
	Create(ctx context.Context, s model.StockRecord) (*model.StockRecord, error)
	Get(ctx context.Context, id int64) (*model.StockRecord, error)
	ListByProduct(ctx context.Context, productID int64) ([]model.StockRecord, error)
	Update(ctx context.Context, s model.StockRecord) (*model.StockRecord, error)
	Delete(ctx context.Context, id int64) error
	Allocate(ctx context.Context, id int64, qty int) (*model.StockRecord, error)
	Release(ctx context.Context, id int64, qty int) (*model.StockRecord, error)
}

type ProductService interface {
	Create(ctx context.Context, p model.Product) (*model.Product, error)
	Get(ctx context.Context, id int64) (*model.Product, error)
	List(ctx context.Context, f model.ProductFilter) ([]model.Product, error)
	Update(ctx context.Context, p model.Product) (*model.Product, error)
	Delete(ctx context.Context, id int64) error

	AddImage(ctx context.Context, img model.ProductImage) (*model.ProductImage, error)
	DeleteImage(ctx context.Context, productID, imageID int64) error
	SetRecommendations(ctx context.Context, productID int64, recs []model.Recommendation) (*model.Product, error)
}
