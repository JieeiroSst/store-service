package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) WorkflowProcessOrder(ctx context.Context, id string) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "workflow-service", rest.Request{Method: "POST", Path: rest.P("/orders/%v/process", id), Body: map[string]any{}}, "message")
}

func (r *mutationResolver) WorkflowStartOrderWorkflow(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "workflow-service", rest.Request{Method: "POST", Path: rest.P("/orders/%v/workflow", id), Body: map[string]any{}}, "")
}
