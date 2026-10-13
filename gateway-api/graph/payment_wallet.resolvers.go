package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) CreateMyWallet(ctx context.Context, currency string) (*model.WalletWallet, error) {
	return restCall[*model.WalletWallet](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: "/api/v1/wallets", Body: withField(map[string]any{"currency": currency}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) CreateMyPocket(ctx context.Context, id string, name string) (*model.WalletPocket, error) {
	return restOwned[*model.WalletPocket](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/pockets", id), Body: map[string]any{"name": name}}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", id)}, "")
}

func (r *mutationResolver) TransferFromMyWallet(ctx context.Context, senderWalletID string, receiverWalletID string, amount int, referenceID string, description *string) (*model.WalletTransfer, error) {
	return restOwned[*model.WalletTransfer](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: "/api/v1/transfers", Body: map[string]any{"sender_wallet_id": senderWalletID, "receiver_wallet_id": receiverWalletID, "amount": amount, "reference_id": referenceID, "description": description}}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", senderWalletID)}, "")
}

func (r *mutationResolver) CreateMyPaymentRequest(ctx context.Context, requesterWalletID string, payerWalletID *string, amount int, description *string, expiresInSeconds *int) (*model.WalletPaymentRequest, error) {
	return restOwned[*model.WalletPaymentRequest](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: "/api/v1/payment-requests", Body: map[string]any{"requester_wallet_id": requesterWalletID, "payer_wallet_id": payerWalletID, "amount": amount, "description": description, "expires_in_seconds": expiresInSeconds}}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", requesterWalletID)}, "")
}

func (r *mutationResolver) PayPaymentRequestFromMyWallet(ctx context.Context, id string, payerWalletID string) (any, error) {
	return restOwned[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/payment-requests/%v/pay", id), Body: map[string]any{"payer_wallet_id": payerWalletID}}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", payerWalletID)}, "")
}

func (r *mutationResolver) AddMyPaymentMethod(ctx context.Context, typeArg string, provider string, accountNumber string) (*model.WalletPaymentMethod, error) {
	return restCall[*model.WalletPaymentMethod](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: "/api/v1/payment-methods", Body: withField(map[string]any{"type": typeArg, "provider": provider, "account_number": accountNumber}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) WalletCreateWallet(ctx context.Context, userID string, currency string) (*model.WalletWallet, error) {
	return restCall[*model.WalletWallet](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: "/api/v1/wallets", Body: map[string]any{"user_id": userID, "currency": currency}}, "")
}

func (r *mutationResolver) WalletDeposit(ctx context.Context, id string, amount int, referenceID string, description *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/deposit", id), Body: map[string]any{"amount": amount, "reference_id": referenceID, "description": description}}, "")
}

func (r *mutationResolver) WalletWithdraw(ctx context.Context, id string, amount int, referenceID string, description *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/withdraw", id), Body: map[string]any{"amount": amount, "reference_id": referenceID, "description": description}}, "")
}

func (r *mutationResolver) WalletFreeze(ctx context.Context, id string, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/freeze", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) WalletUnfreeze(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/unfreeze", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) WalletClose(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/close", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) WalletSetLimits(ctx context.Context, id string, dailyLimit int, perTransactionLimit int) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/wallets/%v/limits", id), Body: map[string]any{"daily_limit": dailyLimit, "per_transaction_limit": perTransactionLimit}}, "")
}

func (r *mutationResolver) WalletCreatePocket(ctx context.Context, id string, name string) (*model.WalletPocket, error) {
	return restCall[*model.WalletPocket](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/pockets", id), Body: map[string]any{"name": name}}, "")
}

func (r *mutationResolver) WalletDepositToPocket(ctx context.Context, pocketID string, amount int) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/pockets/%v/deposit", pocketID), Body: map[string]any{"amount": amount}}, "")
}

func (r *mutationResolver) WalletWithdrawFromPocket(ctx context.Context, pocketID string, amount int) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/pockets/%v/withdraw", pocketID), Body: map[string]any{"amount": amount}}, "")
}

func (r *mutationResolver) WalletClosePocket(ctx context.Context, pocketID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/pockets/%v", pocketID)}, "")
}

func (r *mutationResolver) WalletTransfer(ctx context.Context, senderWalletID string, receiverWalletID string, amount int, referenceID string, description *string) (*model.WalletTransfer, error) {
	return restCall[*model.WalletTransfer](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: "/api/v1/transfers", Body: map[string]any{"sender_wallet_id": senderWalletID, "receiver_wallet_id": receiverWalletID, "amount": amount, "reference_id": referenceID, "description": description}}, "")
}

