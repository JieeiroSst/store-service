package threads_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "threads-service"

const DefaultBaseURL = "http://threads-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Posts(ctx context.Context, authorID *string, cursor *string, limit *int) (*model.ThreadsPostPage, error) {
	path := "/api/v1/posts"
	q := url.Values{}
	if authorID != nil {
		q.Set("author_id", *authorID)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ThreadsPostPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Post(ctx context.Context, id string) (*model.ThreadsPost, error) {
	path := "/api/v1/posts/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.ThreadsPost
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Comments(ctx context.Context, id string, cursor *string, limit *int) (*model.ThreadsCommentPage, error) {
	path := "/api/v1/posts/" + url.PathEscape(id) + "/comments"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ThreadsCommentPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Followers(ctx context.Context, id string, cursor *string, limit *int) (*model.ThreadsAuthorPage, error) {
	path := "/api/v1/users/" + url.PathEscape(id) + "/followers"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ThreadsAuthorPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Following(ctx context.Context, id string, cursor *string, limit *int) (*model.ThreadsAuthorPage, error) {
	path := "/api/v1/users/" + url.PathEscape(id) + "/following"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ThreadsAuthorPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) HomeFeed(ctx context.Context, cursor *string, limit *int) (*model.ThreadsPostPage, error) {
	path := "/api/v1/feed"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ThreadsPostPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Bookmarks(ctx context.Context, cursor *string, limit *int) (*model.ThreadsPostPage, error) {
	path := "/api/v1/bookmarks"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ThreadsPostPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) TrendingTags(ctx context.Context, limit *int) ([]*model.ThreadsTagCount, error) {
	path := "/api/v1/tags/trending"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.ThreadsTagCount
	err := c.rest.Get(ctx, path, q, h, "tags", &out)
	return out, err
}
