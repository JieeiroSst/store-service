package bankgateway

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

var ErrDeclined = errors.New("simulated bank declined the transfer")

type simulated struct {
	declineAbove decimal.Decimal
	logger       *zap.Logger
}

func NewSimulated(declineAbove decimal.Decimal, logger *zap.Logger) port.BankGateway {
	return &simulated{declineAbove: declineAbove, logger: logger}
}

func (s *simulated) Collect(_ context.Context, t port.BankTransfer) (string, error) {
	return s.process("collect", t)
}

func (s *simulated) Payout(_ context.Context, t port.BankTransfer) (string, error) {
	return s.process("payout", t)
}

func (s *simulated) process(op string, t port.BankTransfer) (string, error) {
	if s.declineAbove.IsPositive() && t.Amount.GreaterThan(s.declineAbove) {
		return "", ErrDeclined
	}
	ref := fmt.Sprintf("SIM-%s-%s", op, t.Reference.String()[:8])
	s.logger.Info("simulated bank transfer",
		zap.String("op", op),
		zap.String("reference", t.Reference.String()),
		zap.String("account_id", t.Account.ID.String()),
		zap.String("amount", t.Amount.String()),
		zap.String("currency", string(t.Currency)))
	return ref, nil
}
