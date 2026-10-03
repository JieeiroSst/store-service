package face_recognition_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "face-recognition-service"

const DefaultBaseURL = "http://face-recognition-service:1234"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) FacePosition(ctx context.Context) (*model.FaceRecPositions, error) {
	path := "/api/face/position"
	q := url.Values{}
	h := http.Header{}
	var out *model.FaceRecPositions
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
