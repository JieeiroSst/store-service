package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ShareAsk(ctx context.Context, content string) (*model.ShareAnswer, error) {
	return restCall[*model.ShareAnswer](ctx, r.Proxy, "share-service", rest.Request{Method: "POST", Path: "/", Body: map[string]any{"content": content}}, "")
}
