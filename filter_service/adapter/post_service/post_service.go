package post_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "post_service"

const DefaultBaseURL = "http://post-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Posts(ctx context.Context, cursor *string, limit *int) (*model.PostPostPage, error) {
	path := "/api/v1/post"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.PostPostPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Post(ctx context.Context, id string) (*model.PostPost, error) {
	path := "/api/v1/post/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PostPost
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Categories(ctx context.Context, cursor *string, limit *int) (*model.PostCategoryPage, error) {
	path := "/api/v1/category"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.PostCategoryPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Category(ctx context.Context, id string) (*model.PostCategory, error) {
	path := "/api/v1/category/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PostCategory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
