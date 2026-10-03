package livestream_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "livestream_service"

const DefaultBaseURL = "http://livestream-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Rooms(ctx context.Context, live *bool) ([]*model.LiveRoom, error) {
	path := "/api/v1/rooms"
	q := url.Values{}
	if live != nil {
		q.Set("live", strconv.FormatBool(*live))
	}
	h := http.Header{}
	var out []*model.LiveRoom
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Room(ctx context.Context, id string) (*model.LiveRoom, error) {
	path := "/api/v1/rooms/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.LiveRoom
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ActiveStream(ctx context.Context, id string) (*model.LiveStream, error) {
	path := "/api/v1/rooms/" + url.PathEscape(id) + "/stream"
	q := url.Values{}
	h := http.Header{}
	var out *model.LiveStream
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Recordings(ctx context.Context, id string) ([]*model.LiveRecording, error) {
	path := "/api/v1/rooms/" + url.PathEscape(id) + "/recordings"
	q := url.Values{}
	h := http.Header{}
	var out []*model.LiveRecording
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ViewerCount(ctx context.Context, id string) (*model.LiveViewerCount, error) {
	path := "/api/v1/rooms/" + url.PathEscape(id) + "/viewers"
	q := url.Values{}
	h := http.Header{}
	var out *model.LiveViewerCount
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Playback(ctx context.Context, id string) (*model.LivePlaybackInfo, error) {
	path := "/api/v1/rooms/" + url.PathEscape(id) + "/playback"
	q := url.Values{}
	h := http.Header{}
	var out *model.LivePlaybackInfo
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
