package application

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const cardNumberAttempts = 3

type cardService struct {
	core
}

func NewCardService(d Deps) port.CardService {
	return &cardService{core: newCore(d)}
}

func (s *cardService) Issue(ctx context.Context, cmd port.IssueCardCommand) (*model.Card, error) {
	w, err := s.Wallets.GetByID(ctx, cmd.WalletID)
	if err != nil {
		return nil, err
	}
	if err := w.EnsureActive(); err != nil {
		return nil, err
	}
	bank, err := s.Banks.GetByID(ctx, cmd.IssuerBankID)
	if err != nil {
		return nil, err
	}
	network := model.CardNetwork(strings.ToUpper(string(cmd.Network)))
	for attempt := 0; ; attempt++ {
		card, err := model.NewCard(w.ID, bank.ID, network, cmd.HolderName, s.now())
		if err != nil {
			return nil, err
		}
		err = s.Cards.Create(ctx, card)
		if err == nil {
			s.Logger.Info("card issued", zap.String("card_id", card.ID.String()), zap.String("network", string(network)))
			return card, nil
		}
		if !errors.Is(err, model.ErrCardNumberTaken) || attempt+1 >= cardNumberAttempts {
			return nil, err
		}
	}
}

func (s *cardService) Get(ctx context.Context, cardID uuid.UUID) (*model.Card, error) {
	return s.Cards.GetByID(ctx, cardID)
}

func (s *cardService) ListByWallet(ctx context.Context, walletID uuid.UUID) ([]*model.Card, error) {
	if _, err := s.Wallets.GetByID(ctx, walletID); err != nil {
		return nil, err
	}
	return s.Cards.ListByWallet(ctx, walletID)
}

func (s *cardService) Block(ctx context.Context, cardID uuid.UUID) (*model.Card, error) {
	return s.mutate(ctx, cardID, func(c *model.Card) error {
		c.Block(s.now())
		return nil
	})
}

func (s *cardService) Unblock(ctx context.Context, cardID uuid.UUID) (*model.Card, error) {
	return s.mutate(ctx, cardID, func(c *model.Card) error { return c.Unblock(s.now()) })
}

func (s *cardService) mutate(ctx context.Context, cardID uuid.UUID, fn func(*model.Card) error) (*model.Card, error) {
	card, err := s.Cards.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if err := fn(card); err != nil {
		return nil, err
	}
	if err := s.Cards.Update(ctx, card); err != nil {
		return nil, err
	}
	return card, nil
}
