package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) WalletQuery() generated.WalletQueryResolver { return &walletQueryResolver{r} }

type walletQueryResolver struct{ *Resolver }

func (r *walletQueryResolver) Wallet(ctx context.Context, obj *model.WalletQuery, id string) (*model.WalletWallet, error) {
	return r.Clients.PaymentWalletService.Wallet(ctx, id)
}

func (r *walletQueryResolver) WalletByUser(ctx context.Context, obj *model.WalletQuery, userID string) (*model.WalletWallet, error) {
	return r.Clients.PaymentWalletService.WalletByUser(ctx, userID)
}

func (r *walletQueryResolver) Transactions(ctx context.Context, obj *model.WalletQuery, id string, limit *int, offset *int) ([]*model.WalletTransaction, error) {
	return r.Clients.PaymentWalletService.Transactions(ctx, id, limit, offset)
}

func (r *walletQueryResolver) TransactionByReference(ctx context.Context, obj *model.WalletQuery, id string, referenceID string) (*model.WalletTransaction, error) {
	return r.Clients.PaymentWalletService.TransactionByReference(ctx, id, referenceID)
}

func (r *walletQueryResolver) Pockets(ctx context.Context, obj *model.WalletQuery, id string) ([]*model.WalletPocket, error) {
	return r.Clients.PaymentWalletService.Pockets(ctx, id)
}

func (r *walletQueryResolver) Transfer(ctx context.Context, obj *model.WalletQuery, id string) (*model.WalletTransfer, error) {
	return r.Clients.PaymentWalletService.Transfer(ctx, id)
}

func (r *walletQueryResolver) Transaction(ctx context.Context, obj *model.WalletQuery, id string) (*model.WalletTransaction, error) {
	return r.Clients.PaymentWalletService.Transaction(ctx, id)
}

func (r *walletQueryResolver) PaymentRequests(ctx context.Context, obj *model.WalletQuery, walletID string) ([]*model.WalletPaymentRequest, error) {
	return r.Clients.PaymentWalletService.PaymentRequests(ctx, walletID)
}

func (r *walletQueryResolver) PaymentRequest(ctx context.Context, obj *model.WalletQuery, id string) (*model.WalletPaymentRequest, error) {
	return r.Clients.PaymentWalletService.PaymentRequest(ctx, id)
}

func (r *walletQueryResolver) PaymentRequestQR(ctx context.Context, obj *model.WalletQuery, id string) (*model.WalletPaymentRequestQR, error) {
	return r.Clients.PaymentWalletService.PaymentRequestQR(ctx, id)
}

func (r *walletQueryResolver) PaymentMethods(ctx context.Context, obj *model.WalletQuery, userID string) ([]*model.WalletPaymentMethod, error) {
	return r.Clients.PaymentWalletService.PaymentMethods(ctx, userID)
}
