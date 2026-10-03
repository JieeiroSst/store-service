package identified_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "identified_service"

const DefaultBaseURL = "http://identified-service:8000"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Classes(ctx context.Context) ([]*model.GymClass, error) {
	path := "/classes"
	q := url.Values{}
	h := http.Header{}
	var out []*model.GymClass
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Class(ctx context.Context, classID string) (*model.GymClass, error) {
	path := "/classes/" + url.PathEscape(classID)
	q := url.Values{}
	h := http.Header{}
	var out *model.GymClass
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GymPassVerification(ctx context.Context, gymPassID string) (*model.GymPass, error) {
	path := "/gym-passes/" + url.PathEscape(gymPassID) + "/verification"
	q := url.Values{}
	h := http.Header{}
	var out *model.GymPass
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
