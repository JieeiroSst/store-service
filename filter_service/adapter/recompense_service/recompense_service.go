package recompense_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "recompense-service"

const DefaultBaseURL = "http://recompense-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Recompenses(ctx context.Context, page *int, limit *int) ([]*model.RecompRecompense, error) {
	path := "/api/recompenses"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.RecompRecompense
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Recompense(ctx context.Context, id string) (*model.RecompRecompense, error) {
	path := "/api/recompenses/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.RecompRecompense
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MemberRecompenses(ctx context.Context, memberID int, page *int, limit *int) ([]*model.RecompRecompense, error) {
	path := "/api/recompenses/member/" + url.PathEscape(strconv.Itoa(memberID))
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out []*model.RecompRecompense
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MemberPoints(ctx context.Context, memberID int) (*model.RecompMemberStatus, error) {
	path := "/api/members/" + url.PathEscape(strconv.Itoa(memberID)) + "/points"
	q := url.Values{}
	h := http.Header{}
	var out *model.RecompMemberStatus
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) BestRecompense(ctx context.Context, memberID int, amount float64) (*model.RecompRedemption, error) {
	path := "/api/members/" + url.PathEscape(strconv.Itoa(memberID)) + "/recompenses/best"
	q := url.Values{}
	q.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	h := http.Header{}
	var out *model.RecompRedemption
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
