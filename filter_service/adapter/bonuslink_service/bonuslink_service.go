package bonuslink_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "bonuslink-service"

const DefaultBaseURL = "http://bonuslink-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) UserRewards(ctx context.Context, userID string, limit *int, offset *int) (*model.BonusRewardList, error) {
	path := "/api/v1/bonus/users/" + url.PathEscape(userID) + "/rewards"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.BonusRewardList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserBalances(ctx context.Context, userID string) ([]*model.BonusBalance, error) {
	path := "/api/v1/bonus/users/" + url.PathEscape(userID) + "/balances"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BonusBalance
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
