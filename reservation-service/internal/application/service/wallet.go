package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type WalletService struct {
	wallets outbound.WalletGateway
}

var _ inbound.WalletUseCase = (*WalletService)(nil)

func NewWalletService(w outbound.WalletGateway) *WalletService { return &WalletService{wallets: w} }

func (s *WalletService) enabled() error {
	if s.wallets == nil {
		return fmt.Errorf("%w: wallets are not enabled", domain.ErrInvalid)
	}
	return nil
}

func (s *WalletService) Get(ctx context.Context, a inbound.Principal) (domain.Wallet, error) {
	if err := s.enabled(); err != nil {
		return domain.Wallet{}, err
	}
	return s.wallets.GetByUser(ctx, a.UserID)
}

func (s *WalletService) Create(ctx context.Context, a inbound.Principal, currency string) (domain.Wallet, error) {
	if err := s.enabled(); err != nil {
		return domain.Wallet{}, err
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return domain.Wallet{}, fmt.Errorf("%w: currency must be a 3-letter code", domain.ErrInvalid)
	}
	return s.wallets.Create(ctx, a.UserID, currency)
}

func (s *WalletService) Transactions(ctx context.Context, a inbound.Principal, limit, offset int) ([]domain.WalletTxn, error) {
	w, err := s.Get(ctx, a)
	if err != nil {
		return nil, err
	}
	limit, offset = page(limit, offset)
	return s.wallets.Transactions(ctx, w.ID, limit, offset)
}
