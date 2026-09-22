package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"github.com/google/uuid"
)

type walletService struct {
	wallets port.WalletRepository
}

func NewWalletService(wallets port.WalletRepository) port.WalletUsecase {
	return &walletService{wallets: wallets}
}

func (s *walletService) CreateWallet(ctx context.Context, userID, currency string) (*model.Wallet, error) {
	_, err := s.wallets.GetByUserID(ctx, userID)
	if err == nil {
		return nil, port.ErrWalletAlreadyExists
	}
	if !errors.Is(err, port.ErrNotFound) {
		return nil, fmt.Errorf("check existing wallet: %w", err)
	}

	wallet := &model.Wallet{
		WalletID: uuid.NewString(),
		UserID:   userID,
		Balance:  0,
		Currency: currency,
		Status:   model.WalletActive,
	}
	if err := s.wallets.Create(ctx, wallet); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return wallet, nil
}

func (s *walletService) GetWallet(ctx context.Context, walletID string) (*model.Wallet, error) {
	return s.wallets.GetByID(ctx, walletID)
}

func (s *walletService) GetWalletByUser(ctx context.Context, userID string) (*model.Wallet, error) {
	return s.wallets.GetByUserID(ctx, userID)
}

func (s *walletService) FreezeWallet(ctx context.Context, walletID, reason string) (*model.Wallet, error) {
	wallet, err := s.wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if wallet.Status != model.WalletActive {
		return nil, port.ErrInvalidWalletStatusTransition
	}
	return s.wallets.UpdateStatus(ctx, walletID, model.WalletFrozen, &reason)
}

func (s *walletService) UnfreezeWallet(ctx context.Context, walletID string) (*model.Wallet, error) {
	wallet, err := s.wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if wallet.Status != model.WalletFrozen {
		return nil, port.ErrInvalidWalletStatusTransition
	}
	return s.wallets.UpdateStatus(ctx, walletID, model.WalletActive, nil)
}

func (s *walletService) CloseWallet(ctx context.Context, walletID string) (*model.Wallet, error) {
	wallet, err := s.wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if wallet.Status == model.WalletClosed {
		return nil, port.ErrInvalidWalletStatusTransition
	}
	return s.wallets.Close(ctx, walletID)
}

func (s *walletService) SetLimits(ctx context.Context, walletID string, dailyLimit, perTransactionLimit int64) (*model.Wallet, error) {
	if dailyLimit < 0 || perTransactionLimit < 0 {
		return nil, port.ErrInvalidAmount
	}
	return s.wallets.UpdateLimits(ctx, walletID, dailyLimit, perTransactionLimit)
}
