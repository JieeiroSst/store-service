package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) OllamaPyGenerate(ctx context.Context, input model.OllamaPyGenerateInput) (*model.OllamaPyGenerate, error) {
	return restCall[*model.OllamaPyGenerate](ctx, r.Proxy, "ollama-model-py", rest.Request{Method: "POST", Path: "/api/generate", Body: withField(mergeBody(input, "{\"stream\":false}"), "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) OllamaPyChat(ctx context.Context, input model.OllamaPyChatInput) (*model.OllamaPyChat, error) {
	return restCall[*model.OllamaPyChat](ctx, r.Proxy, "ollama-model-py", rest.Request{Method: "POST", Path: "/api/chat", Body: withField(mergeBody(input, "{\"stream\":false}"), "user_id", userIDOf(ctx))}, "")
}

func (r *queryResolver) OllamaPyTags(ctx context.Context) (*model.OllamaPyTags, error) {
	return restCall[*model.OllamaPyTags](ctx, r.Proxy, "ollama-model-py", rest.Request{Method: "GET", Path: "/api/tags"}, "")
}

func (r *queryResolver) OllamaPyShow(ctx context.Context, input model.OllamaPyShowInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ollama-model-py", rest.Request{Method: "POST", Path: "/api/show", Body: input}, "")
}
