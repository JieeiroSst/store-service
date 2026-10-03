package catalogues_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "catalogues-service"

const DefaultBaseURL = "http://catalogues-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Options(ctx context.Context) ([]*model.CatalogOption, error) {
	path := "/v1/options"
	q := url.Values{}
	h := http.Header{}
	var out []*model.CatalogOption
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Option(ctx context.Context, id int) (*model.CatalogOption, error) {
	path := "/v1/options/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CatalogOption
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ProductClasses(ctx context.Context) ([]*model.CatalogProductClass, error) {
	path := "/v1/product-classes"
	q := url.Values{}
	h := http.Header{}
	var out []*model.CatalogProductClass
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ProductClass(ctx context.Context, id int) (*model.CatalogProductClass, error) {
	path := "/v1/product-classes/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CatalogProductClass
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Products(ctx context.Context, qArg *string, public *bool, categoryID *int, limit *int, offset *int) ([]*model.CatalogProduct, error) {
	path := "/v1/products"
	q := url.Values{}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if public != nil {
		q.Set("public", strconv.FormatBool(*public))
	}
	if categoryID != nil {
		q.Set("category_id", strconv.Itoa(*categoryID))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.CatalogProduct
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Product(ctx context.Context, id int) (*model.CatalogProduct, error) {
	path := "/v1/products/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CatalogProduct
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Categories(ctx context.Context) ([]*model.CatalogCategory, error) {
	path := "/v1/categories"
	q := url.Values{}
	h := http.Header{}
	var out []*model.CatalogCategory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Category(ctx context.Context, id int) (*model.CatalogCategory, error) {
	path := "/v1/categories/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CatalogCategory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
