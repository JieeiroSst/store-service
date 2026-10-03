package video_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "video-service"

const DefaultBaseURL = "http://video-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Videos(ctx context.Context, page *int, pageSize *int, sort *string, qArg *string) (*model.VideoVideoList, error) {
	path := "/api/videos"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	h := http.Header{}
	var out *model.VideoVideoList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Video(ctx context.Context, id string) (*model.VideoVideo, error) {
	path := "/api/videos/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.VideoVideo
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Related(ctx context.Context, id string, limit *int) ([]*model.VideoVideo, error) {
	path := "/api/videos/" + url.PathEscape(id) + "/related"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.VideoVideo
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}
