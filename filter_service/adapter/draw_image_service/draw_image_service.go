package draw_image_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "draw-image-service"

const DefaultBaseURL = "http://draw-image-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) ImageInfo(ctx context.Context, id string) (*model.DrawImageCollage, error) {
	path := "/images/" + url.PathEscape(id) + "/info"
	q := url.Values{}
	h := http.Header{}
	var out *model.DrawImageCollage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
