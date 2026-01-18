package services

import (
	"context"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
	"github.com/JIeeiroSst/solana-service/internal/core/ports"
	"github.com/gagliardetto/solana-go"
)

type AccountService struct {
	blockchain ports.BlockchainPort
	repo       ports.AccountRepository
}

func NewAccountService(blockchain ports.BlockchainPort, repo ports.AccountRepository) *AccountService {
	return &AccountService{
		blockchain: blockchain,
		repo:       repo,
	}
}

func (s *AccountService) GetAccountInfo(ctx context.Context, address string) (*domain.AccountInfo, error) {
	pubkey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return nil, domain.ErrInvalidPublicKey
	}

	account, err := s.blockchain.GetAccount(ctx, pubkey)
	if err != nil {
		return nil, err
	}

	return &domain.AccountInfo{
		Address:    address,
		Balance:    account.Lamports,
		Owner:      account.Owner.String(),
		Executable: account.Executable,
		RentEpoch:  account.RentEpoch,
	}, nil
}

func (s *AccountService) GetBalance(ctx context.Context, address string) (uint64, error) {
	pubkey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return 0, domain.ErrInvalidPublicKey
	}

	return s.blockchain.GetBalance(ctx, pubkey)
}

func (s *AccountService) CreateAccount(ctx context.Context) (*solana.PrivateKey, error) {
	privateKey := solana.NewWallet()
	return &privateKey.PrivateKey, nil
}

