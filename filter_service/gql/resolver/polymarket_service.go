package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) PolyQuery() generated.PolyQueryResolver { return &polyQueryResolver{r} }

type polyQueryResolver struct{ *Resolver }

func (r *polyQueryResolver) Categories(ctx context.Context, obj *model.PolyQuery) ([]*model.PolyCategoryCount, error) {
	return r.Clients.PolymarketService.Categories(ctx)
}

func (r *polyQueryResolver) Leaderboard(ctx context.Context, obj *model.PolyQuery, metric *string, window *string, limit *int) ([]*model.PolyLeaderboardEntry, error) {
	return r.Clients.PolymarketService.Leaderboard(ctx, metric, window, limit)
}

func (r *polyQueryResolver) Events(ctx context.Context, obj *model.PolyQuery, status *string, category *string, tag *string, qArg *string, featured *bool, sort *string, limit *int, cursor *string) (*model.PolyEventPage, error) {
	return r.Clients.PolymarketService.Events(ctx, status, category, tag, qArg, featured, sort, limit, cursor)
}

func (r *polyQueryResolver) Event(ctx context.Context, obj *model.PolyQuery, id string) (*model.PolyEvent, error) {
	return r.Clients.PolymarketService.Event(ctx, id)
}

func (r *polyQueryResolver) RelatedEvents(ctx context.Context, obj *model.PolyQuery, id string, limit *int) ([]*model.PolyEvent, error) {
	return r.Clients.PolymarketService.RelatedEvents(ctx, id, limit)
}

func (r *polyQueryResolver) Comments(ctx context.Context, obj *model.PolyQuery, id string, sort *string, limit *int, cursor *string) (*model.PolyCommentPage, error) {
	return r.Clients.PolymarketService.Comments(ctx, id, sort, limit, cursor)
}

func (r *polyQueryResolver) Markets(ctx context.Context, obj *model.PolyQuery, qArg *string, status *string, limit *int, offset *int) (*model.PolyMarketSearch, error) {
	return r.Clients.PolymarketService.Markets(ctx, qArg, status, limit, offset)
}

func (r *polyQueryResolver) Market(ctx context.Context, obj *model.PolyQuery, id string) (*model.PolyMarket, error) {
	return r.Clients.PolymarketService.Market(ctx, id)
}

func (r *polyQueryResolver) OrderBook(ctx context.Context, obj *model.PolyQuery, id string, outcome *string) (*model.PolyOrderBook, error) {
	return r.Clients.PolymarketService.OrderBook(ctx, id, outcome)
}

func (r *polyQueryResolver) Quote(ctx context.Context, obj *model.PolyQuery, id string, side *string, outcome *string, amount *int) (*model.PolyQuote, error) {
	return r.Clients.PolymarketService.Quote(ctx, id, side, outcome, amount)
}

func (r *polyQueryResolver) MarketTrades(ctx context.Context, obj *model.PolyQuery, id string, limit *int, offset *int) ([]*model.PolyTrade, error) {
	return r.Clients.PolymarketService.MarketTrades(ctx, id, limit, offset)
}

func (r *polyQueryResolver) PriceHistory(ctx context.Context, obj *model.PolyQuery, id string, rangeArg *string, bucket *int) ([]*model.PolyCandle, error) {
	return r.Clients.PolymarketService.PriceHistory(ctx, id, rangeArg, bucket)
}

func (r *polyQueryResolver) TopHolders(ctx context.Context, obj *model.PolyQuery, id string, outcome *string, limit *int) ([]*model.PolyHolderEntry, error) {
	return r.Clients.PolymarketService.TopHolders(ctx, id, outcome, limit)
}

func (r *polyQueryResolver) MarketRewards(ctx context.Context, obj *model.PolyQuery, id string) (*model.PolyMarketRewards, error) {
	return r.Clients.PolymarketService.MarketRewards(ctx, id)
}

func (r *polyQueryResolver) Exchange(ctx context.Context, obj *model.PolyQuery) (*model.PolyExchangeSummary, error) {
	return r.Clients.PolymarketService.Exchange(ctx)
}

func (r *polyQueryResolver) Order(ctx context.Context, obj *model.PolyQuery, orderID int) (*model.PolyOrder, error) {
	return r.Clients.PolymarketService.Order(ctx, orderID)
}

func (r *polyQueryResolver) Profile(ctx context.Context, obj *model.PolyQuery, userID string) (*model.PolyProfile, error) {
	return r.Clients.PolymarketService.Profile(ctx, userID)
}

func (r *polyQueryResolver) Balance(ctx context.Context, obj *model.PolyQuery, userID string) (*model.PolyBalance, error) {
	return r.Clients.PolymarketService.Balance(ctx, userID)
}

func (r *polyQueryResolver) Wallet(ctx context.Context, obj *model.PolyQuery, userID string) (*model.PolyWallet, error) {
	return r.Clients.PolymarketService.Wallet(ctx, userID)
}

func (r *polyQueryResolver) Portfolio(ctx context.Context, obj *model.PolyQuery, userID string) (*model.PolyPortfolio, error) {
	return r.Clients.PolymarketService.Portfolio(ctx, userID)
}

func (r *polyQueryResolver) ClosedPositions(ctx context.Context, obj *model.PolyQuery, userID string) ([]*model.PolyPosition, error) {
	return r.Clients.PolymarketService.ClosedPositions(ctx, userID)
}

func (r *polyQueryResolver) Activity(ctx context.Context, obj *model.PolyQuery, userID string, limit *int) ([]*model.PolyActivity, error) {
	return r.Clients.PolymarketService.Activity(ctx, userID, limit)
}

func (r *polyQueryResolver) UserOrders(ctx context.Context, obj *model.PolyQuery, userID string, marketID *int, status *string, limit *int, cursor *string) (*model.PolyOrderPage, error) {
	return r.Clients.PolymarketService.UserOrders(ctx, userID, marketID, status, limit, cursor)
}

func (r *polyQueryResolver) Watchlist(ctx context.Context, obj *model.PolyQuery, userID string) ([]*model.PolyEvent, error) {
	return r.Clients.PolymarketService.Watchlist(ctx, userID)
}

func (r *polyQueryResolver) UserRewards(ctx context.Context, obj *model.PolyQuery, userID string, limit *int, cursor *string) (*model.PolyUserRewardsPage, error) {
	return r.Clients.PolymarketService.UserRewards(ctx, userID, limit, cursor)
}

func (r *polyQueryResolver) Referral(ctx context.Context, obj *model.PolyQuery, userID string) (*model.PolyReferralSummary, error) {
	return r.Clients.PolymarketService.Referral(ctx, userID)
}
