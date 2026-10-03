package ekyc_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "ekyc-service"

const DefaultBaseURL = "http://ekyc-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Status(ctx context.Context, userID string) (*model.EkycEkycStatus, error) {
	path := "/api/v1/ekyc/" + url.PathEscape(userID)
	q := url.Values{}
	h := http.Header{}
	var out *model.EkycEkycStatus
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
