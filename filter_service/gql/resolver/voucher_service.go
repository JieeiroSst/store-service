package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) VoucherQuery() generated.VoucherQueryResolver { return &voucherQueryResolver{r} }

type voucherQueryResolver struct{ *Resolver }

func (r *voucherQueryResolver) Voucher(ctx context.Context, obj *model.VoucherQuery, id string) (*model.VoucherVoucher, error) {
	return r.Clients.VoucherService.Voucher(ctx, id)
}

func (r *voucherQueryResolver) Validate(ctx context.Context, obj *model.VoucherQuery, id string, pin *string) (*model.VoucherValidationResult, error) {
	return r.Clients.VoucherService.Validate(ctx, id, pin)
}

func (r *voucherQueryResolver) Vouchers(ctx context.Context, obj *model.VoucherQuery, ownerID *string) ([]*model.VoucherVoucher, error) {
	return r.Clients.VoucherService.Vouchers(ctx, ownerID)
}

func (r *voucherQueryResolver) Order(ctx context.Context, obj *model.VoucherQuery, id string) (*model.VoucherOrder, error) {
	return r.Clients.VoucherService.Order(ctx, id)
}

func (r *voucherQueryResolver) Merchant(ctx context.Context, obj *model.VoucherQuery, id string) (*model.VoucherMerchant, error) {
	return r.Clients.VoucherService.Merchant(ctx, id)
}

func (r *voucherQueryResolver) Merchants(ctx context.Context, obj *model.VoucherQuery) ([]*model.VoucherMerchant, error) {
	return r.Clients.VoucherService.Merchants(ctx)
}

func (r *voucherQueryResolver) WalletBalance(ctx context.Context, obj *model.VoucherQuery, ownerType string, ownerID string) (*model.VoucherBalance, error) {
	return r.Clients.VoucherService.WalletBalance(ctx, ownerType, ownerID)
}

func (r *voucherQueryResolver) RedemptionReports(ctx context.Context, obj *model.VoucherQuery, from *string, to *string) ([]*model.VoucherRedemptionReport, error) {
	return r.Clients.VoucherService.RedemptionReports(ctx, from, to)
}

func (r *voucherQueryResolver) CorporateSpend(ctx context.Context, obj *model.VoucherQuery, id string, from *string, to *string) (*model.VoucherCorporateSpendReport, error) {
	return r.Clients.VoucherService.CorporateSpend(ctx, id, from, to)
}
