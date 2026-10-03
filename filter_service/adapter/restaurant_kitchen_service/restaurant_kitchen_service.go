package restaurant_kitchen_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "restaurant-kitchen-service"

const DefaultBaseURL = "http://kitchen-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Categories(ctx context.Context) ([]*model.RKitchenCategory, error) {
	path := "/api/v1/category"
	q := url.Values{}
	h := http.Header{}
	var out []*model.RKitchenCategory
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Foods(ctx context.Context, limit *int, page *int, sort *string) (*model.RKitchenFoodPage, error) {
	path := "/api/v1/food"
	q := url.Values{}
	if limit != nil {
		q.Set("Limit", strconv.Itoa(*limit))
	}
	if page != nil {
		q.Set("Page", strconv.Itoa(*page))
	}
	if sort != nil {
		q.Set("Sort", *sort)
	}
	h := http.Header{}
	var out *model.RKitchenFoodPage
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) FoodsByIDs(ctx context.Context, ids string) ([]*model.RKitchenFood, error) {
	path := "/api/v1/food/batch"
	q := url.Values{}
	q.Set("ids", ids)
	h := http.Header{}
	var out []*model.RKitchenFood
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) Kitchens(ctx context.Context, limit *int, page *int, sort *string) (*model.RKitchenKitchenPage, error) {
	path := "/api/v1/kitchen"
	q := url.Values{}
	if limit != nil {
		q.Set("Limit", strconv.Itoa(*limit))
	}
	if page != nil {
		q.Set("Page", strconv.Itoa(*page))
	}
	if sort != nil {
		q.Set("Sort", *sort)
	}
	h := http.Header{}
	var out *model.RKitchenKitchenPage
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
