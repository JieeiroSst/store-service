package admanagement_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "admanagement-service"

const DefaultBaseURL = "http://admanagement-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Campaigns(ctx context.Context) ([]*model.AdmAdCampaign, error) {
	path := "/campaigns"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAdCampaign
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Campaign(ctx context.Context, id int) (*model.AdmAdCampaign, error) {
	path := "/campaigns/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.AdmAdCampaign
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CampaignAds(ctx context.Context, id int) ([]*model.AdmAd, error) {
	path := "/campaigns/" + url.PathEscape(strconv.Itoa(id)) + "/ads"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAd
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CampaignPerformance(ctx context.Context, id int, from *string, to *string) (*model.AdmCampaignPerformance, error) {
	path := "/campaigns/" + url.PathEscape(strconv.Itoa(id)) + "/performance"
	q := url.Values{}
	if from != nil {
		q.Set("from", *from)
	}
	if to != nil {
		q.Set("to", *to)
	}
	h := http.Header{}
	var out *model.AdmCampaignPerformance
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Categories(ctx context.Context) ([]*model.AdmAdCategory, error) {
	path := "/categories"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAdCategory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Category(ctx context.Context, id int) (*model.AdmAdCategory, error) {
	path := "/categories/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.AdmAdCategory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Positions(ctx context.Context) ([]*model.AdmAdPosition, error) {
	path := "/positions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAdPosition
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Position(ctx context.Context, id int) (*model.AdmAdPosition, error) {
	path := "/positions/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.AdmAdPosition
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ServeAd(ctx context.Context, id int, sessionID *string, country *string, device *string, gender *string, referrerURL *string, pageURL *string, age *int, userID *int) (*model.AdmServedAd, error) {
	path := "/positions/" + url.PathEscape(strconv.Itoa(id)) + "/serve"
	q := url.Values{}
	if sessionID != nil {
		q.Set("session_id", *sessionID)
	}
	if country != nil {
		q.Set("country", *country)
	}
	if device != nil {
		q.Set("device", *device)
	}
	if gender != nil {
		q.Set("gender", *gender)
	}
	if referrerURL != nil {
		q.Set("referrer_url", *referrerURL)
	}
	if pageURL != nil {
		q.Set("page_url", *pageURL)
	}
	if age != nil {
		q.Set("age", strconv.Itoa(*age))
	}
	if userID != nil {
		q.Set("user_id", strconv.Itoa(*userID))
	}
	h := http.Header{}
	var out *model.AdmServedAd
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Ads(ctx context.Context) ([]*model.AdmAd, error) {
	path := "/ads"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAd
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Ad(ctx context.Context, id int) (*model.AdmAd, error) {
	path := "/ads/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.AdmAd
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AdImpressions(ctx context.Context, id int) ([]*model.AdmAdImpression, error) {
	path := "/ads/" + url.PathEscape(strconv.Itoa(id)) + "/impressions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAdImpression
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AdClicks(ctx context.Context, id int) ([]*model.AdmAdClick, error) {
	path := "/ads/" + url.PathEscape(strconv.Itoa(id)) + "/clicks"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAdClick
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PositionMappings(ctx context.Context) ([]*model.AdmAdPositionMapping, error) {
	path := "/position-mappings"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAdPositionMapping
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PositionMapping(ctx context.Context, id int) (*model.AdmAdPositionMapping, error) {
	path := "/position-mappings/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.AdmAdPositionMapping
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Impressions(ctx context.Context) ([]*model.AdmAdImpression, error) {
	path := "/impressions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAdImpression
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Impression(ctx context.Context, id int) (*model.AdmAdImpression, error) {
	path := "/impressions/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.AdmAdImpression
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Clicks(ctx context.Context) ([]*model.AdmAdClick, error) {
	path := "/clicks"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAdClick
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Click(ctx context.Context, id int) (*model.AdmAdClick, error) {
	path := "/clicks/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.AdmAdClick
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) TargetingRules(ctx context.Context) ([]*model.AdmAdTargetingRule, error) {
	path := "/targeting-rules"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAdTargetingRule
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) TargetingRule(ctx context.Context, id int) (*model.AdmAdTargetingRule, error) {
	path := "/targeting-rules/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.AdmAdTargetingRule
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PerformanceSummaries(ctx context.Context) ([]*model.AdmAdPerformanceSummary, error) {
	path := "/performance-summaries"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AdmAdPerformanceSummary
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PerformanceSummary(ctx context.Context, id int) (*model.AdmAdPerformanceSummary, error) {
	path := "/performance-summaries/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.AdmAdPerformanceSummary
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
