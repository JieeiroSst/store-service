package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) PostCreatePost(ctx context.Context, name string, content string, description *string, categoryID *string, file *graphql.Upload) (*model.PostPost, error) {
	return restMultipart[*model.PostPost](ctx, r.Proxy, "post-service", rest.Request{Method: "POST", Path: "/api/v1/post"}, map[string]any{"name": name, "content": content, "description": description, "category_id": categoryID}, uploadFiles("file", file), "")
}

func (r *mutationResolver) PostUpdatePost(ctx context.Context, id string, name *string, content *string, description *string) (*model.PostPost, error) {
	return restCall[*model.PostPost](ctx, r.Proxy, "post-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/post/%v", id), Body: map[string]any{"name": name, "content": content, "description": description}}, "")
}

func (r *mutationResolver) PostCreateCategory(ctx context.Context, name string, description *string) (*model.PostCategory, error) {
	return restCall[*model.PostCategory](ctx, r.Proxy, "post-service", rest.Request{Method: "POST", Path: "/api/v1/category", Body: map[string]any{"name": name, "description": description}}, "")
}

func (r *mutationResolver) PostUpdateCategory(ctx context.Context, id string, name *string, description *string) (*model.PostCategory, error) {
	return restCall[*model.PostCategory](ctx, r.Proxy, "post-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/category/%v", id), Body: map[string]any{"name": name, "description": description}}, "")
}

func (r *mutationResolver) PostDeleteCategory(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "post-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/category/%v", id)}, "")
}

func (r *queryResolver) Posts(ctx context.Context, cursor *string, limit *int) (*model.PostPostPage, error) {
	return restCall[*model.PostPostPage](ctx, r.Proxy, "post-service", rest.Request{Method: "GET", Path: "/api/v1/post", Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) Post(ctx context.Context, id string) (*model.PostPost, error) {
	return restCall[*model.PostPost](ctx, r.Proxy, "post-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/post/%v", id)}, "")
}

func (r *queryResolver) PostCategories(ctx context.Context, cursor *string, limit *int) (*model.PostCategoryPage, error) {
	return restCall[*model.PostCategoryPage](ctx, r.Proxy, "post-service", rest.Request{Method: "GET", Path: "/api/v1/category", Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) PostCategory(ctx context.Context, id string) (*model.PostCategory, error) {
	return restCall[*model.PostCategory](ctx, r.Proxy, "post-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/category/%v", id)}, "")
}
