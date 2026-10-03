package shipping_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "shipping-service"

const DefaultBaseURL = "http://shipping-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Shipments(ctx context.Context, status *string, limit *int, offset *int) (*model.ShipShipmentList, error) {
	path := "/api/v1/shipments"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.ShipShipmentList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Shipment(ctx context.Context, id int) (*model.ShipShipmentDetail, error) {
	path := "/api/v1/shipments/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.ShipShipmentDetail
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ShipmentByClientCode(ctx context.Context, code string) (*model.ShipShipmentDetail, error) {
	path := "/api/v1/shipments/by-client-code/" + url.PathEscape(code)
	q := url.Values{}
	h := http.Header{}
	var out *model.ShipShipmentDetail
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Warehouses(ctx context.Context) ([]*model.ShipWarehouse, error) {
	path := "/api/v1/warehouses"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ShipWarehouse
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Provinces(ctx context.Context) ([]*model.ShipProvince, error) {
	path := "/api/v1/locations/provinces"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ShipProvince
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Districts(ctx context.Context, provinceID int) ([]*model.ShipDistrict, error) {
	path := "/api/v1/locations/districts"
	q := url.Values{}
	q.Set("province_id", strconv.Itoa(provinceID))
	h := http.Header{}
	var out []*model.ShipDistrict
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Wards(ctx context.Context, districtID int) ([]*model.ShipWard, error) {
	path := "/api/v1/locations/wards"
	q := url.Values{}
	q.Set("district_id", strconv.Itoa(districtID))
	h := http.Header{}
	var out []*model.ShipWard
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}
