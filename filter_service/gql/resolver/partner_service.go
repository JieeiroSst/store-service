package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) PartnerQuery() generated.PartnerQueryResolver { return &partnerQueryResolver{r} }

type partnerQueryResolver struct{ *Resolver }

func (r *partnerQueryResolver) Partners(ctx context.Context, obj *model.PartnerQuery, limit *int, page *int, sort *string) (*model.PartnerPartnerPage, error) {
	return r.Clients.PartnerService.Partners(ctx, limit, page, sort)
}

func (r *partnerQueryResolver) SearchPartners(ctx context.Context, obj *model.PartnerQuery, limit *int, page *int, sort *string, typeArg *string, status *string, name *string, userID *string) (*model.PartnerPartnerPage, error) {
	return r.Clients.PartnerService.SearchPartners(ctx, limit, page, sort, typeArg, status, name, userID)
}

func (r *partnerQueryResolver) Statistics(ctx context.Context, obj *model.PartnerQuery) (*model.PartnerPartnerStatistics, error) {
	return r.Clients.PartnerService.Statistics(ctx)
}

func (r *partnerQueryResolver) Activity(ctx context.Context, obj *model.PartnerQuery, id string) (*model.PartnerPartnerActivity, error) {
	return r.Clients.PartnerService.Activity(ctx, id)
}

func (r *partnerQueryResolver) Partner(ctx context.Context, obj *model.PartnerQuery, id string) (*model.PartnerPartner, error) {
	return r.Clients.PartnerService.Partner(ctx, id)
}

func (r *partnerQueryResolver) Partnership(ctx context.Context, obj *model.PartnerQuery, id string) (*model.PartnerPartnership, error) {
	return r.Clients.PartnerService.Partnership(ctx, id)
}

func (r *partnerQueryResolver) Partnerships(ctx context.Context, obj *model.PartnerQuery, limit *int, page *int, sort *string) (*model.PartnerPartnershipPage, error) {
	return r.Clients.PartnerService.Partnerships(ctx, limit, page, sort)
}

func (r *partnerQueryResolver) PartnershipsPartner(ctx context.Context, obj *model.PartnerQuery, id string) (*model.PartnerPartnershipsPartner, error) {
	return r.Clients.PartnerService.PartnershipsPartner(ctx, id)
}

func (r *partnerQueryResolver) PartnershipsPartners(ctx context.Context, obj *model.PartnerQuery, limit *int, page *int, sort *string) (*model.PartnerPartnershipsPartnerPage, error) {
	return r.Clients.PartnerService.PartnershipsPartners(ctx, limit, page, sort)
}

func (r *partnerQueryResolver) Projects(ctx context.Context, obj *model.PartnerQuery, limit *int, page *int, sort *string) (*model.PartnerProjectPage, error) {
	return r.Clients.PartnerService.Projects(ctx, limit, page, sort)
}

func (r *partnerQueryResolver) Project(ctx context.Context, obj *model.PartnerQuery, id string) (*model.PartnerProject, error) {
	return r.Clients.PartnerService.Project(ctx, id)
}
