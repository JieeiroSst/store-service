package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) CatalogueCreateCategory(ctx context.Context, input model.CatalogueCategoryInput) (*model.CatalogueCategory, error) {
	return restCall[*model.CatalogueCategory](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "POST", Path: "/v1/categories", Body: input}, "")
}

func (r *mutationResolver) CatalogueUpdateCategory(ctx context.Context, id int, input model.CatalogueCategoryInput) (*model.CatalogueCategory, error) {
	return restCall[*model.CatalogueCategory](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "PUT", Path: rest.P("/v1/categories/%v", id), Body: input}, "")
}

func (r *mutationResolver) CatalogueDeleteCategory(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "DELETE", Path: rest.P("/v1/categories/%v", id)}, "")
}

func (r *mutationResolver) CatalogueCreateOption(ctx context.Context, input model.CatalogueOptionInput) (*model.CatalogueOption, error) {
	return restCall[*model.CatalogueOption](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "POST", Path: "/v1/options", Body: input}, "")
}

func (r *mutationResolver) CatalogueUpdateOption(ctx context.Context, id int, input model.CatalogueOptionInput) (*model.CatalogueOption, error) {
	return restCall[*model.CatalogueOption](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "PUT", Path: rest.P("/v1/options/%v", id), Body: input}, "")
}

func (r *mutationResolver) CatalogueDeleteOption(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "DELETE", Path: rest.P("/v1/options/%v", id)}, "")
}

func (r *mutationResolver) CatalogueCreateProductClass(ctx context.Context, input model.CatalogueProductClassInput) (*model.CatalogueProductClass, error) {
	return restCall[*model.CatalogueProductClass](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "POST", Path: "/v1/product-classes", Body: input}, "")
}

func (r *mutationResolver) CatalogueUpdateProductClass(ctx context.Context, id int, input model.CatalogueProductClassInput) (*model.CatalogueProductClass, error) {
	return restCall[*model.CatalogueProductClass](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "PUT", Path: rest.P("/v1/product-classes/%v", id), Body: input}, "")
}

func (r *mutationResolver) CatalogueDeleteProductClass(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "DELETE", Path: rest.P("/v1/product-classes/%v", id)}, "")
}

func (r *mutationResolver) CatalogueCreateProduct(ctx context.Context, input model.CatalogueProductInput) (*model.CatalogueProduct, error) {
	return restCall[*model.CatalogueProduct](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "POST", Path: "/v1/products", Body: input}, "")
}

func (r *mutationResolver) CatalogueUpdateProduct(ctx context.Context, id int, input model.CatalogueProductInput) (*model.CatalogueProduct, error) {
	return restCall[*model.CatalogueProduct](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "PUT", Path: rest.P("/v1/products/%v", id), Body: input}, "")
}

func (r *mutationResolver) CatalogueDeleteProduct(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "DELETE", Path: rest.P("/v1/products/%v", id)}, "")
}

func (r *mutationResolver) CatalogueAddProductImage(ctx context.Context, id int, input model.CatalogueImageInput) (*model.CatalogueProductImage, error) {
	return restCall[*model.CatalogueProductImage](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "POST", Path: rest.P("/v1/products/%v/images", id), Body: input}, "")
}

func (r *mutationResolver) CatalogueDeleteProductImage(ctx context.Context, id int, imageID int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "DELETE", Path: rest.P("/v1/products/%v/images/%v", id, imageID)}, "")
}

func (r *mutationResolver) CatalogueSetRecommendations(ctx context.Context, id int, recommendations []*model.CatalogueRecommendationInput) (*model.CatalogueProduct, error) {
	return restCall[*model.CatalogueProduct](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "PUT", Path: rest.P("/v1/products/%v/recommendations", id), Body: map[string]any{"recommendations": recommendations}}, "")
}

func (r *queryResolver) CatalogueCategories(ctx context.Context) ([]*model.CatalogueCategory, error) {
	return restCall[[]*model.CatalogueCategory](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "GET", Path: "/v1/categories"}, "")
}

func (r *queryResolver) CatalogueCategory(ctx context.Context, id int) (*model.CatalogueCategory, error) {
	return restCall[*model.CatalogueCategory](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "GET", Path: rest.P("/v1/categories/%v", id)}, "")
}

func (r *queryResolver) CatalogueOptions(ctx context.Context) ([]*model.CatalogueOption, error) {
	return restCall[[]*model.CatalogueOption](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "GET", Path: "/v1/options"}, "")
}

func (r *queryResolver) CatalogueOption(ctx context.Context, id int) (*model.CatalogueOption, error) {
	return restCall[*model.CatalogueOption](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "GET", Path: rest.P("/v1/options/%v", id)}, "")
}

func (r *queryResolver) CatalogueProductClasses(ctx context.Context) ([]*model.CatalogueProductClass, error) {
	return restCall[[]*model.CatalogueProductClass](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "GET", Path: "/v1/product-classes"}, "")
}

func (r *queryResolver) CatalogueProductClass(ctx context.Context, id int) (*model.CatalogueProductClass, error) {
	return restCall[*model.CatalogueProductClass](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "GET", Path: rest.P("/v1/product-classes/%v", id)}, "")
}

func (r *queryResolver) CatalogueProducts(ctx context.Context, q *string, categoryID *int, limit *int, offset *int) ([]*model.CatalogueProduct, error) {
	return restCall[[]*model.CatalogueProduct](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "GET", Path: "/v1/products", Query: mergeQuery(restQuery("q", q, "category_id", categoryID, "limit", limit, "offset", offset), "{\"public\":\"true\"}")}, "")
}

func (r *queryResolver) CatalogueAdminProducts(ctx context.Context, q *string, categoryID *int, public *bool, limit *int, offset *int) ([]*model.CatalogueProduct, error) {
	return restCall[[]*model.CatalogueProduct](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "GET", Path: "/v1/products", Query: restQuery("q", q, "category_id", categoryID, "public", public, "limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) CatalogueProduct(ctx context.Context, id int) (*model.CatalogueProduct, error) {
	return restCall[*model.CatalogueProduct](ctx, r.Proxy, "catalogues-service", rest.Request{Method: "GET", Path: rest.P("/v1/products/%v", id)}, "")
}
