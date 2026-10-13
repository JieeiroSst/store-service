package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) SolEstimateTransferFee(ctx context.Context, from string, to string, amount int) (any, error) {
	return restCall[any](ctx, r.Proxy, "solana-service", rest.Request{Method: "POST", Path: "/api/v1/solana/transactions/estimate-fee", Body: map[string]any{"from": from, "to": to, "amount": amount}}, "")
}

func (r *mutationResolver) SolFindPda(ctx context.Context, programID string, userPubkey string, identifier string) (any, error) {
	return restCall[any](ctx, r.Proxy, "solana-service", rest.Request{Method: "POST", Path: "/api/v1/solana/programs/pda", Body: map[string]any{"program_id": programID, "user_pubkey": userPubkey, "identifier": identifier}}, "")
}

func (r *mutationResolver) CircleCreateWalletSet(ctx context.Context, name string, custodyType *string) (*model.SolCircleWalletSet, error) {
	return restCall[*model.SolCircleWalletSet](ctx, r.Proxy, "solana-service", rest.Request{Method: "POST", Path: "/api/v1/circle/wallet-sets", Body: map[string]any{"name": name, "custodyType": custodyType}}, "")
}

func (r *mutationResolver) CircleCreateWallet(ctx context.Context, walletSetID string, blockchain string, accountType *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "solana-service", rest.Request{Method: "POST", Path: "/api/v1/circle/wallets", Body: map[string]any{"walletSetId": walletSetID, "blockchain": blockchain, "accountType": accountType}}, "")
}

func (r *mutationResolver) CircleFreezeWallet(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "solana-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/circle/wallets/%v/freeze", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) CircleUnfreezeWallet(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "solana-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/circle/wallets/%v/unfreeze", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) CircleCreateTransfer(ctx context.Context, walletID string, destinationAddress string, tokenID string, amount string, fee *model.SolFeeInput, idempotencyKey *string) (*model.SolCircleTransaction, error) {
	return restCall[*model.SolCircleTransaction](ctx, r.Proxy, "solana-service", rest.Request{Method: "POST", Path: "/api/v1/circle/transactions/transfer", Body: map[string]any{"wallet_id": walletID, "destination_address": destinationAddress, "token_id": tokenID, "amount": amount, "fee": fee, "idempotency_key": idempotencyKey}}, "")
}

func (r *mutationResolver) CircleCreateNftTransfer(ctx context.Context, input map[string]any) (*model.SolCircleTransaction, error) {
	return restCall[*model.SolCircleTransaction](ctx, r.Proxy, "solana-service", rest.Request{Method: "POST", Path: "/api/v1/circle/transactions/nft-transfer", Body: input}, "")
}

func (r *mutationResolver) CircleExecuteContract(ctx context.Context, input map[string]any) (*model.SolCircleTransaction, error) {
	return restCall[*model.SolCircleTransaction](ctx, r.Proxy, "solana-service", rest.Request{Method: "POST", Path: "/api/v1/circle/transactions/contract-execution", Body: input}, "")
}

func (r *mutationResolver) CircleImportToken(ctx context.Context, input map[string]any) (*model.SolToken, error) {
	return restCall[*model.SolToken](ctx, r.Proxy, "solana-service", rest.Request{Method: "POST", Path: "/api/v1/circle/tokens/import", Body: input}, "")
}

func (r *mutationResolver) BridgeSolanaToCircle(ctx context.Context, fromSolanaAddress string, toCircleWalletID string, amount int, privateKey string) (*model.SolBridgeTransfer, error) {
	return restCall[*model.SolBridgeTransfer](ctx, r.Proxy, "solana-service", rest.Request{Method: "POST", Path: "/api/v1/bridge/solana-to-circle", Body: map[string]any{"fromSolanaAddress": fromSolanaAddress, "toCircleWalletId": toCircleWalletID, "amount": amount, "privateKey": privateKey}}, "")
}

