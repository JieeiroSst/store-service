package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) QueuePublish(ctx context.Context, topic string, payload any) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "message-service", rest.Request{Method: "POST", Path: "/api/v1/queue/pub", Query: restQuery("topic", topic), Body: payload}, "")
}

func (r *mutationResolver) QueueConsume(ctx context.Context, topic string) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "message-service", rest.Request{Method: "POST", Path: "/api/v1/queue/sub", Query: restQuery("topic", topic)}, "")
}
