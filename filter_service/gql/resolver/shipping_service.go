package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ShipQuery() generated.ShipQueryResolver { return &shipQueryResolver{r} }

type shipQueryResolver struct{ *Resolver }

func (r *shipQueryResolver) Shipments(ctx context.Context, obj *model.ShipQuery, status *string, limit *int, offset *int) (*model.ShipShipmentList, error) {
	return r.Clients.ShippingService.Shipments(ctx, status, limit, offset)
}

func (r *shipQueryResolver) Shipment(ctx context.Context, obj *model.ShipQuery, id int) (*model.ShipShipmentDetail, error) {
	return r.Clients.ShippingService.Shipment(ctx, id)
}

func (r *shipQueryResolver) ShipmentByClientCode(ctx context.Context, obj *model.ShipQuery, code string) (*model.ShipShipmentDetail, error) {
	return r.Clients.ShippingService.ShipmentByClientCode(ctx, code)
}

func (r *shipQueryResolver) Warehouses(ctx context.Context, obj *model.ShipQuery) ([]*model.ShipWarehouse, error) {
	return r.Clients.ShippingService.Warehouses(ctx)
}

func (r *shipQueryResolver) Provinces(ctx context.Context, obj *model.ShipQuery) ([]*model.ShipProvince, error) {
	return r.Clients.ShippingService.Provinces(ctx)
}

func (r *shipQueryResolver) Districts(ctx context.Context, obj *model.ShipQuery, provinceID int) ([]*model.ShipDistrict, error) {
	return r.Clients.ShippingService.Districts(ctx, provinceID)
}

func (r *shipQueryResolver) Wards(ctx context.Context, obj *model.ShipQuery, districtID int) ([]*model.ShipWard, error) {
	return r.Clients.ShippingService.Wards(ctx, districtID)
}
