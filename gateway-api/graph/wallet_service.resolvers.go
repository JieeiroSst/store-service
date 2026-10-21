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

func (r *mutationResolver) WalletSvcMyCreateWallet(ctx context.Context, currency string) (*model.WalletSvcWallet, error) {
	return restCall[*model.WalletSvcWallet](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: "/api/v1/wallets", Body: withField(map[string]any{"currency": currency}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) WalletSvcMySetPin(ctx context.Context, walletID string, currentPin *string, newPin string) (any, error) {
	return restCall[any](ctx, r.Proxy, "wallet-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/wallets/%v/pin", walletID), Body: withField(map[string]any{"current_pin": currentPin, "new_pin": newPin}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) WalletSvcMyCloseWallet(ctx context.Context, walletID string) (*model.WalletSvcWallet, error) {
	return restCall[*model.WalletSvcWallet](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/close", walletID), Body: withField(map[string]any{}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) WalletSvcMyTopUp(ctx context.Context, walletID string, bankAccountID string, amount string, description *string, idempotencyKey string) (*model.WalletSvcTransaction, error) {
	return restCall[*model.WalletSvcTransaction](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/topup", walletID), Body: withField(map[string]any{"bank_account_id": bankAccountID, "amount": amount, "description": description, "idempotency_key": idempotencyKey}, "user_id", userIDOf(ctx)), Header: restHeader("Idempotency-Key", idempotencyKey)}, "")
}

func (r *mutationResolver) WalletSvcMyWithdraw(ctx context.Context, walletID string, bankAccountID string, amount string, pin string, description *string, idempotencyKey string) (*model.WalletSvcTransaction, error) {
	return restCall[*model.WalletSvcTransaction](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/withdraw", walletID), Body: withField(map[string]any{"bank_account_id": bankAccountID, "amount": amount, "pin": pin, "description": description, "idempotency_key": idempotencyKey}, "user_id", userIDOf(ctx)), Header: restHeader("Idempotency-Key", idempotencyKey)}, "")
}

func (r *mutationResolver) WalletSvcMyTransfer(ctx context.Context, fromWalletID string, toWalletID *string, toUserID *string, amount string, pin string, description *string, idempotencyKey string) (*model.WalletSvcTransaction, error) {
	return restCall[*model.WalletSvcTransaction](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: "/api/v1/transfers", Body: withField(map[string]any{"from_wallet_id": fromWalletID, "to_wallet_id": toWalletID, "to_user_id": toUserID, "amount": amount, "pin": pin, "description": description, "idempotency_key": idempotencyKey}, "user_id", userIDOf(ctx)), Header: restHeader("Idempotency-Key", idempotencyKey)}, "")
}

func (r *mutationResolver) WalletSvcMyPay(ctx context.Context, walletID string, merchantID string, amount string, pin string, description *string, metadata map[string]any, idempotencyKey string) (*model.WalletSvcTransaction, error) {
	return restCall[*model.WalletSvcTransaction](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: "/api/v1/payments", Body: withField(map[string]any{"wallet_id": walletID, "merchant_id": merchantID, "amount": amount, "pin": pin, "description": description, "metadata": metadata, "idempotency_key": idempotencyKey}, "user_id", userIDOf(ctx)), Header: restHeader("Idempotency-Key", idempotencyKey)}, "")
}

func (r *mutationResolver) WalletSvcMyLinkBankAccount(ctx context.Context, bankID string, accountNumber string, accountName string) (*model.WalletSvcBankAccount, error) {
	return restCall[*model.WalletSvcBankAccount](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/users/%v/bank-accounts", userIDOf(ctx)), Body: map[string]any{"bank_id": bankID, "account_number": accountNumber, "account_name": accountName}}, "")
}

func (r *mutationResolver) WalletSvcMyRemoveBankAccount(ctx context.Context, accountID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "wallet-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/users/%v/bank-accounts/%v", userIDOf(ctx), accountID)}, "")
}

func (r *mutationResolver) WalletSvcMySetDefaultBankAccount(ctx context.Context, accountID string) (*model.WalletSvcBankAccount, error) {
	return restCall[*model.WalletSvcBankAccount](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/users/%v/bank-accounts/%v/default", userIDOf(ctx), accountID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) WalletSvcMyIssueCard(ctx context.Context, walletID string, issuerBankID string, network string, holderName string) (*model.WalletSvcCard, error) {
	return restOwned[*model.WalletSvcCard](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/cards", walletID), Body: map[string]any{"issuer_bank_id": issuerBankID, "network": network, "holder_name": holderName}}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", walletID)}, "")
}

func (r *mutationResolver) WalletSvcMyRequestPayment(ctx context.Context, walletID string, payerUserID string, amount string, note *string) (*model.WalletSvcPaymentRequest, error) {
	return restCall[*model.WalletSvcPaymentRequest](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: "/api/v1/payment-requests", Body: withField(map[string]any{"wallet_id": walletID, "payer_user_id": payerUserID, "amount": amount, "note": note}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) WalletSvcMyPayPaymentRequest(ctx context.Context, id string, walletID string, pin string) (*model.WalletSvcPaymentRequest, error) {
	return restCall[*model.WalletSvcPaymentRequest](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/payment-requests/%v/pay", id), Body: withField(map[string]any{"wallet_id": walletID, "pin": pin}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) WalletSvcMyDeclinePaymentRequest(ctx context.Context, id string) (*model.WalletSvcPaymentRequest, error) {
	return restCall[*model.WalletSvcPaymentRequest](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/payment-requests/%v/decline", id), Body: withField(map[string]any{}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) WalletSvcMyCancelPaymentRequest(ctx context.Context, id string) (*model.WalletSvcPaymentRequest, error) {
	return restCall[*model.WalletSvcPaymentRequest](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/payment-requests/%v/cancel", id), Body: withField(map[string]any{}, "user_id", userIDOf(ctx))}, "")
}

func (r *mutationResolver) WalletSvcUnfreeze(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wallets/%v/unfreeze", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) WalletSvcChangeTier(ctx context.Context, id string, tier string) (*model.WalletSvcWallet, error) {
	return restCall[*model.WalletSvcWallet](ctx, r.Proxy, "wallet-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/wallets/%v/tier", id), Body: map[string]any{"tier": tier}}, "")
}

func (r *mutationResolver) WalletSvcRefund(ctx context.Context, id string, amount *string, reason *string, idempotencyKey string) (*model.WalletSvcTransaction, error) {
	return restCall[*model.WalletSvcTransaction](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/transactions/%v/refund", id), Body: map[string]any{"amount": amount, "reason": reason, "idempotency_key": idempotencyKey}, Header: restHeader("Idempotency-Key", idempotencyKey)}, "")
}

func (r *mutationResolver) WalletSvcBlockCard(ctx context.Context, id string) (*model.WalletSvcCard, error) {
	return restCall[*model.WalletSvcCard](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/cards/%v/block", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) WalletSvcUnblockCard(ctx context.Context, id string) (*model.WalletSvcCard, error) {
	return restCall[*model.WalletSvcCard](ctx, r.Proxy, "wallet-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/cards/%v/unblock", id), Body: map[string]any{}}, "")
}

func (r *queryResolver) WalletSvcWallet(ctx context.Context, id string) (*model.WalletSvcWallet, error) {
	return restCall[*model.WalletSvcWallet](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", id)}, "")
}

func (r *queryResolver) WalletSvcTransaction(ctx context.Context, id string) (*model.WalletSvcTransaction, error) {
	return restCall[*model.WalletSvcTransaction](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/transactions/%v", id)}, "")
}

func (r *queryResolver) WalletSvcMyWallets(ctx context.Context) ([]*model.WalletSvcWallet, error) {
	return restCall[[]*model.WalletSvcWallet](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/wallets", userIDOf(ctx))}, "items")
}

func (r *queryResolver) WalletSvcMyWallet(ctx context.Context, currency string) (*model.WalletSvcWallet, error) {
	return restCall[*model.WalletSvcWallet](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/wallets/%v", userIDOf(ctx), currency)}, "")
}

func (r *queryResolver) WalletSvcMyBalance(ctx context.Context, walletID string) (*model.WalletSvcBalance, error) {
	return restOwned[*model.WalletSvcBalance](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/balance", walletID)}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", walletID)}, "")
}

func (r *queryResolver) WalletSvcMyLimits(ctx context.Context, walletID string) (*model.WalletSvcLimits, error) {
	return restOwned[*model.WalletSvcLimits](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/limits", walletID)}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", walletID)}, "")
}

func (r *queryResolver) WalletSvcMyTransactions(ctx context.Context, walletID string, typeArg *string, status *string, from *string, to *string, page *int, pageSize *int) (*model.WalletSvcTransactionPage, error) {
	return restOwned[*model.WalletSvcTransactionPage](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/transactions", walletID), Query: restQuery("type", typeArg, "status", status, "from", from, "to", to, "page", page, "page_size", pageSize)}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", walletID)}, "")
}

func (r *queryResolver) WalletSvcMyLedger(ctx context.Context, walletID string, from *string, to *string, page *int, pageSize *int) (*model.WalletSvcLedgerPage, error) {
	return restOwned[*model.WalletSvcLedgerPage](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/ledger", walletID), Query: restQuery("from", from, "to", to, "page", page, "page_size", pageSize)}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", walletID)}, "")
}

