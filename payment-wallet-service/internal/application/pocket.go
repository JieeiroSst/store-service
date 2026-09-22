package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"github.com/google/uuid"
)

type pocketService struct {
	wallets port.WalletRepository
	pockets port.PocketRepository
}

func NewPocketService(wallets port.WalletRepository, pockets port.PocketRepository) port.PocketUsecase {
	return &pocketService{wallets: wallets, pockets: pockets}
}

func (s *pocketService) CreatePocket(ctx context.Context, walletID, name string) (*model.Pocket, error) {
	wallet, err := s.wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}

	pocket := &model.Pocket{
		PocketID: uuid.NewString(),
		WalletID: walletID,
		Name:     name,
		Balance:  0,
		Currency: wallet.Currency,
	}
	if err := s.pockets.Create(ctx, pocket); err != nil {
		return nil, fmt.Errorf("create pocket: %w", err)
	}
	return pocket, nil
}

func (s *pocketService) ListPockets(ctx context.Context, walletID string) ([]model.Pocket, error) {
	return s.pockets.ListByWallet(ctx, walletID)
}

func (s *pocketService) DepositToPocket(ctx context.Context, pocketID string, amount int64) (*model.Pocket, error) {
	if amount <= 0 {
		return nil, port.ErrInvalidAmount
	}
	pocket, err := s.pockets.GetByID(ctx, pocketID)
	if err != nil {
		return nil, fmt.Errorf("get pocket: %w", err)
	}

	txn := &model.Transaction{
		TransactionID: uuid.NewString(),
		WalletID:      pocket.WalletID,
		Type:          model.TxnPocketOut,
		Amount:        amount,
		Currency:      pocket.Currency,
		Status:        model.TxnCompleted,
		PocketID:      &pocket.PocketID,
	}
	_, updated, err := s.pockets.MoveToPocket(ctx, pocket.WalletID, pocketID, txn)
	if err != nil {
		return nil, fmt.Errorf("move to pocket: %w", err)
	}
	return updated, nil
}

func (s *pocketService) WithdrawFromPocket(ctx context.Context, pocketID string, amount int64) (*model.Pocket, error) {
	if amount <= 0 {
		return nil, port.ErrInvalidAmount
	}
	pocket, err := s.pockets.GetByID(ctx, pocketID)
	if err != nil {
		return nil, fmt.Errorf("get pocket: %w", err)
	}

	txn := &model.Transaction{
		TransactionID: uuid.NewString(),
		WalletID:      pocket.WalletID,
		Type:          model.TxnPocketIn,
		Amount:        amount,
		Currency:      pocket.Currency,
		Status:        model.TxnCompleted,
		PocketID:      &pocket.PocketID,
	}
	_, updated, err := s.pockets.MoveFromPocket(ctx, pocket.WalletID, pocketID, txn)
	if err != nil {
		return nil, fmt.Errorf("move from pocket: %w", err)
	}
	return updated, nil
}

func (s *pocketService) ClosePocket(ctx context.Context, pocketID string) (*model.Wallet, error) {
	pocket, err := s.pockets.GetByID(ctx, pocketID)
	if err != nil {
		return nil, fmt.Errorf("get pocket: %w", err)
	}
	wallet, err := s.pockets.Close(ctx, pocket.WalletID, pocketID)
	if err != nil {
		return nil, fmt.Errorf("close pocket: %w", err)
	}
	return wallet, nil
}
