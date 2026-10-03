package webrtc_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "webrtc-service"

const DefaultBaseURL = "http://webrtc-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) IceServers(ctx context.Context, userID *string) (*model.WebrtcIceServers, error) {
	path := "/api/ice-servers"
	q := url.Values{}
	if userID != nil {
		q.Set("user_id", *userID)
	}
	h := http.Header{}
	var out *model.WebrtcIceServers
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Room(ctx context.Context, roomID string) (*model.WebrtcRoomInfo, error) {
	path := "/api/rooms/" + url.PathEscape(roomID)
	q := url.Values{}
	h := http.Header{}
	var out *model.WebrtcRoomInfo
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
