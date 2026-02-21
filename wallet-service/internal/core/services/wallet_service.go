package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
	"github.com/JIeeiroSst/wallet-service/internal/core/ports"
)

const (
	walletCacheTTL = 10 * time.Minute
	walletByIDKey  = "wallet:id:%s"
	walletByUserKey = "wallet:user:%s"
)

type walletService struct {
	walletRepo ports.WalletRepository
	txRepo     ports.TransactionRepository
	cache      ports.CacheRepository
	logger     *zap.Logger
}

func NewWalletService(
	walletRepo ports.WalletRepository,
	txRepo ports.TransactionRepository,
	cache ports.CacheRepository,
	logger *zap.Logger,
) ports.WalletUseCase {
	return &walletService{walletRepo: walletRepo, txRepo: txRepo, cache: cache, logger: logger}
}

func (s *walletService) CreateWallet(ctx context.Context, userID uuid.UUID, currency domain.Currency) (*domain.Wallet, error) {
	now := time.Now()
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Balance:      decimal.Zero,
		FrozenAmount: decimal.Zero,
		Currency:     currency,
		Status:       domain.WalletStatusActive,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		s.logger.Error("create wallet failed", zap.Error(err))
		return nil, err
	}
	s.setWalletCache(ctx, wallet)
	s.logger.Info("wallet created", zap.String("wallet_id", wallet.ID.String()))
	return wallet, nil
}

func (s *walletService) GetWallet(ctx context.Context, walletID uuid.UUID) (*domain.Wallet, error) {
	key := fmt.Sprintf(walletByIDKey, walletID)
	if w, ok := s.getWalletFromCache(ctx, key); ok {
		return w, nil
	}
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	s.setWalletCache(ctx, wallet)
	return wallet, nil
}

func (s *walletService) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	key := fmt.Sprintf(walletByUserKey, userID)
	if w, ok := s.getWalletFromCache(ctx, key); ok {
		return w, nil
	}
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	s.setWalletCache(ctx, wallet)
	return wallet, nil
}

func (s *walletService) Credit(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal, description string) (*domain.Transaction, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if wallet.Status != domain.WalletStatusActive {
		return nil, domain.ErrWalletInactive
	}
	if err := wallet.Credit(amount); err != nil {
		return nil, err
	}
	if err := s.walletRepo.UpdateWithVersion(ctx, wallet); err != nil {
		return nil, err
	}
	s.invalidateCache(ctx, wallet)

	tx := &domain.Transaction{
		ID: uuid.New(), WalletID: walletID, Amount: amount, Currency: wallet.Currency,
		Type: domain.TransactionTypeSettlement, Status: domain.TransactionStatusSettled,
		Description: description, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *walletService) FreezeWallet(ctx context.Context, walletID uuid.UUID) error {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return err
	}
	wallet.Status = domain.WalletStatusFrozen
	wallet.UpdatedAt = time.Now()
	if err := s.walletRepo.UpdateWithVersion(ctx, wallet); err != nil {
		return err
	}
	s.invalidateCache(ctx, wallet)
	return nil
}

func (s *walletService) UnfreezeWallet(ctx context.Context, walletID uuid.UUID) error {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return err
	}
	wallet.Status = domain.WalletStatusActive
	wallet.UpdatedAt = time.Now()
	if err := s.walletRepo.UpdateWithVersion(ctx, wallet); err != nil {
		return err
	}
	s.invalidateCache(ctx, wallet)
	return nil
}

func (s *walletService) getWalletFromCache(ctx context.Context, key string) (*domain.Wallet, bool) {
	val, err := s.cache.Get(ctx, key)
	if err != nil || val == "" {
		return nil, false
	}
	var w domain.Wallet
	if err := json.Unmarshal([]byte(val), &w); err != nil {
		return nil, false
	}
	return &w, true
}

func (s *walletService) setWalletCache(ctx context.Context, wallet *domain.Wallet) {
	data, err := json.Marshal(wallet)
	if err != nil {
		return
	}
	raw := string(data)
	s.cache.Set(ctx, fmt.Sprintf(walletByIDKey, wallet.ID), raw, walletCacheTTL)
	s.cache.Set(ctx, fmt.Sprintf(walletByUserKey, wallet.UserID), raw, walletCacheTTL)
}

// invalidateCache deletes all cache keys for the wallet (write-invalidate pattern).
// The next read will fetch fresh data from Postgres and re-populate the cache.
func (s *walletService) invalidateCache(ctx context.Context, wallet *domain.Wallet) {
	keys := []string{
		fmt.Sprintf(walletByIDKey, wallet.ID),
		fmt.Sprintf(walletByUserKey, wallet.UserID),
	}
	if err := s.cache.Delete(ctx, keys...); err != nil {
		s.logger.Warn("cache invalidation failed", zap.Strings("keys", keys), zap.Error(err))
	}
}
