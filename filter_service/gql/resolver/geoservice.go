package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) GeoQuery() generated.GeoQueryResolver { return &geoQueryResolver{r} }

type geoQueryResolver struct{ *Resolver }

func (r *geoQueryResolver) Locate(ctx context.Context, obj *model.GeoQuery, lng float64, lat float64, tolerance *float64) (*model.GeoLocationResult, error) {
	return r.Clients.Geoservice.Locate(ctx, lng, lat, tolerance)
}

func (r *geoQueryResolver) Provinces(ctx context.Context, obj *model.GeoQuery) ([]*model.GeoProvince, error) {
	return r.Clients.Geoservice.Provinces(ctx)
}

func (r *geoQueryResolver) Province(ctx context.Context, obj *model.GeoQuery, code string) (*model.GeoProvince, error) {
	return r.Clients.Geoservice.Province(ctx, code)
}

func (r *geoQueryResolver) Districts(ctx context.Context, obj *model.GeoQuery, code string) ([]*model.GeoDistrict, error) {
	return r.Clients.Geoservice.Districts(ctx, code)
}

func (r *geoQueryResolver) Wards(ctx context.Context, obj *model.GeoQuery, code string) ([]*model.GeoWard, error) {
	return r.Clients.Geoservice.Wards(ctx, code)
}

func (r *geoQueryResolver) Streets(ctx context.Context, obj *model.GeoQuery, code string) ([]*model.GeoStreet, error) {
	return r.Clients.Geoservice.Streets(ctx, code)
}

func (r *geoQueryResolver) Stats(ctx context.Context, obj *model.GeoQuery) (*model.GeoStats, error) {
	return r.Clients.Geoservice.Stats(ctx)
}

func (r *geoQueryResolver) WhitelistCheck(ctx context.Context, obj *model.GeoQuery, lng float64, lat float64, tolerance *float64) (*model.GeoWhitelistResult, error) {
	return r.Clients.Geoservice.WhitelistCheck(ctx, lng, lat, tolerance)
}

func (r *geoQueryResolver) WhitelistCities(ctx context.Context, obj *model.GeoQuery) ([]*model.GeoWhitelistCity, error) {
	return r.Clients.Geoservice.WhitelistCities(ctx)
}
