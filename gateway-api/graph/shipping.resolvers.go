package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ShipCreateShipment(ctx context.Context, input model.ShipCreateShipmentInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "shipping-service", rest.Request{Method: "POST", Path: "/api/v1/shipments", Body: input}, "")
}

func (r *mutationResolver) ShipPlaceShipment(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "shipping-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/shipments/%v/place", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ShipCancelShipment(ctx context.Context, id int, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "shipping-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/shipments/%v/cancel", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) ShipSyncShipment(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "shipping-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/shipments/%v/sync", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ShipCreateWarehouse(ctx context.Context, code string, name string, phone *string, street *string, wardCode *string, districtID *int, provinceID *int, carrierShopID *int) (*model.ShipWarehouse, error) {
	return restCall[*model.ShipWarehouse](ctx, r.Proxy, "shipping-service", rest.Request{Method: "POST", Path: "/api/v1/warehouses", Body: map[string]any{"code": code, "name": name, "phone": phone, "street": street, "ward_code": wardCode, "district_id": districtID, "province_id": provinceID, "carrier_shop_id": carrierShopID}}, "")
}

func (r *mutationResolver) ShipSetWarehouseActive(ctx context.Context, id int, active bool) (*model.ShipWarehouse, error) {
	return restCall[*model.ShipWarehouse](ctx, r.Proxy, "shipping-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/warehouses/%v", id), Body: map[string]any{"active": active}}, "")
}

func (r *queryResolver) Shipments(ctx context.Context, status *string, limit *int, offset *int) (*model.ShipShipmentList, error) {
	return restCall[*model.ShipShipmentList](ctx, r.Proxy, "shipping-service", rest.Request{Method: "GET", Path: "/api/v1/shipments", Query: restQuery("status", status, "limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) Shipment(ctx context.Context, id int) (*model.ShipShipmentDetail, error) {
	return restCall[*model.ShipShipmentDetail](ctx, r.Proxy, "shipping-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/shipments/%v", id)}, "")
}

func (r *queryResolver) ShipmentByClientCode(ctx context.Context, code string) (*model.ShipShipmentDetail, error) {
	return restCall[*model.ShipShipmentDetail](ctx, r.Proxy, "shipping-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/shipments/by-client-code/%v", code)}, "")
}

func (r *queryResolver) ShipWarehouses(ctx context.Context) ([]*model.ShipWarehouse, error) {
	return restCall[[]*model.ShipWarehouse](ctx, r.Proxy, "shipping-service", rest.Request{Method: "GET", Path: "/api/v1/warehouses"}, "items")
}

func (r *queryResolver) ShipProvinces(ctx context.Context) ([]*model.ShipProvince, error) {
	return restCall[[]*model.ShipProvince](ctx, r.Proxy, "shipping-service", rest.Request{Method: "GET", Path: "/api/v1/locations/provinces"}, "items")
}

func (r *queryResolver) ShipDistricts(ctx context.Context, provinceID int) ([]*model.ShipDistrict, error) {
	return restCall[[]*model.ShipDistrict](ctx, r.Proxy, "shipping-service", rest.Request{Method: "GET", Path: "/api/v1/locations/districts", Query: restQuery("province_id", provinceID)}, "items")
}

func (r *queryResolver) ShipWards(ctx context.Context, districtID int) ([]*model.ShipWard, error) {
	return restCall[[]*model.ShipWard](ctx, r.Proxy, "shipping-service", rest.Request{Method: "GET", Path: "/api/v1/locations/wards", Query: restQuery("district_id", districtID)}, "items")
}

func (r *queryResolver) ShipQuote(ctx context.Context, input model.ShipQuoteInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "shipping-service", rest.Request{Method: "POST", Path: "/api/v1/shipments/quote", Body: input}, "")
}
