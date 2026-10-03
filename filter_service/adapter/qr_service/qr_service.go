package qr_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "qr-service"

const DefaultBaseURL = "http://qr-service:8080"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) QRCodes(ctx context.Context, page *int, limit *int, status *string, typeArg *string, search *string, createdBy *string) (*model.QRQRListResponse, error) {
	path := "/api/v1/qr"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if status != nil {
		q.Set("status", *status)
	}
	if typeArg != nil {
		q.Set("type", *typeArg)
	}
	if search != nil {
		q.Set("search", *search)
	}
	if createdBy != nil {
		q.Set("created_by", *createdBy)
	}
	h := http.Header{}
	var out *model.QRQRListResponse
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) QRCode(ctx context.Context, id string) (*model.QRQRCode, error) {
	path := "/api/v1/qr/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.QRQRCode
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) ScanHistory(ctx context.Context, id string, page *int, limit *int) (*model.QRScanHistoryResponse, error) {
	path := "/api/v1/qr/" + url.PathEscape(id) + "/history"
	q := url.Values{}
	if page != nil {
		q.Set("page", strconv.Itoa(*page))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.QRScanHistoryResponse
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) ScanStats(ctx context.Context, id string) (*model.QRScanStats, error) {
	path := "/api/v1/qr/" + url.PathEscape(id) + "/stats"
	q := url.Values{}
	h := http.Header{}
	var out *model.QRScanStats
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
