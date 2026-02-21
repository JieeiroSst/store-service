package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
	"github.com/JIeeiroSst/wallet-service/internal/core/ports"
)

type cardService struct {
	cardRepo ports.CardRepository
	cache    ports.CacheRepository
	logger   *zap.Logger
}

func NewCardService(
	cardRepo ports.CardRepository,
	cache ports.CacheRepository,
	logger *zap.Logger,
) ports.CardUseCase {
	return &cardService{cardRepo: cardRepo, cache: cache, logger: logger}
}

func (s *cardService) IssueCard(ctx context.Context, walletID uuid.UUID, network domain.CardNetwork, holderName string) (*domain.Card, error) {
	cardNumber := generateCardNumber(network)
	hash := sha256CardNumber(cardNumber)
	now := time.Now()

	card := &domain.Card{
		ID: uuid.New(), WalletID: walletID,
		IssuerBankID: uuid.New(), 
		CardNumber: maskCardNumber(cardNumber), CardNumberHash: hash,
		HolderName: holderName, Network: network,
		ExpiryMonth: int(now.Month()), ExpiryYear: now.Year() + 3,
		Status: domain.CardStatusActive, CreatedAt: now, UpdatedAt: now,
	}

	if err := s.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}
	s.logger.Info("card issued", zap.String("card_id", card.ID.String()), zap.String("network", string(network)))
	return card, nil
}

func (s *cardService) GetCard(ctx context.Context, cardID uuid.UUID) (*domain.Card, error) {
	return s.cardRepo.GetByID(ctx, cardID)
}

func (s *cardService) BlockCard(ctx context.Context, cardID uuid.UUID) error {
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return domain.ErrCardNotFound
	}
	card.Status = domain.CardStatusBlocked
	card.UpdatedAt = time.Now()
	if err := s.cardRepo.Update(ctx, card); err != nil {
		return err
	}
	s.cache.Delete(ctx, fmt.Sprintf("card:%s", cardID))
	return nil
}

func (s *cardService) ListCardsByWallet(ctx context.Context, walletID uuid.UUID) ([]*domain.Card, error) {
	return s.cardRepo.GetByWalletID(ctx, walletID)
}

func generateCardNumber(network domain.CardNetwork) string {
	prefix := map[domain.CardNetwork]string{
		domain.CardNetworkVisa: "4", domain.CardNetworkMastercard: "5", domain.CardNetworkAmex: "37",
	}
	p := prefix[network]
	return fmt.Sprintf("%s%015d", p, time.Now().UnixNano()%1000000000000000)
}

func maskCardNumber(number string) string {
	if len(number) < 4 {
		return "****"
	}
	return "**** **** **** " + number[len(number)-4:]
}

func sha256CardNumber(number string) string {
	h := sha256.Sum256([]byte(number))
	return hex.EncodeToString(h[:])
}
