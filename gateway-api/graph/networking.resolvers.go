package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) NetPutAgentServiceRegister(ctx context.Context, query map[string]any, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: "/v1/agent/service/register", Query: restQuery("query", query), Body: input}, "")
}

func (r *mutationResolver) NetPutAgentServiceDeregister(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/agent/service/deregister/%v", id), Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutAgentServiceMaintenance(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/agent/service/maintenance/%v", id), Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutAgentMaintenance(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: "/v1/agent/maintenance", Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutAgentCheckRegister(ctx context.Context, query map[string]any, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: "/v1/agent/check/register", Query: restQuery("query", query), Body: input}, "")
}

func (r *mutationResolver) NetPutAgentCheckDeregister(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/agent/check/deregister/%v", id), Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutAgentCheckPass(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/agent/check/pass/%v", id), Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutAgentCheckWarn(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/agent/check/warn/%v", id), Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutAgentCheckFail(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/agent/check/fail/%v", id), Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutAgentCheckUpdate(ctx context.Context, id string, query map[string]any, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/agent/check/update/%v", id), Query: restQuery("query", query), Body: input}, "")
}

func (r *mutationResolver) NetPutCatalogRegister(ctx context.Context, query map[string]any, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: "/v1/catalog/register", Query: restQuery("query", query), Body: input}, "")
}

func (r *mutationResolver) NetPutCatalogDeregister(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: "/v1/catalog/deregister", Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutKv(ctx context.Context, key string, value string, query map[string]any) (any, error) {
	return restText[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/kv/%s", rest.Raw(key)), Query: restQuery("query", query)}, value, "")
}

func (r *mutationResolver) NetDeleteKv(ctx context.Context, key string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "DELETE", Path: rest.P("/v1/kv/%s", rest.Raw(key)), Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutSessionCreate(ctx context.Context, query map[string]any, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: "/v1/session/create", Query: restQuery("query", query), Body: input}, "")
}

func (r *mutationResolver) NetPutSessionDestroy(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/session/destroy/%v", id), Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutSessionRenew(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/session/renew/%v", id), Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPostConnectIntentions(ctx context.Context, query map[string]any, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "POST", Path: "/v1/connect/intentions", Query: restQuery("query", query), Body: input}, "")
}

func (r *mutationResolver) NetPutConnectIntentionsExact(ctx context.Context, query map[string]any, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: "/v1/connect/intentions/exact", Query: restQuery("query", query), Body: input}, "")
}

func (r *mutationResolver) NetDeleteConnectIntentionsExact(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "DELETE", Path: "/v1/connect/intentions/exact", Query: restQuery("query", query)}, "")
}

func (r *mutationResolver) NetPutConnectIntentions(ctx context.Context, id string, query map[string]any, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "PUT", Path: rest.P("/v1/connect/intentions/%v", id), Query: restQuery("query", query), Body: input}, "")
}

func (r *mutationResolver) NetDeleteConnectIntentions(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "DELETE", Path: rest.P("/v1/connect/intentions/%v", id), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetAgentSelf(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/agent/self", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetAgentMembers(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/agent/members", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetAgentServices(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/agent/services", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetAgentService(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/agent/service/%v", id), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetAgentChecks(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/agent/checks", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetAgentHealthServiceName(ctx context.Context, name string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/agent/health/service/name/%v", name), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetCatalogDatacenters(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/catalog/datacenters", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetCatalogNodes(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/catalog/nodes", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetCatalogNode(ctx context.Context, node string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/catalog/node/%v", node), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetCatalogServices(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/catalog/services", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetCatalogService(ctx context.Context, service string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/catalog/service/%v", service), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetHealthService(ctx context.Context, service string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/health/service/%v", service), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetHealthChecks(ctx context.Context, service string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/health/checks/%v", service), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetHealthNode(ctx context.Context, node string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/health/node/%v", node), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetHealthState(ctx context.Context, state string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/health/state/%v", state), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetKv(ctx context.Context, key string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/kv/%s", rest.Raw(key)), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetSessionInfo(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/session/info/%v", id), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetSessionList(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/session/list", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetSessionNode(ctx context.Context, node string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/session/node/%v", node), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetConnectIntentions(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/connect/intentions", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetConnectIntentionsExact(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/connect/intentions/exact", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetConnectIntentionsCheck(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/connect/intentions/check", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetConnectIntentionsMatch(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/connect/intentions/match", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetConnectIntentions2(ctx context.Context, id string, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: rest.P("/v1/connect/intentions/%v", id), Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetStatusLeader(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/status/leader", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetStatusPeers(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/status/peers", Query: restQuery("query", query)}, "")
}

func (r *queryResolver) NetInternalUIServices(ctx context.Context, query map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "networking-service", rest.Request{Method: "GET", Path: "/v1/internal/ui/services", Query: restQuery("query", query)}, "")
}