func (r *queryResolver) WalletSvcMyStatement(ctx context.Context, walletID string, from string, to string) (*model.WalletSvcStatement, error) {
	return restOwned[*model.WalletSvcStatement](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/statement", walletID), Query: restQuery("from", from, "to", to)}, "", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", walletID)}, "")
}

func (r *queryResolver) WalletSvcMyCards(ctx context.Context, walletID string) ([]*model.WalletSvcCard, error) {
	return restOwned[[]*model.WalletSvcCard](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/cards", walletID)}, "items", "user_id", &rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v", walletID)}, "")
}

func (r *queryResolver) WalletSvcMyBankAccounts(ctx context.Context) ([]*model.WalletSvcBankAccount, error) {
	return restCall[[]*model.WalletSvcBankAccount](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/bank-accounts", userIDOf(ctx))}, "items")
}

func (r *queryResolver) WalletSvcMyPaymentRequests(ctx context.Context, role *string, status *string, page *int, pageSize *int) (*model.WalletSvcPaymentRequestPage, error) {
	return restCall[*model.WalletSvcPaymentRequestPage](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/payment-requests", userIDOf(ctx)), Query: restQuery("role", role, "status", status, "page", page, "page_size", pageSize)}, "")
}

func (r *queryResolver) WalletSvcUserWallets(ctx context.Context, userID string) ([]*model.WalletSvcWallet, error) {
	return restCall[[]*model.WalletSvcWallet](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/wallets", userID)}, "items")
}

func (r *queryResolver) WalletSvcBalance(ctx context.Context, id string) (*model.WalletSvcBalance, error) {
	return restCall[*model.WalletSvcBalance](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/balance", id)}, "")
}

func (r *queryResolver) WalletSvcTransactions(ctx context.Context, id string, typeArg *string, status *string, from *string, to *string, page *int, pageSize *int) (*model.WalletSvcTransactionPage, error) {
	return restCall[*model.WalletSvcTransactionPage](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/transactions", id), Query: restQuery("type", typeArg, "status", status, "from", from, "to", to, "page", page, "page_size", pageSize)}, "")
}

func (r *queryResolver) WalletSvcStatement(ctx context.Context, id string, from string, to string) (*model.WalletSvcStatement, error) {
	return restCall[*model.WalletSvcStatement](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wallets/%v/statement", id), Query: restQuery("from", from, "to", to)}, "")
}

func (r *queryResolver) WalletSvcCard(ctx context.Context, id string) (*model.WalletSvcCard, error) {
	return restCall[*model.WalletSvcCard](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/cards/%v", id)}, "")
}

func (r *queryResolver) WalletSvcPaymentRequest(ctx context.Context, id string) (*model.WalletSvcPaymentRequest, error) {
	return restCall[*model.WalletSvcPaymentRequest](ctx, r.Proxy, "wallet-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/payment-requests/%v", id)}, "")
}
