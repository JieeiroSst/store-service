package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) PolyCreateEvent(ctx context.Context, input model.PolyEventInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: "/api/v1/events", Body: input}, "")
}

func (r *mutationResolver) PolyAddComment(ctx context.Context, id string, body string, parentID *int) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/comments", id), Body: withField(map[string]any{"body": body, "parent_id": parentID}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolyConvertPositions(ctx context.Context, id string, marketIds []int, amount int) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/convert", id), Body: withField(map[string]any{"market_ids": marketIds, "amount": amount}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolyProposeEvent(ctx context.Context, id string, winnerMarketID int) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/propose", id), Body: map[string]any{"winner_market_id": winnerMarketID}}, "")
}

func (r *mutationResolver) PolyResolveEvent(ctx context.Context, id string, winnerMarketID int) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/resolve", id), Body: map[string]any{"winner_market_id": winnerMarketID}}, "")
}

func (r *mutationResolver) PolyWatchEvent(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/watch", id), Body: withField(map[string]any{}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolyUnwatchEvent(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/events/%v/watch", id), Query: restQuery("user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolyDeleteComment(ctx context.Context, commentID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/comments/%v", commentID), Query: restQuery("user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolyLikeComment(ctx context.Context, commentID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/comments/%v/like", commentID), Body: withField(map[string]any{}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolyUnlikeComment(ctx context.Context, commentID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/comments/%v/like", commentID), Query: restQuery("user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolySetMarketRewards(ctx context.Context, id string, dailyPool int, maxSpread int, minSize int) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/markets/%v/rewards", id), Body: map[string]any{"daily_pool": dailyPool, "max_spread": maxSpread, "min_size": minSize}}, "")
}

func (r *mutationResolver) PolyPlaceOrder(ctx context.Context, id string, input model.PolyOrderInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/markets/%v/orders", id), Body: withField(input, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolyCancelAllOrders(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/markets/%v/orders", id), Query: restQuery("user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolyProposeResolution(ctx context.Context, id string, outcome string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/markets/%v/propose", id), Body: map[string]any{"outcome": outcome}}, "")
}

func (r *mutationResolver) PolyDisputeResolution(ctx context.Context, id string, reason string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/markets/%v/dispute", id), Body: withField(map[string]any{"reason": reason}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolyFinalizeResolution(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/markets/%v/finalize", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) PolyResolveMarket(ctx context.Context, id string, outcome string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/markets/%v/resolve", id), Body: map[string]any{"outcome": outcome}}, "")
}

func (r *mutationResolver) PolyFundExchange(ctx context.Context, userID string, amount int) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: "/api/v1/admin/exchange/fund", Body: map[string]any{"user_id": userID, "amount": amount}}, "")
}

func (r *mutationResolver) PolyCancelOrder(ctx context.Context, orderID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/orders/%v", orderID), Query: restQuery("user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) PolyUpdateMyProfile(ctx context.Context, username *string, bio *string, avatarURL *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/users/%v/profile", userIDOf(ctx)), Body: map[string]any{"username": username, "bio": bio, "avatar_url": avatarURL}}, "")
}

func (r *mutationResolver) PolyDeposit(ctx context.Context, amount int) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/users/%v/deposit", userIDOf(ctx)), Body: map[string]any{"amount": amount}}, "")
}

func (r *mutationResolver) PolyWithdraw(ctx context.Context, amount int) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/users/%v/withdraw", userIDOf(ctx)), Body: map[string]any{"amount": amount}}, "")
}

func (r *mutationResolver) PolyRedeemReferral(ctx context.Context, refCode string) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/users/%v/referral/redeem", userIDOf(ctx)), Body: map[string]any{"ref_code": refCode}}, "")
}

func (r *queryResolver) PolyCategories(ctx context.Context) ([]*model.PolyCategoryCount, error) {
	return restCall[[]*model.PolyCategoryCount](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: "/api/v1/categories"}, "")
}

func (r *queryResolver) PolyLeaderboard(ctx context.Context, metric *string, window *string, limit *int) ([]*model.PolyLeaderboardEntry, error) {
	return restCall[[]*model.PolyLeaderboardEntry](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: "/api/v1/leaderboard", Query: restQuery("metric", metric, "window", window, "limit", limit)}, "")
}

func (r *queryResolver) PolyEvents(ctx context.Context, status *string, category *string, tag *string, q *string, featured *bool, sort *string, limit *int, cursor *string) (*model.PolyEventPage, error) {
	return restCall[*model.PolyEventPage](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: "/api/v1/events", Query: restQuery("status", status, "category", category, "tag", tag, "q", q, "featured", featured, "sort", sort, "limit", limit, "cursor", cursor)}, "")
}

