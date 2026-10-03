package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) WalletSvcQuery() generated.WalletSvcQueryResolver {
	return &walletSvcQueryResolver{r}
}

type walletSvcQueryResolver struct{ *Resolver }

func (r *walletSvcQueryResolver) Wallet(ctx context.Context, obj *model.WalletSvcQuery, id string) (*model.WalletSvcWallet, error) {
	return r.Clients.WalletService.Wallet(ctx, id)
}

func (r *walletSvcQueryResolver) Transaction(ctx context.Context, obj *model.WalletSvcQuery, id string) (*model.WalletSvcTransaction, error) {
	return r.Clients.WalletService.Transaction(ctx, id)
}
