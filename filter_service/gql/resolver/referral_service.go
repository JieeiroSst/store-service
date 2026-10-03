package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ReferralQuery() generated.ReferralQueryResolver { return &referralQueryResolver{r} }

type referralQueryResolver struct{ *Resolver }

func (r *referralQueryResolver) Link(ctx context.Context, obj *model.ReferralQuery, refCode string) (*model.ReferralReferralLink, error) {
	return r.Clients.ReferralService.Link(ctx, refCode)
}

func (r *referralQueryResolver) UserLinks(ctx context.Context, obj *model.ReferralQuery, userID string, limit *int, cursor *string) (*model.ReferralLinkList, error) {
	return r.Clients.ReferralService.UserLinks(ctx, userID, limit, cursor)
}

func (r *referralQueryResolver) Status(ctx context.Context, obj *model.ReferralQuery, refCode string) (*model.ReferralReferralStatusResponse, error) {
	return r.Clients.ReferralService.Status(ctx, refCode)
}

func (r *referralQueryResolver) UserStats(ctx context.Context, obj *model.ReferralQuery, userID string) (*model.ReferralUserReferralStats, error) {
	return r.Clients.ReferralService.UserStats(ctx, userID)
}

func (r *referralQueryResolver) ShareTargets(ctx context.Context, obj *model.ReferralQuery, refCode string) ([]*model.ReferralShareTarget, error) {
	return r.Clients.ReferralService.ShareTargets(ctx, refCode)
}

func (r *referralQueryResolver) ActiveRewardProgram(ctx context.Context, obj *model.ReferralQuery) (*model.ReferralRewardProgram, error) {
	return r.Clients.ReferralService.ActiveRewardProgram(ctx)
}
