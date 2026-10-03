package food_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "food-service"

const DefaultBaseURL = "http://food-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Food(ctx context.Context, id int) (*model.FoodFood, error) {
	path := "/api/food/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.FoodFood
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Foods(ctx context.Context) ([]*model.FoodFood, error) {
	path := "/api/food/"
	q := url.Values{}
	h := http.Header{}
	var out []*model.FoodFood
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) SearchFoods(ctx context.Context, name string) ([]*model.FoodFood, error) {
	path := "/api/food/search"
	q := url.Values{}
	q.Set("name", name)
	h := http.Header{}
	var out []*model.FoodFood
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
