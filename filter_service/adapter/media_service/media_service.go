package media_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "media-service"

const DefaultBaseURL = "http://media-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Video(ctx context.Context, videoID int, userID int) (*model.MediaVideo, error) {
	path := "/api/internal/v1/videos/" + url.PathEscape(strconv.Itoa(videoID)) + "/user/" + url.PathEscape(strconv.Itoa(userID))
	q := url.Values{}
	h := http.Header{}
	var out *model.MediaVideo
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