func (r *mutationResolver) BridgeCircleToSolana(ctx context.Context, fromCircleWalletID string, toSolanaAddress string, amount string) (*model.SolBridgeTransfer, error) {
	return restCall[*model.SolBridgeTransfer](ctx, r.Proxy, "solana-service", rest.Request{Method: "POST", Path: "/api/v1/bridge/circle-to-solana", Body: map[string]any{"fromCircleWalletId": fromCircleWalletID, "toSolanaAddress": toSolanaAddress, "amount": amount}}, "")
}

func (r *queryResolver) SolAccount(ctx context.Context, address string) (*model.SolAccountInfo, error) {
	return restCall[*model.SolAccountInfo](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/solana/accounts/%v", address)}, "")
}

func (r *queryResolver) SolBalance(ctx context.Context, address string) (*model.SolBalance, error) {
	return restCall[*model.SolBalance](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/solana/accounts/%v/balance", address)}, "")
}

func (r *queryResolver) SolTransaction(ctx context.Context, signature string) (*model.SolTransaction, error) {
	return restCall[*model.SolTransaction](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/solana/transactions/%v", signature)}, "")
}

func (r *queryResolver) SolProgram(ctx context.Context, id string) (*model.SolProgram, error) {
	return restCall[*model.SolProgram](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/solana/programs/%v", id)}, "")
}

func (r *queryResolver) SolCircleWalletSet(ctx context.Context, id string) (*model.SolCircleWalletSet, error) {
	return restCall[*model.SolCircleWalletSet](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/circle/wallet-sets/%v", id)}, "")
}

func (r *queryResolver) SolCircleWalletSets(ctx context.Context) ([]*model.SolCircleWalletSet, error) {
	return restCall[[]*model.SolCircleWalletSet](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: "/api/v1/circle/wallet-sets"}, "walletSets")
}

func (r *queryResolver) SolCircleWallet(ctx context.Context, id string) (*model.SolCircleWallet, error) {
	return restCall[*model.SolCircleWallet](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/circle/wallets/%v", id)}, "")
}

func (r *queryResolver) SolCircleWallets(ctx context.Context, walletSetID *string, blockchain *string) ([]*model.SolCircleWallet, error) {
	return restCall[[]*model.SolCircleWallet](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: "/api/v1/circle/wallets", Query: restQuery("walletSetId", walletSetID, "blockchain", blockchain)}, "wallets")
}

func (r *queryResolver) SolCircleWalletBalance(ctx context.Context, id string) ([]*model.SolCircleBalance, error) {
	return restCall[[]*model.SolCircleBalance](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/circle/wallets/%v/balance", id)}, "balances")
}

func (r *queryResolver) SolCircleWalletNFTs(ctx context.Context, id string) ([]*model.SolNFTBalance, error) {
	return restCall[[]*model.SolNFTBalance](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/circle/wallets/%v/nfts", id)}, "nfts")
}

func (r *queryResolver) SolCircleTransaction(ctx context.Context, id string) (*model.SolCircleTransaction, error) {
	return restCall[*model.SolCircleTransaction](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/circle/transactions/%v", id)}, "")
}

func (r *queryResolver) SolCircleEstimateFee(ctx context.Context, walletID string, destinationAddress string, tokenID string, amount string) (*model.SolFeeEstimate, error) {
	return restCall[*model.SolFeeEstimate](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: "/api/v1/circle/transactions/estimate-fee", Query: restQuery("walletId", walletID, "destinationAddress", destinationAddress, "tokenId", tokenID, "amount", amount)}, "")
}

func (r *queryResolver) SolCircleToken(ctx context.Context, id string) (*model.SolToken, error) {
	return restCall[*model.SolToken](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/circle/tokens/%v", id)}, "")
}

func (r *queryResolver) SolCircleTokens(ctx context.Context, blockchain *string) ([]*model.SolToken, error) {
	return restCall[[]*model.SolToken](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: "/api/v1/circle/tokens", Query: restQuery("blockchain", blockchain)}, "tokens")
}

func (r *queryResolver) SolBridgeTransfer(ctx context.Context, id string) (*model.SolBridgeTransfer, error) {
	return restCall[*model.SolBridgeTransfer](ctx, r.Proxy, "solana-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/bridge/transfers/%v", id)}, "")
}
