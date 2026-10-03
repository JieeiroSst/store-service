package partner_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "partner-service"

const DefaultBaseURL = "http://partner-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Partners(ctx context.Context, limit *int, page *int, sort *string) (*model.PartnerPartnerPage, error) {
	path := "/v1/partner/"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	h := http.Header{}
	var out *model.PartnerPartnerPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) SearchPartners(ctx context.Context, limit *int, page *int, sort *string, typeArg *string, status *string, name *string, userID *string) (*model.PartnerPartnerPage, error) {
	path := "/v1/partner/search"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	if typeArg != nil {
		q.Set("type", *typeArg)
	}
	if status != nil {
		q.Set("status", *status)
	}
	if name != nil {
		q.Set("name", *name)
	}
	if userID != nil {
		q.Set("user_id", *userID)
	}
	h := http.Header{}
	var out *model.PartnerPartnerPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Statistics(ctx context.Context) (*model.PartnerPartnerStatistics, error) {
	path := "/v1/partner/statistics"
	q := url.Values{}
	h := http.Header{}
	var out *model.PartnerPartnerStatistics
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Activity(ctx context.Context, id string) (*model.PartnerPartnerActivity, error) {
	path := "/v1/partner/active"
	q := url.Values{}
	q.Set("id", id)
	h := http.Header{}
	var out *model.PartnerPartnerActivity
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Partner(ctx context.Context, id string) (*model.PartnerPartner, error) {
	path := "/v1/partner/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PartnerPartner
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Partnership(ctx context.Context, id string) (*model.PartnerPartnership, error) {
	path := "/v1/partnership/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PartnerPartnership
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Partnerships(ctx context.Context, limit *int, page *int, sort *string) (*model.PartnerPartnershipPage, error) {
	path := "/v1/partnership/"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	h := http.Header{}
	var out *model.PartnerPartnershipPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PartnershipsPartner(ctx context.Context, id string) (*model.PartnerPartnershipsPartner, error) {
	path := "/v1/partnerships-partner/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PartnerPartnershipsPartner
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PartnershipsPartners(ctx context.Context, limit *int, page *int, sort *string) (*model.PartnerPartnershipsPartnerPage, error) {
	path := "/v1/partnerships-partner/"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	h := http.Header{}
	var out *model.PartnerPartnershipsPartnerPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Projects(ctx context.Context, limit *int, page *int, sort *string) (*model.PartnerProjectPage, error) {
	path := "/v1/project/"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	h := http.Header{}
	var out *model.PartnerProjectPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Project(ctx context.Context, id string) (*model.PartnerProject, error) {
	path := "/v1/project/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.PartnerProject
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
