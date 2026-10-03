package shortlink_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "shortlink-service"

const DefaultBaseURL = "http://shortlink-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Links(ctx context.Context, userID *string) ([]*model.ShortlinkLink, error) {
	path := "/api/links"
	q := url.Values{}
	if userID != nil {
		q.Set("userId", *userID)
	}
	h := http.Header{}
	var out []*model.ShortlinkLink
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Link(ctx context.Context, id string, userID *string) (*model.ShortlinkLink, error) {
	path := "/api/links/" + url.PathEscape(id)
	q := url.Values{}
	if userID != nil {
		q.Set("userId", *userID)
	}
	h := http.Header{}
	var out *model.ShortlinkLink
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AnalyticsOverview(ctx context.Context, userID *string, days *int) (*model.ShortlinkAnalytics, error) {
	path := "/api/analytics/overview"
	q := url.Values{}
	if userID != nil {
		q.Set("userId", *userID)
	}
	if days != nil {
		q.Set("days", strconv.Itoa(*days))
	}
	h := http.Header{}
	var out *model.ShortlinkAnalytics
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) LinkAnalytics(ctx context.Context, linkID string, userID *string, days *int) (*model.ShortlinkAnalytics, error) {
	path := "/api/analytics/links/" + url.PathEscape(linkID)
	q := url.Values{}
	if userID != nil {
		q.Set("userId", *userID)
	}
	if days != nil {
		q.Set("days", strconv.Itoa(*days))
	}
	h := http.Header{}
	var out *model.ShortlinkAnalytics
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Resolve(ctx context.Context, shortCode string) (*model.ShortlinkResolve, error) {
	path := "/api/sdk/v1/resolve/" + url.PathEscape(shortCode)
	q := url.Values{}
	h := http.Header{}
	var out *model.ShortlinkResolve
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ResolveWithTemplate(ctx context.Context, templateSlug string, shortCode string) (*model.ShortlinkResolve, error) {
	path := "/api/sdk/v1/resolve/" + url.PathEscape(templateSlug) + "/" + url.PathEscape(shortCode)
	q := url.Values{}
	h := http.Header{}
	var out *model.ShortlinkResolve
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Attribution(ctx context.Context, fingerprint string) (*model.ShortlinkAttribution, error) {
	path := "/api/sdk/v1/attribution/" + url.PathEscape(fingerprint)
	q := url.Values{}
	h := http.Header{}
	var out *model.ShortlinkAttribution
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) SdkHealth(ctx context.Context) (*model.ShortlinkSdkHealth, error) {
	path := "/api/sdk/v1/health"
	q := url.Values{}
	h := http.Header{}
	var out *model.ShortlinkSdkHealth
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Webhooks(ctx context.Context, userID *string) ([]*model.ShortlinkWebhook, error) {
	path := "/api/webhooks"
	q := url.Values{}
	if userID != nil {
		q.Set("userId", *userID)
	}
	h := http.Header{}
	var out []*model.ShortlinkWebhook
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Webhook(ctx context.Context, id string, userID *string) (*model.ShortlinkWebhook, error) {
	path := "/api/webhooks/" + url.PathEscape(id)
	q := url.Values{}
	if userID != nil {
		q.Set("userId", *userID)
	}
	h := http.Header{}
	var out *model.ShortlinkWebhook
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Templates(ctx context.Context, userID *string) ([]*model.ShortlinkTemplate, error) {
	path := "/api/templates"
	q := url.Values{}
	if userID != nil {
		q.Set("userId", *userID)
	}
	h := http.Header{}
	var out []*model.ShortlinkTemplate
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Template(ctx context.Context, id string, userID *string) (*model.ShortlinkTemplate, error) {
	path := "/api/templates/" + url.PathEscape(id)
	q := url.Values{}
	if userID != nil {
		q.Set("userId", *userID)
	}
	h := http.Header{}
	var out *model.ShortlinkTemplate
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AppleAppSiteAssociation(ctx context.Context) (map[string]interface{}, error) {
	path := "/.well-known/apple-app-site-association"
	q := url.Values{}
	h := http.Header{}
	var out map[string]interface{}
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AssetLinks(ctx context.Context) ([]map[string]interface{}, error) {
	path := "/.well-known/assetlinks.json"
	q := url.Values{}
	h := http.Header{}
	var out []map[string]interface{}
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
