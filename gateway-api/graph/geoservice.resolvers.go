package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *queryResolver) GeoLocate(ctx context.Context, lng float64, lat float64, tolerance *float64) (*model.GeoLocationResult, error) {
	return restCall[*model.GeoLocationResult](ctx, r.Proxy, "geoservice", rest.Request{Method: "GET", Path: "/api/v1/locate", Query: restQuery("lng", lng, "lat", lat, "tolerance", tolerance)}, "")
}

func (r *queryResolver) GeoProvinces(ctx context.Context) ([]*model.GeoProvince, error) {
	return restCall[[]*model.GeoProvince](ctx, r.Proxy, "geoservice", rest.Request{Method: "GET", Path: "/api/v1/provinces"}, "")
}

func (r *queryResolver) GeoProvince(ctx context.Context, code string) (*model.GeoProvince, error) {
	return restCall[*model.GeoProvince](ctx, r.Proxy, "geoservice", rest.Request{Method: "GET", Path: rest.P("/api/v1/provinces/%v", code)}, "")
}

func (r *queryResolver) GeoDistricts(ctx context.Context, code string) ([]*model.GeoDistrict, error) {
	return restCall[[]*model.GeoDistrict](ctx, r.Proxy, "geoservice", rest.Request{Method: "GET", Path: rest.P("/api/v1/provinces/%v/districts", code)}, "")
}

func (r *queryResolver) GeoWards(ctx context.Context, code string) ([]*model.GeoWard, error) {
	return restCall[[]*model.GeoWard](ctx, r.Proxy, "geoservice", rest.Request{Method: "GET", Path: rest.P("/api/v1/districts/%v/wards", code)}, "")
}

func (r *queryResolver) GeoStreets(ctx context.Context, code string) ([]*model.GeoStreet, error) {
	return restCall[[]*model.GeoStreet](ctx, r.Proxy, "geoservice", rest.Request{Method: "GET", Path: rest.P("/api/v1/provinces/%v/streets", code)}, "")
}

func (r *queryResolver) GeoStats(ctx context.Context) (*model.GeoStats, error) {
	return restCall[*model.GeoStats](ctx, r.Proxy, "geoservice", rest.Request{Method: "GET", Path: "/api/v1/stats"}, "")
}

func (r *queryResolver) GeoWhitelistCheck(ctx context.Context, lng float64, lat float64, tolerance *float64) (*model.GeoWhitelistResult, error) {
	return restCall[*model.GeoWhitelistResult](ctx, r.Proxy, "geoservice", rest.Request{Method: "GET", Path: "/api/v1/whitelist/check", Query: restQuery("lng", lng, "lat", lat, "tolerance", tolerance)}, "")
}

func (r *queryResolver) GeoWhitelistCities(ctx context.Context) ([]*model.GeoWhitelistCity, error) {
	return restCall[[]*model.GeoWhitelistCity](ctx, r.Proxy, "geoservice", rest.Request{Method: "GET", Path: "/api/v1/whitelist/cities"}, "")
}
