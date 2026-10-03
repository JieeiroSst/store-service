package reward_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "reward-service"

const DefaultBaseURL = "http://reward-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Rewards(ctx context.Context, name *string, page *int, pageSize *int) (*model.RewardPage, error) {
	path := "/reward"
	q := url.Values{}
	if name != nil {
		q.Set("name", *name)
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	h := http.Header{}
	var out *model.RewardPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Reward(ctx context.Context, rewardID string) (*model.RewardPage, error) {
	path := "/reward/" + url.PathEscape(rewardID)
	q := url.Values{}
	h := http.Header{}
	var out *model.RewardPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
