package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) PartnerCreatePartner(ctx context.Context, input model.PartnerPartnerInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "POST", Path: "/v1/partner/", Query: restQuery("user_id", userIDOf(ctx)), Body: input}, "")
}

func (r *mutationResolver) PartnerUpdatePartner(ctx context.Context, id string, input model.PartnerPartnerInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "PUT", Path: "/v1/partner/", Query: restQuery("id", id), Body: input}, "")
}

func (r *mutationResolver) PartnerDeletePartner(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "DELETE", Path: "/v1/partner/", Query: restQuery("id", id)}, "")
}

func (r *mutationResolver) PartnerCreatePartnership(ctx context.Context, input model.PartnerPartnershipInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "POST", Path: "/v1/partnership/", Query: restQuery("user_id", userIDOf(ctx)), Body: input}, "")
}

func (r *mutationResolver) PartnerUpdatePartnership(ctx context.Context, id string, input model.PartnerPartnershipInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "PUT", Path: "/v1/partnership/", Query: restQuery("id", id), Body: input}, "")
}

func (r *mutationResolver) PartnerDeletePartnership(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "DELETE", Path: "/v1/partnership/", Query: restQuery("id", id)}, "")
}

func (r *mutationResolver) PartnerCreatePartnershipsPartner(ctx context.Context, input model.PartnerPartnershipsPartnerInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "POST", Path: "/v1/partnerships-partner/", Query: restQuery("user_id", userIDOf(ctx)), Body: input}, "")
}

func (r *mutationResolver) PartnerUpdatePartnershipsPartner(ctx context.Context, id string, input model.PartnerPartnershipsPartnerInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "PUT", Path: "/v1/partnerships-partner/", Query: restQuery("id", id), Body: input}, "")
}

func (r *mutationResolver) PartnerDeletePartnershipsPartner(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "DELETE", Path: "/v1/partnerships-partner/", Query: restQuery("id", id)}, "")
}

func (r *mutationResolver) PartnerCreateProject(ctx context.Context, input model.PartnerProjectInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "POST", Path: "/v1/project/", Query: restQuery("user_id", userIDOf(ctx)), Body: input}, "")
}

func (r *mutationResolver) PartnerUpdateProject(ctx context.Context, id string, input model.PartnerProjectInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "PUT", Path: "/v1/project/", Query: restQuery("id", id), Body: input}, "")
}

func (r *mutationResolver) PartnerDeleteProject(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "DELETE", Path: "/v1/project/", Query: restQuery("id", id)}, "")
}

func (r *mutationResolver) PartnerUpdateStatus(ctx context.Context, id string, status string) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "PATCH", Path: "/v1/partner/status", Query: restQuery("id", id), Body: map[string]any{"status": status}}, "")
}

func (r *mutationResolver) PartnerRestore(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "PATCH", Path: "/v1/partner/restore", Query: restQuery("id", id)}, "")
}

func (r *mutationResolver) PartnerAdjustScore(ctx context.Context, id string, points int) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "PATCH", Path: "/v1/partner/score", Query: restQuery("id", id), Body: map[string]any{"points": points}}, "")
}

func (r *mutationResolver) PartnerCloseInactive(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "partner-service", rest.Request{Method: "POST", Path: "/v1/partner/close-inactive", Body: map[string]any{}}, "")
}

func (r *queryResolver) Partners(ctx context.Context, limit *int, page *int, sort *string) (*model.PartnerPartnerPage, error) {
	return restCall[*model.PartnerPartnerPage](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: "/v1/partner/", Query: restQuery("limit", limit, "page", page, "sort", sort)}, "")
}

func (r *queryResolver) PartnerSearchPartners(ctx context.Context, limit *int, page *int, sort *string, typeArg *string, status *string, name *string, userID *string) (*model.PartnerPartnerPage, error) {
	return restCall[*model.PartnerPartnerPage](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: "/v1/partner/search", Query: restQuery("limit", limit, "page", page, "sort", sort, "type", typeArg, "status", status, "name", name, "user_id", userID)}, "")
}

func (r *queryResolver) PartnerStatistics(ctx context.Context) (*model.PartnerPartnerStatistics, error) {
	return restCall[*model.PartnerPartnerStatistics](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: "/v1/partner/statistics"}, "")
}

func (r *queryResolver) PartnerActivity(ctx context.Context, id string) (*model.PartnerPartnerActivity, error) {
	return restCall[*model.PartnerPartnerActivity](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: "/v1/partner/active", Query: restQuery("id", id)}, "")
}

func (r *queryResolver) Partner(ctx context.Context, id string) (*model.PartnerPartner, error) {
	return restCall[*model.PartnerPartner](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: rest.P("/v1/partner/%v", id)}, "")
}

func (r *queryResolver) Partnership(ctx context.Context, id string) (*model.PartnerPartnership, error) {
	return restCall[*model.PartnerPartnership](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: rest.P("/v1/partnership/%v", id)}, "")
}

func (r *queryResolver) Partnerships(ctx context.Context, limit *int, page *int, sort *string) (*model.PartnerPartnershipPage, error) {
	return restCall[*model.PartnerPartnershipPage](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: "/v1/partnership/", Query: restQuery("limit", limit, "page", page, "sort", sort)}, "")
}

func (r *queryResolver) PartnershipsPartner(ctx context.Context, id string) (*model.PartnerPartnershipsPartner, error) {
	return restCall[*model.PartnerPartnershipsPartner](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: rest.P("/v1/partnerships-partner/%v", id)}, "")
}

func (r *queryResolver) PartnershipsPartners(ctx context.Context, limit *int, page *int, sort *string) (*model.PartnerPartnershipsPartnerPage, error) {
	return restCall[*model.PartnerPartnershipsPartnerPage](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: "/v1/partnerships-partner/", Query: restQuery("limit", limit, "page", page, "sort", sort)}, "")
}

func (r *queryResolver) PartnerProjects(ctx context.Context, limit *int, page *int, sort *string) (*model.PartnerProjectPage, error) {
	return restCall[*model.PartnerProjectPage](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: "/v1/project/", Query: restQuery("limit", limit, "page", page, "sort", sort)}, "")
}

func (r *queryResolver) PartnerProject(ctx context.Context, id string) (*model.PartnerProject, error) {
	return restCall[*model.PartnerProject](ctx, r.Proxy, "partner-service", rest.Request{Method: "GET", Path: rest.P("/v1/project/%v", id)}, "")
}
