package notification_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "notification_service"

const DefaultBaseURL = "http://notification-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Notifications(ctx context.Context) ([]*model.NotifNotification, error) {
	path := "/api/v1/notifications"
	q := url.Values{}
	h := http.Header{}
	var out []*model.NotifNotification
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Notification(ctx context.Context, id int) (*model.NotifNotification, error) {
	path := "/api/v1/notifications/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.NotifNotification
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Campaigns(ctx context.Context, limit *int, offset *int) ([]*model.NotifCampaign, error) {
	path := "/api/v1/campaigns"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.NotifCampaign
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Campaign(ctx context.Context, id int) (*model.NotifCampaignView, error) {
	path := "/api/v1/campaigns/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.NotifCampaignView
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AuditDeliveries(ctx context.Context, email *string, phone *string, channel *string, status *string, sourceType *string, requestedBy *string, userID *int, sourceID *int, beforeID *int, limit *int, from *string, to *string) (*model.NotifAuditPage, error) {
	path := "/api/v1/audit/deliveries"
	q := url.Values{}
	if email != nil {
		q.Set("email", *email)
	}
	if phone != nil {
		q.Set("phone", *phone)
	}
	if channel != nil {
		q.Set("channel", *channel)
	}
	if status != nil {
		q.Set("status", *status)
	}
	if sourceType != nil {
		q.Set("source_type", *sourceType)
	}
	if requestedBy != nil {
		q.Set("requested_by", *requestedBy)
	}
	if userID != nil {
		q.Set("user_id", strconv.Itoa(*userID))
	}
	if sourceID != nil {
		q.Set("source_id", strconv.Itoa(*sourceID))
	}
	if beforeID != nil {
		q.Set("before_id", strconv.Itoa(*beforeID))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if from != nil {
		q.Set("from", *from)
	}
	if to != nil {
		q.Set("to", *to)
	}
	h := http.Header{}
	var out *model.NotifAuditPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AuditContent(ctx context.Context, id int) (*model.NotifAuditContent, error) {
	path := "/api/v1/audit/contents/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.NotifAuditContent
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserContact(ctx context.Context, id int) (*model.NotifUserContact, error) {
	path := "/api/v1/users/" + url.PathEscape(strconv.Itoa(id)) + "/contact"
	q := url.Values{}
	h := http.Header{}
	var out *model.NotifUserContact
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Devices(ctx context.Context, userID int) ([]*model.NotifDevice, error) {
	path := "/api/v1/devices"
	q := url.Values{}
	q.Set("user_id", strconv.Itoa(userID))
	h := http.Header{}
	var out []*model.NotifDevice
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Device(ctx context.Context, id int) (*model.NotifDevice, error) {
	path := "/api/v1/devices/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.NotifDevice
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
