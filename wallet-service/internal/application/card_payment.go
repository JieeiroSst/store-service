package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

var cardTypes = []model.TransactionType{
	model.TransactionTypeAuthorization,
	model.TransactionTypeCapture,
	model.TransactionTypeVoid,
}

type cardPaymentService struct {
	core
}

func NewCardPaymentService(d Deps) port.CardPaymentService {
	return &cardPaymentService{core: newCore(d)}
}

func (s *cardPaymentService) Authorize(ctx context.Context, cmd port.AuthorizeCommand) (*model.Transaction, error) {
	if err := validateIdempotencyKey(cmd.IdempotencyKey); err != nil {
		return nil, err
	}
	card, err := s.Cards.GetByID(ctx, cmd.CardID)
	if err != nil {
		return nil, err
	}
	if prior, err := s.replay(ctx, cmd.IdempotencyKey, card.WalletID, cardTypes...); err != nil || prior != nil {
		return prior, err
	}
	now := s.now()
	if err := card.EnsureUsable(now); err != nil {
		return nil, err
	}
	merchant, err := s.Merchants.GetByID(ctx, cmd.MerchantID)
	if err != nil {
		return nil, err
	}
	issuer, err := s.Banks.GetByID(ctx, card.IssuerBankID)
	if err != nil {
		return nil, err
	}
	currency := model.Currency(strings.ToUpper(string(cmd.Currency)))
	if err := model.ValidateAmount(cmd.Amount, currency); err != nil {
		return nil, err
	}

	var tx *model.Transaction
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		w, err := s.Wallets.GetByIDForUpdate(ctx, card.WalletID)
		if err != nil {
			return err
		}
		if err := w.EnsureActive(); err != nil {
			return err
		}
		if w.Currency != currency {
			return model.ErrCurrencyMismatch
		}
		if err := s.checkLimits(ctx, w, cmd.Amount); err != nil {
			return err
		}
		now := s.now()
		if err := w.Hold(cmd.Amount, now); err != nil {
			s.Logger.Warn("authorization declined", zap.String("wallet_id", w.ID.String()), zap.Error(err))
			return err
		}
		tx = model.NewTransaction(w, model.TransactionTypeAuthorization, model.TransactionStatusPending, cmd.Amount, now)
		tx.IdempotencyKey = cmd.IdempotencyKey
		tx.MerchantID = model.NullID(merchant.ID)
		tx.AcquirerBankID = model.NullID(merchant.AcquirerBankID)
		tx.IssuerBankID = model.NullID(issuer.ID)
		tx.CardID = model.NullID(card.ID)
		tx.CardNetwork = card.Network
		tx.Fee = s.Fees.CardInterchange(cmd.Amount, card.Network, currency)
		tx.Description = cmd.Description
		tx.Metadata = cmd.Metadata
		authCode, err := generateAuthCode()
		if err != nil {
			return err
		}
		tx.Authorize(authCode, now)
		if err := s.Transactions.Create(ctx, tx); err != nil {
			return err
		}
		return s.Wallets.Update(ctx, w)
	})
	if err != nil {
		return s.resolveDuplicate(ctx, err, cmd.IdempotencyKey, card.WalletID, cardTypes...)
	}
	s.invalidateWallets(ctx, card.WalletID)
	s.Logger.Info("card authorized", zap.String("tx_id", tx.ID.String()), zap.String("auth_code", tx.AuthorizationCode))
	return tx, nil
}

func (s *cardPaymentService) Capture(ctx context.Context, txID uuid.UUID) (*model.Transaction, error) {
	return s.settleHold(ctx, txID, func(ctx context.Context, w *model.Wallet, tx *model.Transaction) error {
		now := s.now()
		if err := tx.Capture(now); err != nil {
			return err
		}
		if err := w.CaptureHold(tx.Amount, now); err != nil {
			return err
		}
		return s.Ledger.Append(ctx, model.NewLedgerEntry(w, tx.ID, model.EntryDebit, tx.Amount, now))
	})
}

func (s *cardPaymentService) Void(ctx context.Context, txID uuid.UUID) (*model.Transaction, error) {
	return s.settleHold(ctx, txID, func(_ context.Context, w *model.Wallet, tx *model.Transaction) error {
		now := s.now()
		if err := tx.Void(now); err != nil {
			return err
		}
		w.ReleaseHold(tx.Amount, now)
		return nil
	})
}

func (s *cardPaymentService) settleHold(ctx context.Context, txID uuid.UUID, apply func(context.Context, *model.Wallet, *model.Transaction) error) (*model.Transaction, error) {
	var out *model.Transaction
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		tx, err := s.Transactions.GetByIDForUpdate(ctx, txID)
		if err != nil {
			return err
		}
		if !tx.CardID.Valid {
			return model.ErrTransactionState
		}
		w, err := s.Wallets.GetByIDForUpdate(ctx, tx.WalletID)
		if err != nil {
			return err
		}
		if err := apply(ctx, w, tx); err != nil {
			return err
		}
		if err := s.Wallets.Update(ctx, w); err != nil {
			return err
		}
		out = tx
		return s.Transactions.Update(ctx, tx)
	})
	if err != nil {
		return nil, err
	}
	s.invalidateWallets(ctx, out.WalletID)
	s.invalidateTransactions(ctx, out.ID)
	return out, nil
}

func generateAuthCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(b)), nil
}
