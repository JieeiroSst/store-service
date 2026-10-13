package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) RecompCreateRecompense(ctx context.Context, input model.RecompRecompenseInput) (*model.RecompRecompense, error) {
	return restCall[*model.RecompRecompense](ctx, r.Proxy, "recompense-service", rest.Request{Method: "POST", Path: "/api/recompenses", Body: input}, "")
}

func (r *mutationResolver) RecompUpdateRecompense(ctx context.Context, id int, input model.RecompRecompenseInput) (*model.RecompRecompense, error) {
	return restCall[*model.RecompRecompense](ctx, r.Proxy, "recompense-service", rest.Request{Method: "PUT", Path: rest.P("/api/recompenses/%v", id), Body: input}, "")
}

func (r *mutationResolver) RecompDeleteRecompense(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "recompense-service", rest.Request{Method: "DELETE", Path: rest.P("/api/recompenses/%v", id)}, "")
}

func (r *mutationResolver) RecompEarnPoints(ctx context.Context, memberID int, amount float64, idempotencyKey string) (*model.RecompMemberStatus, error) {
	return restCall[*model.RecompMemberStatus](ctx, r.Proxy, "recompense-service", rest.Request{Method: "POST", Path: rest.P("/api/members/%v/points/earn", memberID), Body: map[string]any{"amount": amount, "idempotencyKey": idempotencyKey}}, "")
}

func (r *mutationResolver) RecompRedeem(ctx context.Context, id int, memberID int, amount float64) (*model.RecompRedemption, error) {
	return restCall[*model.RecompRedemption](ctx, r.Proxy, "recompense-service", rest.Request{Method: "POST", Path: rest.P("/api/recompenses/%v/redeem", id), Body: map[string]any{"memberId": memberID, "amount": amount}}, "")
}

func (r *mutationResolver) RedeemMyRecompense(ctx context.Context, id int, amount float64) (*model.RecompRedemption, error) {
	return restCall[*model.RecompRedemption](ctx, r.Proxy, "recompense-service", rest.Request{Method: "POST", Path: rest.P("/api/recompenses/%v/redeem", id), Body: withField(map[string]any{"amount": amount}, "memberId", userIDInt(ctx))}, "")
}

func (r *queryResolver) Recompenses(ctx context.Context, page *int, limit *int) ([]*model.RecompRecompense, error) {
	return restCall[[]*model.RecompRecompense](ctx, r.Proxy, "recompense-service", rest.Request{Method: "GET", Path: "/api/recompenses", Query: restQuery("page", page, "limit", limit)}, "")
}

func (r *queryResolver) Recompense(ctx context.Context, id string) (*model.RecompRecompense, error) {
	return restCall[*model.RecompRecompense](ctx, r.Proxy, "recompense-service", rest.Request{Method: "GET", Path: rest.P("/api/recompenses/%v", id)}, "")
}

func (r *queryResolver) RecompMemberRecompenses(ctx context.Context, memberID int, page *int, limit *int) ([]*model.RecompRecompense, error) {
	return restCall[[]*model.RecompRecompense](ctx, r.Proxy, "recompense-service", rest.Request{Method: "GET", Path: rest.P("/api/recompenses/member/%v", memberID), Query: restQuery("page", page, "limit", limit)}, "")
}

func (r *queryResolver) RecompMemberPoints(ctx context.Context, memberID int) (*model.RecompMemberStatus, error) {
	return restCall[*model.RecompMemberStatus](ctx, r.Proxy, "recompense-service", rest.Request{Method: "GET", Path: rest.P("/api/members/%v/points", memberID)}, "")
}

func (r *queryResolver) RecompBestRecompense(ctx context.Context, memberID int, amount float64) (*model.RecompRedemption, error) {
	return restCall[*model.RecompRedemption](ctx, r.Proxy, "recompense-service", rest.Request{Method: "GET", Path: rest.P("/api/members/%v/recompenses/best", memberID), Query: restQuery("amount", amount)}, "")
}

func (r *queryResolver) MyRecompensePoints(ctx context.Context) (*model.RecompMemberStatus, error) {
	return restCall[*model.RecompMemberStatus](ctx, r.Proxy, "recompense-service", rest.Request{Method: "GET", Path: rest.P("/api/members/%v/points", userIDOf(ctx))}, "")
}

func (r *queryResolver) MyBestRecompense(ctx context.Context, amount float64) (*model.RecompRecompense, error) {
	return restCall[*model.RecompRecompense](ctx, r.Proxy, "recompense-service", rest.Request{Method: "GET", Path: rest.P("/api/members/%v/recompenses/best", userIDOf(ctx)), Query: restQuery("amount", amount)}, "")
}

func (r *queryResolver) MyRecompenses(ctx context.Context, page *int, limit *int) (any, error) {
	return restCall[any](ctx, r.Proxy, "recompense-service", rest.Request{Method: "GET", Path: rest.P("/api/recompenses/member/%v", userIDOf(ctx)), Query: restQuery("page", page, "limit", limit)}, "")
}
