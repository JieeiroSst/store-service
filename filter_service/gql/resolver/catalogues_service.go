package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) CatalogQuery() generated.CatalogQueryResolver { return &catalogQueryResolver{r} }

type catalogQueryResolver struct{ *Resolver }

func (r *catalogQueryResolver) Options(ctx context.Context, obj *model.CatalogQuery) ([]*model.CatalogOption, error) {
	return r.Clients.CataloguesService.Options(ctx)
}

func (r *catalogQueryResolver) Option(ctx context.Context, obj *model.CatalogQuery, id int) (*model.CatalogOption, error) {
	return r.Clients.CataloguesService.Option(ctx, id)
}

func (r *catalogQueryResolver) ProductClasses(ctx context.Context, obj *model.CatalogQuery) ([]*model.CatalogProductClass, error) {
	return r.Clients.CataloguesService.ProductClasses(ctx)
}

func (r *catalogQueryResolver) ProductClass(ctx context.Context, obj *model.CatalogQuery, id int) (*model.CatalogProductClass, error) {
	return r.Clients.CataloguesService.ProductClass(ctx, id)
}

func (r *catalogQueryResolver) Products(ctx context.Context, obj *model.CatalogQuery, qArg *string, public *bool, categoryID *int, limit *int, offset *int) ([]*model.CatalogProduct, error) {
	return r.Clients.CataloguesService.Products(ctx, qArg, public, categoryID, limit, offset)
}

func (r *catalogQueryResolver) Product(ctx context.Context, obj *model.CatalogQuery, id int) (*model.CatalogProduct, error) {
	return r.Clients.CataloguesService.Product(ctx, id)
}

func (r *catalogQueryResolver) Categories(ctx context.Context, obj *model.CatalogQuery) ([]*model.CatalogCategory, error) {
	return r.Clients.CataloguesService.Categories(ctx)
}

func (r *catalogQueryResolver) Category(ctx context.Context, obj *model.CatalogQuery, id int) (*model.CatalogCategory, error) {
	return r.Clients.CataloguesService.Category(ctx, id)
}