func (r *queryResolver) PolyEvent(ctx context.Context, id string) (*model.PolyEvent, error) {
	return restCall[*model.PolyEvent](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v", id)}, "")
}

func (r *queryResolver) PolyRelatedEvents(ctx context.Context, id string, limit *int) ([]*model.PolyEvent, error) {
	return restCall[[]*model.PolyEvent](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/related", id), Query: restQuery("limit", limit)}, "")
}

func (r *queryResolver) PolyComments(ctx context.Context, id string, sort *string, limit *int, cursor *string) (*model.PolyCommentPage, error) {
	return restCall[*model.PolyCommentPage](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/comments", id), Query: restQuery("sort", sort, "limit", limit, "cursor", cursor)}, "")
}

func (r *queryResolver) PolyMarkets(ctx context.Context, q *string, status *string, limit *int, offset *int) (*model.PolyMarketSearch, error) {
	return restCall[*model.PolyMarketSearch](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: "/api/v1/markets", Query: restQuery("q", q, "status", status, "limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) PolyMarket(ctx context.Context, id string) (*model.PolyMarket, error) {
	return restCall[*model.PolyMarket](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/markets/%v", id)}, "")
}

func (r *queryResolver) PolyOrderBook(ctx context.Context, id string, outcome *string) (*model.PolyOrderBook, error) {
	return restCall[*model.PolyOrderBook](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/markets/%v/book", id), Query: restQuery("outcome", outcome)}, "")
}

func (r *queryResolver) PolyQuote(ctx context.Context, id string, side *string, outcome *string, amount *int) (*model.PolyQuote, error) {
	return restCall[*model.PolyQuote](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/markets/%v/quote", id), Query: restQuery("side", side, "outcome", outcome, "amount", amount)}, "")
}

func (r *queryResolver) PolyMarketTrades(ctx context.Context, id string, limit *int, offset *int) ([]*model.PolyTrade, error) {
	return restCall[[]*model.PolyTrade](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/markets/%v/trades", id), Query: restQuery("limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) PolyPriceHistory(ctx context.Context, id string, rangeArg *string, bucket *int) ([]*model.PolyCandle, error) {
	return restCall[[]*model.PolyCandle](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/markets/%v/prices-history", id), Query: restQuery("range", rangeArg, "bucket", bucket)}, "")
}

func (r *queryResolver) PolyTopHolders(ctx context.Context, id string, outcome *string, limit *int) ([]*model.PolyHolderEntry, error) {
	return restCall[[]*model.PolyHolderEntry](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/markets/%v/holders", id), Query: restQuery("outcome", outcome, "limit", limit)}, "")
}

func (r *queryResolver) PolyMarketRewards(ctx context.Context, id string) (*model.PolyMarketRewards, error) {
	return restCall[*model.PolyMarketRewards](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/markets/%v/rewards", id)}, "")
}

func (r *queryResolver) PolyExchange(ctx context.Context) (*model.PolyExchangeSummary, error) {
	return restCall[*model.PolyExchangeSummary](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: "/api/v1/admin/exchange"}, "")
}

func (r *queryResolver) PolyOrder(ctx context.Context, orderID int) (*model.PolyOrder, error) {
	return restCall[*model.PolyOrder](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v", orderID)}, "")
}

func (r *queryResolver) PolyProfile(ctx context.Context, userID string) (*model.PolyProfile, error) {
	return restCall[*model.PolyProfile](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/profile", userID)}, "")
}

func (r *queryResolver) PolyBalance(ctx context.Context, userID string) (*model.PolyBalance, error) {
	return restCall[*model.PolyBalance](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/balance", userID)}, "")
}

func (r *queryResolver) PolyWallet(ctx context.Context, userID string) (*model.PolyWallet, error) {
	return restCall[*model.PolyWallet](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/wallet", userID)}, "")
}

func (r *queryResolver) PolyPortfolio(ctx context.Context, userID string) (*model.PolyPortfolio, error) {
	return restCall[*model.PolyPortfolio](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/portfolio", userID)}, "")
}

func (r *queryResolver) PolyClosedPositions(ctx context.Context, userID string) ([]*model.PolyPosition, error) {
	return restCall[[]*model.PolyPosition](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/positions/closed", userID)}, "")
}

func (r *queryResolver) PolyActivity(ctx context.Context, userID string, limit *int) ([]*model.PolyActivity, error) {
	return restCall[[]*model.PolyActivity](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/activity", userID), Query: restQuery("limit", limit)}, "")
}

func (r *queryResolver) PolyUserOrders(ctx context.Context, userID string, marketID *int, status *string, limit *int, cursor *string) (*model.PolyOrderPage, error) {
	return restCall[*model.PolyOrderPage](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/orders", userID), Query: restQuery("market_id", marketID, "status", status, "limit", limit, "cursor", cursor)}, "")
}

func (r *queryResolver) PolyWatchlist(ctx context.Context, userID string) ([]*model.PolyEvent, error) {
	return restCall[[]*model.PolyEvent](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/watchlist", userID)}, "")
}

func (r *queryResolver) PolyUserRewards(ctx context.Context, userID string, limit *int, cursor *string) (*model.PolyUserRewardsPage, error) {
	return restCall[*model.PolyUserRewardsPage](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/rewards", userID), Query: restQuery("limit", limit, "cursor", cursor)}, "")
}

func (r *queryResolver) PolyReferral(ctx context.Context, userID string) (*model.PolyReferralSummary, error) {
	return restCall[*model.PolyReferralSummary](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/referral", userID)}, "")
}

func (r *queryResolver) PolyMyRewards(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/rewards", userIDOf(ctx))}, "")
}

func (r *queryResolver) PolyMyReferral(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "polymarket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/referral", userIDOf(ctx))}, "")
}
