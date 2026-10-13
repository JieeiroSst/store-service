package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) GymCreateClient(ctx context.Context, input model.GymClientInput) (*model.GymClient, error) {
	return restCall[*model.GymClient](ctx, r.Proxy, "identified-service", rest.Request{Method: "POST", Path: "/clients", Body: input}, "")
}

func (r *mutationResolver) GymUpdateClient(ctx context.Context, clientID string, input model.GymClientInput) (*model.GymClient, error) {
	return restCall[*model.GymClient](ctx, r.Proxy, "identified-service", rest.Request{Method: "PUT", Path: rest.P("/clients/%v", clientID), Body: input}, "")
}

func (r *mutationResolver) GymArchiveClient(ctx context.Context, clientID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "identified-service", rest.Request{Method: "DELETE", Path: rest.P("/clients/%v", clientID)}, "")
}

func (r *mutationResolver) GymExportClients(ctx context.Context, format string) (any, error) {
	return restCall[any](ctx, r.Proxy, "identified-service", rest.Request{Method: "POST", Path: "/clients/exports", Body: map[string]any{"format": format}}, "")
}

func (r *mutationResolver) GymCreatePass(ctx context.Context, validity string) (*model.GymPass, error) {
	return restCall[*model.GymPass](ctx, r.Proxy, "identified-service", rest.Request{Method: "POST", Path: "/gym-passes", Body: withField(map[string]any{"validity": validity}, "owner", userIDOf(ctx))}, "")
}

func (r *mutationResolver) GymPausePass(ctx context.Context, gymPassID string, days int) (*model.GymPass, error) {
	return restCall[*model.GymPass](ctx, r.Proxy, "identified-service", rest.Request{Method: "POST", Path: rest.P("/gym-passes/%v/pauses", gymPassID), Body: map[string]any{"days": days}}, "")
}

func (r *mutationResolver) GymRenewPass(ctx context.Context, gymPassID string) (*model.GymPass, error) {
	return restCall[*model.GymPass](ctx, r.Proxy, "identified-service", rest.Request{Method: "PUT", Path: rest.P("/gym-passes/%v/renewal", gymPassID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) GymCreateClass(ctx context.Context, input model.GymClassInput) (*model.GymClass, error) {
	return restCall[*model.GymClass](ctx, r.Proxy, "identified-service", rest.Request{Method: "POST", Path: "/classes", Body: input}, "")
}

func (r *mutationResolver) GymUpdateClass(ctx context.Context, classID string, input model.GymClassInput) (*model.GymClass, error) {
	return restCall[*model.GymClass](ctx, r.Proxy, "identified-service", rest.Request{Method: "PUT", Path: rest.P("/classes/%v", classID), Body: input}, "")
}

func (r *mutationResolver) GymDeleteClass(ctx context.Context, classID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "identified-service", rest.Request{Method: "DELETE", Path: rest.P("/classes/%v", classID)}, "")
}

func (r *queryResolver) GymClasses(ctx context.Context) ([]*model.GymClass, error) {
	return restCall[[]*model.GymClass](ctx, r.Proxy, "identified-service", rest.Request{Method: "GET", Path: "/classes"}, "")
}

func (r *queryResolver) GymClass(ctx context.Context, classID string) (*model.GymClass, error) {
	return restCall[*model.GymClass](ctx, r.Proxy, "identified-service", rest.Request{Method: "GET", Path: rest.P("/classes/%v", classID)}, "")
}

func (r *queryResolver) GymPassVerification(ctx context.Context, gymPassID string) (*model.GymPass, error) {
	return restCall[*model.GymPass](ctx, r.Proxy, "identified-service", rest.Request{Method: "GET", Path: rest.P("/gym-passes/%v/verification", gymPassID)}, "")
}
