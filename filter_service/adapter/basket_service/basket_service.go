package basket_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "basket-service"

const DefaultBaseURL = "http://basket-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Baskets(ctx context.Context) ([]*model.BasketBasket, error) {
	path := "/api/v1/baskets"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BasketBasket
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Basket(ctx context.Context, id int) (*model.BasketBasket, error) {
	path := "/api/v1/baskets/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BasketBasket
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) BasketLines(ctx context.Context) ([]*model.BasketBasketLine, error) {
	path := "/api/v1/basket-lines"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BasketBasketLine
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) BasketLine(ctx context.Context, id int) (*model.BasketBasketLine, error) {
	path := "/api/v1/basket-lines/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BasketBasketLine
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) BasketLineAttributes(ctx context.Context) ([]*model.BasketBasketLineAttribute, error) {
	path := "/api/v1/basket-line-attributes"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BasketBasketLineAttribute
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) BasketLineAttribute(ctx context.Context, id int) (*model.BasketBasketLineAttribute, error) {
	path := "/api/v1/basket-line-attributes/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BasketBasketLineAttribute
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Orders(ctx context.Context) ([]*model.BasketOrder, error) {
	path := "/api/v1/orders"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BasketOrder
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Order(ctx context.Context, id int) (*model.BasketOrder, error) {
	path := "/api/v1/orders/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BasketOrder
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) User(ctx context.Context, id int) (*model.BasketUser, error) {
	path := "/api/v1/users/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BasketUser
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
