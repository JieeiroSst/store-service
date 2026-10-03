package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) SolQuery() generated.SolQueryResolver { return &solQueryResolver{r} }

type solQueryResolver struct{ *Resolver }

func (r *solQueryResolver) Account(ctx context.Context, obj *model.SolQuery, address string) (*model.SolAccountInfo, error) {
	return r.Clients.SolanaService.Account(ctx, address)
}

func (r *solQueryResolver) Balance(ctx context.Context, obj *model.SolQuery, address string) (*model.SolBalance, error) {
	return r.Clients.SolanaService.Balance(ctx, address)
}

func (r *solQueryResolver) Transaction(ctx context.Context, obj *model.SolQuery, signature string) (*model.SolTransaction, error) {
	return r.Clients.SolanaService.Transaction(ctx, signature)
}

func (r *solQueryResolver) Program(ctx context.Context, obj *model.SolQuery, id string) (*model.SolProgram, error) {
	return r.Clients.SolanaService.Program(ctx, id)
}

func (r *solQueryResolver) CircleWalletSet(ctx context.Context, obj *model.SolQuery, id string) (*model.SolCircleWalletSet, error) {
	return r.Clients.SolanaService.CircleWalletSet(ctx, id)
}

func (r *solQueryResolver) CircleWalletSets(ctx context.Context, obj *model.SolQuery) ([]*model.SolCircleWalletSet, error) {
	return r.Clients.SolanaService.CircleWalletSets(ctx)
}

func (r *solQueryResolver) CircleWallet(ctx context.Context, obj *model.SolQuery, id string) (*model.SolCircleWallet, error) {
	return r.Clients.SolanaService.CircleWallet(ctx, id)
}

func (r *solQueryResolver) CircleWallets(ctx context.Context, obj *model.SolQuery, walletSetID *string, blockchain *string) ([]*model.SolCircleWallet, error) {
	return r.Clients.SolanaService.CircleWallets(ctx, walletSetID, blockchain)
}

func (r *solQueryResolver) CircleWalletBalance(ctx context.Context, obj *model.SolQuery, id string) ([]*model.SolCircleBalance, error) {
	return r.Clients.SolanaService.CircleWalletBalance(ctx, id)
}

func (r *solQueryResolver) CircleWalletNFTs(ctx context.Context, obj *model.SolQuery, id string) ([]*model.SolNFTBalance, error) {
	return r.Clients.SolanaService.CircleWalletNFTs(ctx, id)
}

func (r *solQueryResolver) CircleTransaction(ctx context.Context, obj *model.SolQuery, id string) (*model.SolCircleTransaction, error) {
	return r.Clients.SolanaService.CircleTransaction(ctx, id)
}

func (r *solQueryResolver) CircleEstimateFee(ctx context.Context, obj *model.SolQuery, walletID string, destinationAddress string, tokenID string, amount string) (*model.SolFeeEstimate, error) {
	return r.Clients.SolanaService.CircleEstimateFee(ctx, walletID, destinationAddress, tokenID, amount)
}

func (r *solQueryResolver) CircleToken(ctx context.Context, obj *model.SolQuery, id string) (*model.SolToken, error) {
	return r.Clients.SolanaService.CircleToken(ctx, id)
}

func (r *solQueryResolver) CircleTokens(ctx context.Context, obj *model.SolQuery, blockchain *string) ([]*model.SolToken, error) {
	return r.Clients.SolanaService.CircleTokens(ctx, blockchain)
}

func (r *solQueryResolver) BridgeTransfer(ctx context.Context, obj *model.SolQuery, id string) (*model.SolBridgeTransfer, error) {
	return r.Clients.SolanaService.BridgeTransfer(ctx, id)
}
