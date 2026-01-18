package ports

import (
	"context"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
)

type CircleWalletPort interface {
	// Wallet Set operations
	CreateWalletSet(ctx context.Context, name string, custodyType domain.CustodyType) (*domain.CircleWalletSet, error)
	GetWalletSet(ctx context.Context, walletSetID string) (*domain.CircleWalletSet, error)
	ListWalletSets(ctx context.Context) ([]*domain.CircleWalletSet, error)
	UpdateWalletSet(ctx context.Context, walletSetID, name string) (*domain.CircleWalletSet, error)

	// Wallet operations
	CreateWallets(ctx context.Context, req domain.CreateWalletRequest) ([]*domain.CircleWallet, error)
	GetWallet(ctx context.Context, walletID string) (*domain.CircleWallet, error)
	ListWallets(ctx context.Context, walletSetID string, blockchain string) ([]*domain.CircleWallet, error)
	UpdateWallet(ctx context.Context, walletID string, req domain.UpdateWalletRequest) (*domain.CircleWallet, error)

	// Balance operations
	GetWalletBalance(ctx context.Context, walletID string) ([]*domain.CircleBalance, error)
	GetWalletNFTs(ctx context.Context, walletID string) ([]*domain.NFTBalance, error)
}

type CircleTransactionPort interface {
	// Transfer operations
	CreateTransfer(ctx context.Context, req domain.TransferRequest) (*domain.CircleTransaction, error)
	CreateNFTTransfer(ctx context.Context, req domain.NFTTransferRequest) (*domain.CircleTransaction, error)

	// Contract operations
	ExecuteContract(ctx context.Context, req domain.ContractExecutionRequest) (*domain.CircleTransaction, error)

	// Transaction management
	GetTransaction(ctx context.Context, txID string) (*domain.CircleTransaction, error)
	ListTransactions(ctx context.Context, filter domain.TransactionFilter) ([]*domain.CircleTransaction, error)
	ValidateAddress(ctx context.Context, blockchain, address string) (bool, error)

	// Fee operations
	EstimateFee(ctx context.Context, walletID, destinationAddress, tokenID, amount string) (*domain.FeeEstimate, error)

	// Transaction control
	AccelerateTransaction(ctx context.Context, txID string, req domain.AccelerateTransactionRequest) (*domain.CircleTransaction, error)
	CancelTransaction(ctx context.Context, txID string, req domain.CancelTransactionRequest) (*domain.CircleTransaction, error)
}

type CircleTokenPort interface {
	GetToken(ctx context.Context, tokenID string) (*domain.Token, error)
	ListTokens(ctx context.Context, blockchain string) ([]*domain.Token, error)
	ImportToken(ctx context.Context, metadata domain.TokenMetadata) (*domain.Token, error)

	GetNFTToken(ctx context.Context, tokenID string) (*domain.NFTToken, error)
	ListNFTTokens(ctx context.Context, blockchain string) ([]*domain.NFTToken, error)
}
