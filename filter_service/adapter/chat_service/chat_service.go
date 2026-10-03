package chat_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "chat_service"

const DefaultBaseURL = "http://chat-service:6060"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Message(ctx context.Context, id int) (*model.ChatMessage, error) {
	path := "/message/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.ChatMessage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserReports(ctx context.Context, userID int) ([]*model.ChatReport, error) {
	path := "/report/" + url.PathEscape(strconv.Itoa(userID))
	q := url.Values{}
	h := http.Header{}
	var out []*model.ChatReport
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
