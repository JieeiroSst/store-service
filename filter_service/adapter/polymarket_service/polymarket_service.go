package polymarket_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "polymarket-service"

const DefaultBaseURL = "http://polymarket-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Categories(ctx context.Context) ([]*model.PolyCategoryCount, error) {
	path := "/api/v1/categories"
	q := url.Values{}
	h := http.Header{}
	var out []*model.PolyCategoryCount
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Leaderboard(ctx context.Context, metric *string, window *string, limit *int) ([]*model.PolyLeaderboardEntry, error) {
	path := "/api/v1/leaderboard"
	q := url.Values{}
	if metric != nil {
		q.Set("metric", *metric)
	}
	if window != nil {
		q.Set("window", *window)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.PolyLeaderboardEntry
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Events(ctx context.Context, status *string, category *string, tag *string, qArg *string, featured *bool, sort *string, limit *int, cursor *string) (*model.PolyEventPage, error) {
	path := "/api/v1/events"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if category != nil {
		q.Set("category", *category)
	}
	if tag != nil {
		q.Set("tag", *tag)
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if featured != nil {
		q.Set("featured", strconv.FormatBool(*featured))
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	h := http.Header{}
	var out *model.PolyEventPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Event(ctx context.Context, id string) (*model.PolyEvent, error) {
	path := "/api/v1/events/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PolyEvent
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RelatedEvents(ctx context.Context, id string, limit *int) ([]*model.PolyEvent, error) {
	path := "/api/v1/events/" + url.PathEscape(id) + "/related"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.PolyEvent
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Comments(ctx context.Context, id string, sort *string, limit *int, cursor *string) (*model.PolyCommentPage, error) {
	path := "/api/v1/events/" + url.PathEscape(id) + "/comments"
	q := url.Values{}
	if sort != nil {
		q.Set("sort", *sort)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	h := http.Header{}
	var out *model.PolyCommentPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Markets(ctx context.Context, qArg *string, status *string, limit *int, offset *int) (*model.PolyMarketSearch, error) {
	path := "/api/v1/markets"
	q := url.Values{}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if status != nil {
		q.Set("status", *status)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.PolyMarketSearch
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Market(ctx context.Context, id string) (*model.PolyMarket, error) {
	path := "/api/v1/markets/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PolyMarket
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) OrderBook(ctx context.Context, id string, outcome *string) (*model.PolyOrderBook, error) {
	path := "/api/v1/markets/" + url.PathEscape(id) + "/book"
	q := url.Values{}
	if outcome != nil {
		q.Set("outcome", *outcome)
	}
	h := http.Header{}
	var out *model.PolyOrderBook
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Quote(ctx context.Context, id string, side *string, outcome *string, amount *int) (*model.PolyQuote, error) {
	path := "/api/v1/markets/" + url.PathEscape(id) + "/quote"
	q := url.Values{}
	if side != nil {
		q.Set("side", *side)
	}
	if outcome != nil {
		q.Set("outcome", *outcome)
	}
	if amount != nil {
		q.Set("amount", strconv.Itoa(*amount))
	}
	h := http.Header{}
	var out *model.PolyQuote
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MarketTrades(ctx context.Context, id string, limit *int, offset *int) ([]*model.PolyTrade, error) {
	path := "/api/v1/markets/" + url.PathEscape(id) + "/trades"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.PolyTrade
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PriceHistory(ctx context.Context, id string, rangeArg *string, bucket *int) ([]*model.PolyCandle, error) {
	path := "/api/v1/markets/" + url.PathEscape(id) + "/prices-history"
	q := url.Values{}
	if rangeArg != nil {
		q.Set("range", *rangeArg)
	}
	if bucket != nil {
		q.Set("bucket", strconv.Itoa(*bucket))
	}
	h := http.Header{}
	var out []*model.PolyCandle
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) TopHolders(ctx context.Context, id string, outcome *string, limit *int) ([]*model.PolyHolderEntry, error) {
	path := "/api/v1/markets/" + url.PathEscape(id) + "/holders"
	q := url.Values{}
	if outcome != nil {
		q.Set("outcome", *outcome)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.PolyHolderEntry
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MarketRewards(ctx context.Context, id string) (*model.PolyMarketRewards, error) {
	path := "/api/v1/markets/" + url.PathEscape(id) + "/rewards"
	q := url.Values{}
	h := http.Header{}
	var out *model.PolyMarketRewards
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Exchange(ctx context.Context) (*model.PolyExchangeSummary, error) {
	path := "/api/v1/admin/exchange"
	q := url.Values{}
	h := http.Header{}
	var out *model.PolyExchangeSummary
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Order(ctx context.Context, orderID int) (*model.PolyOrder, error) {
	path := "/api/v1/orders/" + url.PathEscape(strconv.Itoa(orderID))
	q := url.Values{}
	h := http.Header{}
	var out *model.PolyOrder
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Profile(ctx context.Context, userID string) (*model.PolyProfile, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/profile"
	q := url.Values{}
	h := http.Header{}
	var out *model.PolyProfile
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Balance(ctx context.Context, userID string) (*model.PolyBalance, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/balance"
	q := url.Values{}
	h := http.Header{}
	var out *model.PolyBalance
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Wallet(ctx context.Context, userID string) (*model.PolyWallet, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/wallet"
	q := url.Values{}
	h := http.Header{}
	var out *model.PolyWallet
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Portfolio(ctx context.Context, userID string) (*model.PolyPortfolio, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/portfolio"
	q := url.Values{}
	h := http.Header{}
	var out *model.PolyPortfolio
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClosedPositions(ctx context.Context, userID string) ([]*model.PolyPosition, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/positions/closed"
	q := url.Values{}
	h := http.Header{}
	var out []*model.PolyPosition
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Activity(ctx context.Context, userID string, limit *int) ([]*model.PolyActivity, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/activity"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.PolyActivity
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserOrders(ctx context.Context, userID string, marketID *int, status *string, limit *int, cursor *string) (*model.PolyOrderPage, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/orders"
	q := url.Values{}
	if marketID != nil {
		q.Set("market_id", strconv.Itoa(*marketID))
	}
	if status != nil {
		q.Set("status", *status)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	h := http.Header{}
	var out *model.PolyOrderPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Watchlist(ctx context.Context, userID string) ([]*model.PolyEvent, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/watchlist"
	q := url.Values{}
	h := http.Header{}
	var out []*model.PolyEvent
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserRewards(ctx context.Context, userID string, limit *int, cursor *string) (*model.PolyUserRewardsPage, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/rewards"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	h := http.Header{}
	var out *model.PolyUserRewardsPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Referral(ctx context.Context, userID string) (*model.PolyReferralSummary, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/referral"
	q := url.Values{}
	h := http.Header{}
	var out *model.PolyReferralSummary
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
