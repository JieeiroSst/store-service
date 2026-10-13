package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) SearchApplySchema(ctx context.Context) (*model.SearchSchemaStatus, error) {
	return restCall[*model.SearchSchemaStatus](ctx, r.Proxy, "search-service", rest.Request{Method: "PUT", Path: "/api/v1/admin/schema", Body: map[string]any{}}, "")
}

func (r *mutationResolver) SearchReindex(ctx context.Context, index []string, unmanaged *bool) ([]*model.SearchReindexResult, error) {
	return restCall[[]*model.SearchReindexResult](ctx, r.Proxy, "search-service", rest.Request{Method: "POST", Path: "/api/v1/admin/reindex", Query: restQuery("index", index, "unmanaged", unmanaged)}, "")
}

func (r *queryResolver) Search(ctx context.Context, q *string, index []string, sort *string, facets []string, cursor *string, size *int, filters map[string]any) (*model.SearchDocumentPage, error) {
	return restCall[*model.SearchDocumentPage](ctx, r.Proxy, "search-service", rest.Request{Method: "GET", Path: "/api/v1/search", Query: restQuery("q", q, "index", index, "sort", sort, "facets", facets, "cursor", cursor, "size", size, "filters", filters)}, "")
}

func (r *queryResolver) SearchAutocomplete(ctx context.Context, q string, index []string, size *int) (any, error) {
	return restCall[any](ctx, r.Proxy, "search-service", rest.Request{Method: "GET", Path: "/api/v1/autocomplete", Query: restQuery("q", q, "index", index, "size", size)}, "")
}

func (r *queryResolver) SearchDocuments(ctx context.Context, q *string, index []string, sort *string, facets []string, cursor *string, size *int, filters map[string]any) (*model.SearchDocumentPage, error) {
	return restCall[*model.SearchDocumentPage](ctx, r.Proxy, "search-service", rest.Request{Method: "GET", Path: "/api/v1/documents", Query: restQuery("q", q, "index", index, "sort", sort, "facets", facets, "cursor", cursor, "size", size, "filters", filters)}, "")
}

func (r *queryResolver) SearchDocument(ctx context.Context, index string, id string) (*model.SearchDocument, error) {
	return restCall[*model.SearchDocument](ctx, r.Proxy, "search-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/documents/%v/%v", index, id)}, "")
}

func (r *queryResolver) SearchSimilar(ctx context.Context, index string, id string, size *int) (any, error) {
	return restCall[any](ctx, r.Proxy, "search-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/documents/%v/%v/similar", index, id), Query: restQuery("size", size)}, "")
}

func (r *queryResolver) SearchIndices(ctx context.Context) ([]*model.SearchIndexInfo, error) {
	return restCall[[]*model.SearchIndexInfo](ctx, r.Proxy, "search-service", rest.Request{Method: "GET", Path: "/api/v1/indices"}, "")
}

func (r *queryResolver) SearchFields(ctx context.Context, index string) ([]*model.SearchFieldInfo, error) {
	return restCall[[]*model.SearchFieldInfo](ctx, r.Proxy, "search-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/indices/%v/fields", index)}, "")
}

func (r *queryResolver) SearchSchemaStatus(ctx context.Context) (*model.SearchSchemaStatus, error) {
	return restCall[*model.SearchSchemaStatus](ctx, r.Proxy, "search-service", rest.Request{Method: "GET", Path: "/api/v1/admin/schema"}, "")
}