func (r *mutationResolver) WalletReverseTransfer(ctx context.Context, id string, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/transfers/%v/reverse", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) WalletReverseTransaction(ctx context.Context, id string, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/transactions/%v/reverse", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) WalletCreatePaymentRequest(ctx context.Context, requesterWalletID string, payerWalletID *string, amount int, description *string, expiresInSeconds *int) (*model.WalletPaymentRequest, error) {
	return restCall[*model.WalletPaymentRequest](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: "/api/v1/payment-requests", Body: map[string]any{"requester_wallet_id": requesterWalletID, "payer_wallet_id": payerWalletID, "amount": amount, "description": description, "expires_in_seconds": expiresInSeconds}}, "")
}

func (r *mutationResolver) WalletPayPaymentRequest(ctx context.Context, id string, payerWalletID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/payment-requests/%v/pay", id), Body: map[string]any{"payer_wallet_id": payerWalletID}}, "")
}

func (r *mutationResolver) WalletCancelPaymentRequest(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/payment-requests/%v/cancel", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) WalletAddPaymentMethod(ctx context.Context, userID string, typeArg string, provider string, accountNumber string) (*model.WalletPaymentMethod, error) {
	return restCall[*model.WalletPaymentMethod](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "POST", Path: "/api/v1/payment-methods", Body: map[string]any{"user_id": userID, "type": typeArg, "provider": provider, "account_number": accountNumber}}, "")
}

func (r *mutationResolver) WalletRemovePaymentMethod(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/payment-methods/%v", id)}, "")
}

func (r *mutationResolver) WalletSetDefaultPaymentMethod(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/payment-methods/%v/default", id), Body: map[string]any{}}, "")
}

func (r *queryResolver) Wallet(ctx context.Context, id string) (*model.WalletWallet, error) {
	return restCall[*model.WalletWallet](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", id)}, "")
}

func (r *queryResolver) WalletByUser(ctx context.Context, userID string) (*model.WalletWallet, error) {
	return restCall[*model.WalletWallet](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/user/%v", userID)}, "")
}

func (r *queryResolver) WalletTransactions(ctx context.Context, id string, limit *int, offset *int) ([]*model.WalletTransaction, error) {
	return restCall[[]*model.WalletTransaction](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/transactions", id), Query: restQuery("limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) WalletTransactionByReference(ctx context.Context, id string, referenceID string) (*model.WalletTransaction, error) {
	return restCall[*model.WalletTransaction](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/transactions/by-reference/%v", id, referenceID)}, "")
}

func (r *queryResolver) WalletPockets(ctx context.Context, id string) ([]*model.WalletPocket, error) {
	return restCall[[]*model.WalletPocket](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/pockets", id)}, "")
}

func (r *queryResolver) WalletTransfer(ctx context.Context, id string) (*model.WalletTransfer, error) {
	return restCall[*model.WalletTransfer](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/transfers/%v", id)}, "")
}

func (r *queryResolver) WalletTransaction(ctx context.Context, id string) (*model.WalletTransaction, error) {
	return restCall[*model.WalletTransaction](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/transactions/%v", id)}, "")
}

func (r *queryResolver) WalletPaymentRequests(ctx context.Context, walletID string) ([]*model.WalletPaymentRequest, error) {
	return restCall[[]*model.WalletPaymentRequest](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: "/api/v1/payment-requests", Query: restQuery("wallet_id", walletID)}, "")
}

func (r *queryResolver) WalletPaymentRequest(ctx context.Context, id string) (*model.WalletPaymentRequest, error) {
	return restCall[*model.WalletPaymentRequest](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/payment-requests/%v", id)}, "")
}

func (r *queryResolver) WalletPaymentRequestQR(ctx context.Context, id string) (*model.WalletPaymentRequestQR, error) {
	return restCall[*model.WalletPaymentRequestQR](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/payment-requests/%v/qr", id)}, "")
}

func (r *queryResolver) WalletPaymentMethods(ctx context.Context, userID string) ([]*model.WalletPaymentMethod, error) {
	return restCall[[]*model.WalletPaymentMethod](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: "/api/v1/payment-methods", Query: restQuery("user_id", userID)}, "")
}

func (r *queryResolver) MyWallet(ctx context.Context) (*model.WalletWallet, error) {
	return restCall[*model.WalletWallet](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/user/%v", userIDOf(ctx))}, "")
}

func (r *queryResolver) MyWalletTransactions(ctx context.Context, id string, limit *int, offset *int) ([]*model.WalletTransaction, error) {
	return restOwned[[]*model.WalletTransaction](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/transactions", id), Query: restQuery("limit", limit, "offset", offset)}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", id)}, "")
}

func (r *queryResolver) MyWalletPockets(ctx context.Context, id string) ([]*model.WalletPocket, error) {
	return restOwned[[]*model.WalletPocket](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/pockets", id)}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", id)}, "")
}

func (r *queryResolver) MyPaymentMethods(ctx context.Context) ([]*model.WalletPaymentMethod, error) {
	return restCall[[]*model.WalletPaymentMethod](ctx, r.Proxy, "payment-wallet-service", rest.Request{Method: "GET", Path: "/api/v1/payment-methods", Query: restQuery("user_id", userIDOf(ctx))}, "")
}
