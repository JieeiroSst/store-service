package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) VendRegisterMachine(ctx context.Context, location string, machineModel string, status *string) (*model.VendMachine, error) {
	return restCall[*model.VendMachine](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: "/api/v1/machines", Body: map[string]any{"location": location, "model": machineModel, "status": status}}, "")
}

func (r *mutationResolver) VendUpdateMachine(ctx context.Context, id string, location *string, machineModel *string) (*model.VendMachine, error) {
	return restCall[*model.VendMachine](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/machines/%v", id), Body: map[string]any{"location": location, "model": machineModel}}, "")
}

func (r *mutationResolver) VendChangeMachineStatus(ctx context.Context, id string, status string) (*model.VendMachine, error) {
	return restCall[*model.VendMachine](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/machines/%v/status", id), Body: map[string]any{"status": status}}, "")
}

func (r *mutationResolver) VendRecordMaintenance(ctx context.Context, id string, technicianID string, maintenanceType string, notes *string) (*model.VendMaintenance, error) {
	return restCall[*model.VendMaintenance](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/machines/%v/maintenance", id), Body: map[string]any{"technician_id": technicianID, "maintenance_type": maintenanceType, "notes": notes}}, "")
}

func (r *mutationResolver) VendAssignSlot(ctx context.Context, id string, slot string, input model.VendSlotInput) (*model.VendInventory, error) {
	return restCall[*model.VendInventory](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/machines/%v/slots/%v", id, slot), Body: input}, "")
}

func (r *mutationResolver) VendRestock(ctx context.Context, id string, quantity int) (*model.VendInventory, error) {
	return restCall[*model.VendInventory](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/inventory/%v/restock", id), Body: map[string]any{"quantity": quantity}}, "")
}

func (r *mutationResolver) VendCreateCategory(ctx context.Context, name string, description *string, displayOrder *int) (*model.VendCategory, error) {
	return restCall[*model.VendCategory](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: "/api/v1/categories", Body: map[string]any{"name": name, "description": description, "display_order": displayOrder}}, "")
}

func (r *mutationResolver) VendCreateProduct(ctx context.Context, input model.VendProductInput) (*model.VendProduct, error) {
	return restCall[*model.VendProduct](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: "/api/v1/products", Body: input}, "")
}

func (r *mutationResolver) VendUpdateProduct(ctx context.Context, id string, input model.VendProductPatchInput) (*model.VendProduct, error) {
	return restCall[*model.VendProduct](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/products/%v", id), Body: input}, "")
}

func (r *mutationResolver) VendStartSession(ctx context.Context, machineID string) (*model.VendSession, error) {
	return restCall[*model.VendSession](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: "/api/v1/sessions", Body: map[string]any{"machine_id": machineID}}, "")
}

func (r *mutationResolver) VendEndSession(ctx context.Context, id string) (*model.VendSession, error) {
	return restCall[*model.VendSession](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/sessions/%v/end", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) VendReserve(ctx context.Context, id string, slot string) (any, error) {
	return restCall[any](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/sessions/%v/reservations", id), Body: map[string]any{"slot": slot}}, "")
}

func (r *mutationResolver) VendCancelReservation(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/reservations/%v", id)}, "")
}

func (r *mutationResolver) VendCheckout(ctx context.Context, id string, paymentMethod string, walletID *string, couponCode *string) (*model.VendOrder, error) {
	return restCall[*model.VendOrder](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/reservations/%v/checkout", id), Body: map[string]any{"payment_method": paymentMethod, "wallet_id": walletID, "coupon_code": couponCode}}, "")
}

func (r *mutationResolver) VendReportDispense(ctx context.Context, id string, dispensed bool) (*model.VendOrder, error) {
	return restCall[*model.VendOrder](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/orders/%v/dispense", id), Body: map[string]any{"dispensed": dispensed}}, "")
}

func (r *mutationResolver) VendRefundOrder(ctx context.Context, id string, reason *string) (*model.VendOrder, error) {
	return restCall[*model.VendOrder](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/orders/%v/refund", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *queryResolver) VendMachines(ctx context.Context, status *string, cursor *string, limit *int) (*model.VendMachinePage, error) {
	return restCall[*model.VendMachinePage](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: "/api/v1/machines", Query: restQuery("status", status, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) VendMachine(ctx context.Context, id string) (*model.VendMachine, error) {
	return restCall[*model.VendMachine](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/machines/%v", id)}, "")
}

func (r *queryResolver) VendMaintenance(ctx context.Context, id string) ([]*model.VendMaintenance, error) {
	return restCall[[]*model.VendMaintenance](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/machines/%v/maintenance", id)}, "")
}

func (r *queryResolver) VendEvents(ctx context.Context, id string, limit *int) ([]*model.VendEvent, error) {
	return restCall[[]*model.VendEvent](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/machines/%v/events", id), Query: restQuery("limit", limit)}, "")
}

func (r *queryResolver) VendSalesReport(ctx context.Context, id string, from *string, to *string) (*model.VendSalesReport, error) {
	return restCall[*model.VendSalesReport](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/machines/%v/sales", id), Query: restQuery("from", from, "to", to)}, "")
}

func (r *queryResolver) VendInventory(ctx context.Context, id string, cursor *string, limit *int) (*model.VendInventoryPage, error) {
	return restCall[*model.VendInventoryPage](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/machines/%v/inventory", id), Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) VendLowInventory(ctx context.Context, machineID *string, cursor *string, limit *int) (*model.VendInventoryPage, error) {
	return restCall[*model.VendInventoryPage](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: "/api/v1/inventory/low", Query: restQuery("machine_id", machineID, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) VendCategories(ctx context.Context, cursor *string, limit *int) (*model.VendCategoryPage, error) {
	return restCall[*model.VendCategoryPage](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: "/api/v1/categories", Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) VendProducts(ctx context.Context, categoryID *string, cursor *string, limit *int) (*model.VendProductPage, error) {
	return restCall[*model.VendProductPage](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: "/api/v1/products", Query: restQuery("category_id", categoryID, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) VendProduct(ctx context.Context, id string) (*model.VendProduct, error) {
	return restCall[*model.VendProduct](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/products/%v", id)}, "")
}

func (r *queryResolver) VendSession(ctx context.Context, id string) (*model.VendSession, error) {
	return restCall[*model.VendSession](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/sessions/%v", id)}, "")
}

func (r *queryResolver) VendSessionOrders(ctx context.Context, id string) ([]*model.VendOrder, error) {
	return restCall[[]*model.VendOrder](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/sessions/%v/orders", id)}, "")
}

func (r *queryResolver) VendPayment(ctx context.Context, id string) (*model.VendPaymentStatus, error) {
	return restCall[*model.VendPaymentStatus](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/payments/%v", id)}, "")
}

func (r *queryResolver) VendOrder(ctx context.Context, id string) (*model.VendOrder, error) {
	return restCall[*model.VendOrder](ctx, r.Proxy, "vending-machine-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v", id)}, "")
}
