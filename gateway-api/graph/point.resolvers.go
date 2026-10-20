package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) PointCreateRewardPoint(ctx context.Context, input model.PointRewardPointInput) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "point-service", rest.Request{Method: "POST", Path: "/api/v1/reward-points", Body: input}, "")
}

func (r *mutationResolver) PointUpdateRewardPoint(ctx context.Context, input model.PointRewardPointInput) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "point-service", rest.Request{Method: "PUT", Path: "/api/v1/reward-points", Body: input}, "")
}

func (r *mutationResolver) PointCreateRewardDiscount(ctx context.Context, input model.PointRewardDiscountInput) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "point-service", rest.Request{Method: "POST", Path: "/api/v1/reward-discounts", Body: input}, "")
}

func (r *mutationResolver) PointUpdateRewardDiscount(ctx context.Context, input model.PointRewardDiscountInput) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "point-service", rest.Request{Method: "PUT", Path: "/api/v1/reward-discounts", Body: input}, "")
}

func (r *mutationResolver) PointCreateConvertedRewardPoint(ctx context.Context, input model.PointConvertedRewardPointInput) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "point-service", rest.Request{Method: "POST", Path: "/api/v1/converted-reward-points", Body: input}, "")
}

func (r *mutationResolver) PointUpdateConvertedRewardPoint(ctx context.Context, input model.PointConvertedRewardPointInput) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "point-service", rest.Request{Method: "PUT", Path: "/api/v1/converted-reward-points", Body: input}, "")
}

func (r *queryResolver) PointMyBalance(ctx context.Context) (*model.PointAccount, error) {
	return restCall[*model.PointAccount](ctx, r.Proxy, "point-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/points", userIDOf(ctx))}, "data")
}

func (r *queryResolver) PointMyTransactions(ctx context.Context, limit *int, offset *int) (*model.PointTransactionPage, error) {
	return restCall[*model.PointTransactionPage](ctx, r.Proxy, "point-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/point-transactions", userIDOf(ctx)), Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) PointUserBalance(ctx context.Context, userID string) (*model.PointAccount, error) {
	return restCall[*model.PointAccount](ctx, r.Proxy, "point-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/points", userID)}, "data")
}

func (r *queryResolver) PointUserTransactions(ctx context.Context, userID string, limit *int, offset *int) (*model.PointTransactionPage, error) {
	return restCall[*model.PointTransactionPage](ctx, r.Proxy, "point-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/point-transactions", userID), Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) PointRewardPoints(ctx context.Context, perPage *int, sortOrder *string, cursor *string) (*model.PointRewardPointPage, error) {
	return restCall[*model.PointRewardPointPage](ctx, r.Proxy, "point-service", rest.Request{Method: "GET", Path: "/api/v1/reward-points", Query: restQuery("per_page", perPage, "sort_order", sortOrder, "cursor", cursor)}, "")
}

func (r *queryResolver) PointRewardPoint(ctx context.Context, id string) (*model.PointRewardPoint, error) {
	return restCall[*model.PointRewardPoint](ctx, r.Proxy, "point-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/reward-points/%v", id)}, "")
}

func (r *queryResolver) PointRewardDiscounts(ctx context.Context, perPage *int, sortOrder *string, cursor *string) (*model.PointRewardDiscountPage, error) {
	return restCall[*model.PointRewardDiscountPage](ctx, r.Proxy, "point-service", rest.Request{Method: "GET", Path: "/api/v1/reward-discounts", Query: restQuery("per_page", perPage, "sort_order", sortOrder, "cursor", cursor)}, "")
}

func (r *queryResolver) PointRewardDiscount(ctx context.Context, id string) (*model.PointRewardDiscount, error) {
	return restCall[*model.PointRewardDiscount](ctx, r.Proxy, "point-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/reward-discounts/%v", id)}, "")
}

func (r *queryResolver) PointConvertedRewardPoints(ctx context.Context, perPage *int, sortOrder *string, cursor *string) (*model.PointConvertedRewardPointPage, error) {
	return restCall[*model.PointConvertedRewardPointPage](ctx, r.Proxy, "point-service", rest.Request{Method: "GET", Path: "/api/v1/converted-reward-points", Query: restQuery("per_page", perPage, "sort_order", sortOrder, "cursor", cursor)}, "")
}

func (r *queryResolver) PointConvertedRewardPoint(ctx context.Context, id string) (*model.PointConvertedRewardPoint, error) {
	return restCall[*model.PointConvertedRewardPoint](ctx, r.Proxy, "point-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/converted-reward-points/%v", id)}, "")
}
