package cdn_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "cdn-service"

const DefaultBaseURL = "http://cdn-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Files(ctx context.Context, limit *int, offset *int) (*model.CdnFileList, error) {
	path := "/api/v1/files"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.CdnFileList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) File(ctx context.Context, id string) (*model.CdnFile, error) {
	path := "/api/v1/files/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.CdnFile
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
