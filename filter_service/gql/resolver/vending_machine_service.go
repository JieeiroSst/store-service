package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) VendQuery() generated.VendQueryResolver { return &vendQueryResolver{r} }

type vendQueryResolver struct{ *Resolver }

func (r *vendQueryResolver) Machines(ctx context.Context, obj *model.VendQuery, status *string, cursor *string, limit *int) (*model.VendMachinePage, error) {
	return r.Clients.VendingMachineService.Machines(ctx, status, cursor, limit)
}

func (r *vendQueryResolver) Machine(ctx context.Context, obj *model.VendQuery, id string) (*model.VendMachine, error) {
	return r.Clients.VendingMachineService.Machine(ctx, id)
}

func (r *vendQueryResolver) Maintenance(ctx context.Context, obj *model.VendQuery, id string) ([]*model.VendMaintenance, error) {
	return r.Clients.VendingMachineService.Maintenance(ctx, id)
}

func (r *vendQueryResolver) Events(ctx context.Context, obj *model.VendQuery, id string, limit *int) ([]*model.VendEvent, error) {
	return r.Clients.VendingMachineService.Events(ctx, id, limit)
}

func (r *vendQueryResolver) SalesReport(ctx context.Context, obj *model.VendQuery, id string, from *string, to *string) (*model.VendSalesReport, error) {
	return r.Clients.VendingMachineService.SalesReport(ctx, id, from, to)
}

func (r *vendQueryResolver) Inventory(ctx context.Context, obj *model.VendQuery, id string, cursor *string, limit *int) (*model.VendInventoryPage, error) {
	return r.Clients.VendingMachineService.Inventory(ctx, id, cursor, limit)
}

func (r *vendQueryResolver) LowInventory(ctx context.Context, obj *model.VendQuery, machineID *string, cursor *string, limit *int) (*model.VendInventoryPage, error) {
	return r.Clients.VendingMachineService.LowInventory(ctx, machineID, cursor, limit)
}

func (r *vendQueryResolver) Categories(ctx context.Context, obj *model.VendQuery, cursor *string, limit *int) (*model.VendCategoryPage, error) {
	return r.Clients.VendingMachineService.Categories(ctx, cursor, limit)
}

func (r *vendQueryResolver) Products(ctx context.Context, obj *model.VendQuery, categoryID *string, cursor *string, limit *int) (*model.VendProductPage, error) {
	return r.Clients.VendingMachineService.Products(ctx, categoryID, cursor, limit)
}

func (r *vendQueryResolver) Product(ctx context.Context, obj *model.VendQuery, id string) (*model.VendProduct, error) {
	return r.Clients.VendingMachineService.Product(ctx, id)
}

func (r *vendQueryResolver) Session(ctx context.Context, obj *model.VendQuery, id string) (*model.VendSession, error) {
	return r.Clients.VendingMachineService.Session(ctx, id)
}

func (r *vendQueryResolver) SessionOrders(ctx context.Context, obj *model.VendQuery, id string) ([]*model.VendOrder, error) {
	return r.Clients.VendingMachineService.SessionOrders(ctx, id)
}

func (r *vendQueryResolver) Payment(ctx context.Context, obj *model.VendQuery, id string) (*model.VendPaymentStatus, error) {
	return r.Clients.VendingMachineService.Payment(ctx, id)
}

func (r *vendQueryResolver) Order(ctx context.Context, obj *model.VendQuery, id string) (*model.VendOrder, error) {
	return r.Clients.VendingMachineService.Order(ctx, id)
}
