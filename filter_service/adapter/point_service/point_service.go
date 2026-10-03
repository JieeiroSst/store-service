package point_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "point-service"

const DefaultBaseURL = "http://point-service:3000"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) RewardPoints(ctx context.Context, perPage *int, sortOrder *string, cursor *string) (*model.PointRewardPointPage, error) {
	path := "/api/v1/reward-points"
	q := url.Values{}
	if perPage != nil {
		q.Set("per_page", strconv.Itoa(*perPage))
	}
	if sortOrder != nil {
		q.Set("sort_order", *sortOrder)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	h := http.Header{}
	var out *model.PointRewardPointPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RewardPoint(ctx context.Context, id string) (*model.PointRewardPoint, error) {
	path := "/api/v1/reward-points/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PointRewardPoint
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RewardDiscounts(ctx context.Context, perPage *int, sortOrder *string, cursor *string) (*model.PointRewardDiscountPage, error) {
	path := "/api/v1/reward-discounts"
	q := url.Values{}
	if perPage != nil {
		q.Set("per_page", strconv.Itoa(*perPage))
	}
	if sortOrder != nil {
		q.Set("sort_order", *sortOrder)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	h := http.Header{}
	var out *model.PointRewardDiscountPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RewardDiscount(ctx context.Context, id string) (*model.PointRewardDiscount, error) {
	path := "/api/v1/reward-discounts/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PointRewardDiscount
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ConvertedRewardPoints(ctx context.Context, perPage *int, sortOrder *string, cursor *string) (*model.PointConvertedRewardPointPage, error) {
	path := "/api/v1/converted-reward-points"
	q := url.Values{}
	if perPage != nil {
		q.Set("per_page", strconv.Itoa(*perPage))
	}
	if sortOrder != nil {
		q.Set("sort_order", *sortOrder)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	h := http.Header{}
	var out *model.PointConvertedRewardPointPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ConvertedRewardPoint(ctx context.Context, id string) (*model.PointConvertedRewardPoint, error) {
	path := "/api/v1/converted-reward-points/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PointConvertedRewardPoint
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
