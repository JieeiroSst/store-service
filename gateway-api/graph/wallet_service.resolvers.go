package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) WalletSvcCreateWallet(ctx context.Context, userID string, currency string) (*model.WalletSvcWallet, error) {
	return restCall[*model.WalletSvcWallet](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: "/api/v1/wallets", Body: map[string]any{"user_id": userID, "currency": currency}}, "")
}

func (r *mutationResolver) WalletSvcCredit(ctx context.Context, id string, amount string, description *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/credit", id), Body: map[string]any{"amount": amount, "description": description}}, "")
}

func (r *mutationResolver) WalletSvcFreeze(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/freeze", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) WalletSvcAuthorize(ctx context.Context, idempotencyKey string, cardID string, merchantID string, amount string, currency string, description *string, metadata map[string]any) (*model.WalletSvcTransaction, error) {
	return restCall[*model.WalletSvcTransaction](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: "/api/v1/transactions/authorize", Body: map[string]any{"idempotency_key": idempotencyKey, "card_id": cardID, "merchant_id": merchantID, "amount": amount, "currency": currency, "description": description, "metadata": metadata}, Header: restHeader("Idempotency-Key", idempotencyKey)}, "")
}

func (r *mutationResolver) WalletSvcCapture(ctx context.Context, id string) (*model.WalletSvcTransaction, error) {
	return restCall[*model.WalletSvcTransaction](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/transactions/%v/capture", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) WalletSvcVoid(ctx context.Context, id string) (*model.WalletSvcTransaction, error) {
	return restCall[*model.WalletSvcTransaction](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/transactions/%v/void", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) WalletSvcCreateSettlementBatch(ctx context.Context, merchantID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: "/api/v1/settlements/batches", Body: map[string]any{"merchant_id": merchantID}}, "")
}

func (r *mutationResolver) WalletSvcProcessClearing(ctx context.Context, batchID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/settlements/batches/%v/clear", batchID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) WalletSvcProcessSettlement(ctx context.Context, batchID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/settlements/batches/%v/settle", batchID), Body: map[string]any{}}, "")
}

func (r *queryResolver) WalletSvcWallet(ctx context.Context, id string) (*model.WalletSvcWallet, error) {
	return restCall[*model.WalletSvcWallet](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", id)}, "")
}

func (r *queryResolver) WalletSvcTransaction(ctx context.Context, id string) (*model.WalletSvcTransaction, error) {
	return restCall[*model.WalletSvcTransaction](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/transactions/%v", id)}, "")
}
