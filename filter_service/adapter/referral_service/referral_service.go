package referral_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "referral-service"

const DefaultBaseURL = "http://referral-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Link(ctx context.Context, refCode string) (*model.ReferralReferralLink, error) {
	path := "/api/v1/referral/link/" + url.PathEscape(refCode)
	q := url.Values{}
	h := http.Header{}
	var out *model.ReferralReferralLink
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) UserLinks(ctx context.Context, userID string, limit *int, cursor *string) (*model.ReferralLinkList, error) {
	path := "/api/v1/referral/user/" + url.PathEscape(userID) + "/links"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	h := http.Header{}
	var out *model.ReferralLinkList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Status(ctx context.Context, refCode string) (*model.ReferralReferralStatusResponse, error) {
	path := "/api/v1/referral/status"
	q := url.Values{}
	q.Set("ref_code", refCode)
	h := http.Header{}
	var out *model.ReferralReferralStatusResponse
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) UserStats(ctx context.Context, userID string) (*model.ReferralUserReferralStats, error) {
	path := "/api/v1/referral/user/" + url.PathEscape(userID) + "/stats"
	q := url.Values{}
	h := http.Header{}
	var out *model.ReferralUserReferralStats
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) ShareTargets(ctx context.Context, refCode string) ([]*model.ReferralShareTarget, error) {
	path := "/api/v1/referral/share/" + url.PathEscape(refCode)
	q := url.Values{}
	h := http.Header{}
	var out []*model.ReferralShareTarget
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) ActiveRewardProgram(ctx context.Context) (*model.ReferralRewardProgram, error) {
	path := "/api/v1/programs/active"
	q := url.Values{}
	h := http.Header{}
	var out *model.ReferralRewardProgram
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
