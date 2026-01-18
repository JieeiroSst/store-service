package services

import (
	"context"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
	"github.com/JIeeiroSst/solana-service/internal/core/ports"
)

type CircleWalletService struct {
	wallet ports.CircleWalletPort
}

func NewCircleWalletService(wallet ports.CircleWalletPort) *CircleWalletService {
	return &CircleWalletService{
		wallet: wallet,
	}
}

func (s *CircleWalletService) CreateWalletSet(ctx context.Context, name string, custodyType domain.CustodyType) (*domain.CircleWalletSet, error) {
	return s.wallet.CreateWalletSet(ctx, name, custodyType)
}

func (s *CircleWalletService) GetWalletSet(ctx context.Context, walletSetID string) (*domain.CircleWalletSet, error) {
	return s.wallet.GetWalletSet(ctx, walletSetID)
}

func (s *CircleWalletService) ListWalletSets(ctx context.Context) ([]*domain.CircleWalletSet, error) {
	return s.wallet.ListWalletSets(ctx)
}

func (s *CircleWalletService) CreateSolanaWallet(ctx context.Context, walletSetID string) (*domain.CircleWallet, error) {
	req := domain.CreateWalletRequest{
		WalletSetID: walletSetID,
		Blockchains: []string{"SOL"},
		Count:       1,
		AccountType: domain.AccountTypeEOA,
	}

	wallets, err := s.wallet.CreateWallets(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(wallets) == 0 {
		return nil, domain.ErrWalletNotFound
	}

	return wallets[0], nil
}

func (s *CircleWalletService) CreateMultiChainWallet(ctx context.Context, walletSetID string, blockchains []string) ([]*domain.CircleWallet, error) {
	req := domain.CreateWalletRequest{
		WalletSetID: walletSetID,
		Blockchains: blockchains,
		Count:       1,
		AccountType: domain.AccountTypeEOA,
	}

	return s.wallet.CreateWallets(ctx, req)
}

func (s *CircleWalletService) GetWallet(ctx context.Context, walletID string) (*domain.CircleWallet, error) {
	return s.wallet.GetWallet(ctx, walletID)
}

func (s *CircleWalletService) ListWallets(ctx context.Context, walletSetID, blockchain string) ([]*domain.CircleWallet, error) {
	return s.wallet.ListWallets(ctx, walletSetID, blockchain)
}

func (s *CircleWalletService) UpdateWallet(ctx context.Context, walletID string, req domain.UpdateWalletRequest) (*domain.CircleWallet, error) {
	return s.wallet.UpdateWallet(ctx, walletID, req)
}

func (s *CircleWalletService) FreezeWallet(ctx context.Context, walletID string) (*domain.CircleWallet, error) {
	req := domain.UpdateWalletRequest{
		State: domain.WalletStateFrozen,
	}
	return s.wallet.UpdateWallet(ctx, walletID, req)
}

func (s *CircleWalletService) UnfreezeWallet(ctx context.Context, walletID string) (*domain.CircleWallet, error) {
	req := domain.UpdateWalletRequest{
		State: domain.WalletStateLive,
	}
	return s.wallet.UpdateWallet(ctx, walletID, req)
}

func (s *CircleWalletService) GetWalletBalance(ctx context.Context, walletID string) ([]*domain.CircleBalance, error) {
	return s.wallet.GetWalletBalance(ctx, walletID)
}

func (s *CircleWalletService) GetWalletNFTs(ctx context.Context, walletID string) ([]*domain.NFTBalance, error) {
	return s.wallet.GetWalletNFTs(ctx, walletID)
}
