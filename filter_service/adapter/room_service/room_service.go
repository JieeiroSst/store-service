package room_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "room-service"

const DefaultBaseURL = "http://room-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Me(ctx context.Context) (*model.RoomUser, error) {
	path := "/api/me"
	q := url.Values{}
	h := http.Header{}
	var out *model.RoomUser
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Rooms(ctx context.Context) ([]*model.RoomRoom, error) {
	path := "/api/rooms"
	q := url.Values{}
	h := http.Header{}
	var out []*model.RoomRoom
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Room(ctx context.Context, id int) (*model.RoomRoom, error) {
	path := "/api/rooms/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.RoomRoom
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Members(ctx context.Context, id int) ([]*model.RoomMember, error) {
	path := "/api/rooms/" + url.PathEscape(strconv.Itoa(id)) + "/members"
	q := url.Values{}
	h := http.Header{}
	var out []*model.RoomMember
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Messages(ctx context.Context, id int, before *int, limit *int) ([]*model.RoomMessage, error) {
	path := "/api/rooms/" + url.PathEscape(strconv.Itoa(id)) + "/messages"
	q := url.Values{}
	if before != nil {
		q.Set("before", strconv.Itoa(*before))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.RoomMessage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
