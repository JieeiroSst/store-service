package geoservice

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "geoservice"

const DefaultBaseURL = "http://geoservice-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Locate(ctx context.Context, lng float64, lat float64, tolerance *float64) (*model.GeoLocationResult, error) {
	path := "/api/v1/locate"
	q := url.Values{}
	q.Set("lng", strconv.FormatFloat(lng, 'f', -1, 64))
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	if tolerance != nil {
		q.Set("tolerance", strconv.FormatFloat(*tolerance, 'f', -1, 64))
	}
	h := http.Header{}
	var out *model.GeoLocationResult
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Provinces(ctx context.Context) ([]*model.GeoProvince, error) {
	path := "/api/v1/provinces"
	q := url.Values{}
	h := http.Header{}
	var out []*model.GeoProvince
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Province(ctx context.Context, code string) (*model.GeoProvince, error) {
	path := "/api/v1/provinces/" + url.PathEscape(code)
	q := url.Values{}
	h := http.Header{}
	var out *model.GeoProvince
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Districts(ctx context.Context, code string) ([]*model.GeoDistrict, error) {
	path := "/api/v1/provinces/" + url.PathEscape(code) + "/districts"
	q := url.Values{}
	h := http.Header{}
	var out []*model.GeoDistrict
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Wards(ctx context.Context, code string) ([]*model.GeoWard, error) {
	path := "/api/v1/districts/" + url.PathEscape(code) + "/wards"
	q := url.Values{}
	h := http.Header{}
	var out []*model.GeoWard
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Streets(ctx context.Context, code string) ([]*model.GeoStreet, error) {
	path := "/api/v1/provinces/" + url.PathEscape(code) + "/streets"
	q := url.Values{}
	h := http.Header{}
	var out []*model.GeoStreet
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Stats(ctx context.Context) (*model.GeoStats, error) {
	path := "/api/v1/stats"
	q := url.Values{}
	h := http.Header{}
	var out *model.GeoStats
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) WhitelistCheck(ctx context.Context, lng float64, lat float64, tolerance *float64) (*model.GeoWhitelistResult, error) {
	path := "/api/v1/whitelist/check"
	q := url.Values{}
	q.Set("lng", strconv.FormatFloat(lng, 'f', -1, 64))
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	if tolerance != nil {
		q.Set("tolerance", strconv.FormatFloat(*tolerance, 'f', -1, 64))
	}
	h := http.Header{}
	var out *model.GeoWhitelistResult
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) WhitelistCities(ctx context.Context) ([]*model.GeoWhitelistCity, error) {
	path := "/api/v1/whitelist/cities"
	q := url.Values{}
	h := http.Header{}
	var out []*model.GeoWhitelistCity
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
