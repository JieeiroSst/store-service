package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) VoucherIssue(ctx context.Context, merchantID string, productSku string, denominationAmount int, currency string, quantity int, expiresInDays *int, idempotencyKey *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: "/api/v1/vouchers/issue", Body: map[string]any{"merchant_id": merchantID, "product_sku": productSku, "denomination_amount": denominationAmount, "currency": currency, "quantity": quantity, "expires_in_days": expiresInDays, "idempotency_key": idempotencyKey}}, "")
}

func (r *mutationResolver) VoucherActivate(ctx context.Context, id string, ownerType string, ownerID string) (*model.VoucherVoucher, error) {
	return restCall[*model.VoucherVoucher](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/vouchers/%v/activate", id), Body: map[string]any{"owner_type": ownerType, "owner_id": ownerID}}, "")
}

func (r *mutationResolver) VoucherRedeem(ctx context.Context, id string, pin string, amount int, currency string, idempotencyKey *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/vouchers/%v/redeem", id), Body: map[string]any{"pin": pin, "amount": amount, "currency": currency, "idempotency_key": idempotencyKey}}, "")
}

func (r *mutationResolver) VoucherRevoke(ctx context.Context, id string) (*model.VoucherVoucher, error) {
	return restCall[*model.VoucherVoucher](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/vouchers/%v/revoke", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) VoucherCreateOrder(ctx context.Context, buyerType string, buyerID string, currency string, items []*model.VoucherOrderItemInput, idempotencyKey *string) (*model.VoucherOrder, error) {
	return restCall[*model.VoucherOrder](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: "/api/v1/orders", Body: map[string]any{"buyer_type": buyerType, "buyer_id": buyerID, "currency": currency, "items": items, "idempotency_key": idempotencyKey}}, "")
}

func (r *mutationResolver) VoucherCheckoutOrder(ctx context.Context, id string, paymentMethod string, idempotencyKey *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/orders/%v/checkout", id), Body: map[string]any{"payment_method": paymentMethod, "idempotency_key": idempotencyKey}}, "")
}

func (r *mutationResolver) VoucherCancelOrder(ctx context.Context, id string) (*model.VoucherOrder, error) {
	return restCall[*model.VoucherOrder](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/orders/%v/cancel", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) VoucherRegisterMerchant(ctx context.Context, name string, providerType string, config map[string]any) (*model.VoucherMerchant, error) {
	return restCall[*model.VoucherMerchant](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: "/api/v1/merchants", Body: map[string]any{"name": name, "provider_type": providerType, "config": config}}, "")
}

func (r *mutationResolver) VoucherRegisterCorporate(ctx context.Context, name string, taxCode *string, contactEmail *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: "/api/v1/corporates", Body: map[string]any{"name": name, "tax_code": taxCode, "contact_email": contactEmail}}, "")
}

func (r *mutationResolver) VoucherSetCorporateBudget(ctx context.Context, id string, amount int, currency string) (any, error) {
	return restCall[any](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/corporates/%v/budget", id), Body: map[string]any{"amount": amount, "currency": currency}}, "")
}

func (r *mutationResolver) VoucherCreateDistributionJob(ctx context.Context, corporateID string, merchantID string, productSku string, denominationAmount int, currency string, recipients []string) (any, error) {
	return restCall[any](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: "/api/v1/distribution-jobs", Body: map[string]any{"corporate_id": corporateID, "merchant_id": merchantID, "product_sku": productSku, "denomination_amount": denominationAmount, "currency": currency, "recipients": recipients}}, "")
}

func (r *mutationResolver) VoucherInitPayment(ctx context.Context, orderID string, amount int, currency string, provider string, returnURL *string, idempotencyKey *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: "/api/v1/payments/init", Body: map[string]any{"order_id": orderID, "amount": amount, "currency": currency, "provider": provider, "return_url": returnURL, "idempotency_key": idempotencyKey}}, "")
}

func (r *mutationResolver) VoucherImportStock(ctx context.Context, id string, file graphql.Upload) (any, error) {
	return restMultipart[any](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/merchants/%v/stock/import", id)}, map[string]any{}, uploadFiles("file", file), "")
}

func (r *mutationResolver) VoucherCreditWallet(ctx context.Context, ownerType string, ownerID string, amount int, currency string, refType *string, refID *string, idempotencyKey *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/%v/credit", ownerType, ownerID), Body: map[string]any{"amount": amount, "currency": currency, "ref_type": refType, "ref_id": refID, "idempotency_key": idempotencyKey}}, "")
}

func (r *mutationResolver) VoucherClaim(ctx context.Context, token string) (any, error) {
	return restCall[any](ctx, r.Proxy, "voucher-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/claims/%v/claim", token), Body: map[string]any{}}, "")
}

func (r *queryResolver) Voucher(ctx context.Context, id string) (*model.VoucherVoucher, error) {
	return restCall[*model.VoucherVoucher](ctx, r.Proxy, "voucher-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/vouchers/%v", id)}, "")
}

func (r *queryResolver) VoucherValidate(ctx context.Context, id string, pin *string) (*model.VoucherValidationResult, error) {
	return restCall[*model.VoucherValidationResult](ctx, r.Proxy, "voucher-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/vouchers/%v/validate", id), Query: restQuery("pin", pin)}, "")
}

func (r *queryResolver) Vouchers(ctx context.Context, ownerID *string) ([]*model.VoucherVoucher, error) {
	return restCall[[]*model.VoucherVoucher](ctx, r.Proxy, "voucher-service", rest.Request{Method: "GET", Path: "/api/v1/vouchers", Query: restQuery("owner_id", ownerID)}, "vouchers")
}

func (r *queryResolver) VoucherOrder(ctx context.Context, id string) (*model.VoucherOrder, error) {
	return restCall[*model.VoucherOrder](ctx, r.Proxy, "voucher-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v", id)}, "")
}

func (r *queryResolver) VoucherMerchant(ctx context.Context, id string) (*model.VoucherMerchant, error) {
	return restCall[*model.VoucherMerchant](ctx, r.Proxy, "voucher-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/merchants/%v", id)}, "")
}

func (r *queryResolver) VoucherMerchants(ctx context.Context) ([]*model.VoucherMerchant, error) {
	return restCall[[]*model.VoucherMerchant](ctx, r.Proxy, "voucher-service", rest.Request{Method: "GET", Path: "/api/v1/merchants"}, "merchants")
}

func (r *queryResolver) VoucherWalletBalance(ctx context.Context, ownerType string, ownerID string) (*model.VoucherBalance, error) {
	return restCall[*model.VoucherBalance](ctx, r.Proxy, "voucher-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/%v/balance", ownerType, ownerID)}, "")
}

func (r *queryResolver) VoucherRedemptionReports(ctx context.Context, from *string, to *string) ([]*model.VoucherRedemptionReport, error) {
	return restCall[[]*model.VoucherRedemptionReport](ctx, r.Proxy, "voucher-service", rest.Request{Method: "GET", Path: "/api/v1/reports/redemptions", Query: restQuery("from", from, "to", to)}, "reports")
}

func (r *queryResolver) VoucherCorporateSpend(ctx context.Context, id string, from *string, to *string) (*model.VoucherCorporateSpendReport, error) {
	return restCall[*model.VoucherCorporateSpendReport](ctx, r.Proxy, "voucher-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/reports/corporates/%v/spend", id), Query: restQuery("from", from, "to", to)}, "")
}
