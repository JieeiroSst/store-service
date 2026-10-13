package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) BotCreatePost(ctx context.Context, input model.BotCreatePostInput) (*model.BotPost, error) {
	return restCall[*model.BotPost](ctx, r.Proxy, "bot-service", rest.Request{Method: "POST", Path: "/bot-content/v1/posts", Body: withField(input, "created_by", userIDOf(ctx))}, "")
}

func (r *mutationResolver) BotUpdatePost(ctx context.Context, id string, input model.BotUpdatePostInput) (*model.BotPost, error) {
	return restCall[*model.BotPost](ctx, r.Proxy, "bot-service", rest.Request{Method: "PATCH", Path: rest.P("/bot-content/v1/posts/%v", id), Body: input}, "")
}

func (r *mutationResolver) BotSubmitPost(ctx context.Context, id string) (*model.BotPost, error) {
	return restCall[*model.BotPost](ctx, r.Proxy, "bot-service", rest.Request{Method: "POST", Path: rest.P("/bot-content/v1/posts/%v/submit", id), Body: withField(map[string]any{}, "changed_by", userIDOf(ctx))}, "")
}

func (r *mutationResolver) BotApprovePost(ctx context.Context, id string) (*model.BotPost, error) {
	return restCall[*model.BotPost](ctx, r.Proxy, "bot-service", rest.Request{Method: "POST", Path: rest.P("/bot-content/v1/posts/%v/approve", id), Body: withField(map[string]any{}, "changed_by", userIDOf(ctx))}, "")
}

func (r *mutationResolver) BotPublishPost(ctx context.Context, id string) (*model.BotPost, error) {
	return restCall[*model.BotPost](ctx, r.Proxy, "bot-service", rest.Request{Method: "POST", Path: rest.P("/bot-content/v1/posts/%v/publish", id), Body: withField(map[string]any{}, "changed_by", userIDOf(ctx))}, "")
}

func (r *mutationResolver) BotCancelPost(ctx context.Context, id string) (*model.BotPost, error) {
	return restCall[*model.BotPost](ctx, r.Proxy, "bot-service", rest.Request{Method: "POST", Path: rest.P("/bot-content/v1/posts/%v/cancel", id), Body: withField(map[string]any{}, "changed_by", userIDOf(ctx))}, "")
}

func (r *mutationResolver) BotRejectPost(ctx context.Context, id string, reason string) (*model.BotPost, error) {
	return restCall[*model.BotPost](ctx, r.Proxy, "bot-service", rest.Request{Method: "POST", Path: rest.P("/bot-content/v1/posts/%v/reject", id), Body: withField(map[string]any{"reason": reason}, "changed_by", userIDOf(ctx))}, "")
}

func (r *queryResolver) BotPosts(ctx context.Context, limit *int, offset *int, status *string, campaign *string) (*model.BotPostList, error) {
	return restCall[*model.BotPostList](ctx, r.Proxy, "bot-service", rest.Request{Method: "GET", Path: "/bot-content/v1/posts", Query: restQuery("limit", limit, "offset", offset, "status", status, "campaign", campaign)}, "")
}

func (r *queryResolver) BotPost(ctx context.Context, id string) (*model.BotPost, error) {
	return restCall[*model.BotPost](ctx, r.Proxy, "bot-service", rest.Request{Method: "GET", Path: rest.P("/bot-content/v1/posts/%v", id)}, "")
}
