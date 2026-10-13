package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) AiAgentGenerate(ctx context.Context, input model.AiAgentGenerateInput) (*model.AiAgentGenerateResponse, error) {
	return restCall[*model.AiAgentGenerateResponse](ctx, r.Proxy, "ai-agent-system", rest.Request{Method: "POST", Path: "/api/generate", Body: mergeBody(input, "{\"stream\":false}")}, "")
}

func (r *mutationResolver) AiAgentChat(ctx context.Context, input model.AiAgentChatInput) (*model.AiAgentChatResponse, error) {
	return restCall[*model.AiAgentChatResponse](ctx, r.Proxy, "ai-agent-system", rest.Request{Method: "POST", Path: "/api/chat", Body: mergeBody(input, "{\"stream\":false}")}, "")
}

func (r *queryResolver) AiAgentChatHistory(ctx context.Context, limit *int) (*model.AiAgentHistory, error) {
	return restCall[*model.AiAgentHistory](ctx, r.Proxy, "ai-agent-system", rest.Request{Method: "GET", Path: "/v1/chat/history", Query: restQuery("limit", limit)}, "")
}

func (r *queryResolver) AiAgentModels(ctx context.Context) (*model.AiAgentTags, error) {
	return restCall[*model.AiAgentTags](ctx, r.Proxy, "ai-agent-system", rest.Request{Method: "GET", Path: "/api/tags"}, "")
}

func (r *queryResolver) AiAgentShowModel(ctx context.Context, input model.AiAgentShowInput) (*model.AiAgentShowResponse, error) {
	return restCall[*model.AiAgentShowResponse](ctx, r.Proxy, "ai-agent-system", rest.Request{Method: "POST", Path: "/api/show", Body: input}, "")
}
