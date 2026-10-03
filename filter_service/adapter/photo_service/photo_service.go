package photo_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "photo-service"

const DefaultBaseURL = "http://photo-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Composition(ctx context.Context, id string) (*model.PhotoComposition, error) {
	path := "/v1/compositions/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PhotoComposition
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
