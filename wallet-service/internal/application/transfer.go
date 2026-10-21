package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type transferService struct {
	core
}

func NewTransferService(d Deps) port.TransferService {
	return &transferService{core: newCore(d)}
}

type transferInput struct {
	idempotencyKey string
	userID         uuid.UUID
	fromWalletID   uuid.UUID
	toWalletID     uuid.UUID
	amount         decimal.Decimal
	description    string
	metadata       model.Metadata
}

func (s *transferService) Transfer(ctx context.Context, cmd port.TransferCommand) (*model.Transaction, error) {
	if err := validateIdempotencyKey(cmd.IdempotencyKey); err != nil {
		return nil, err
	}
	from, err := s.ownedWallet(ctx, cmd.FromWalletID, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if prior, err := s.replay(ctx, cmd.IdempotencyKey, cmd.FromWalletID, model.TransactionTypeTransferOut); err != nil || prior != nil {
		return prior, err
	}
	toID, err := s.resolveRecipient(ctx, from, cmd)
	if err != nil {
		return nil, err
	}
	if err := model.ValidateAmount(cmd.Amount, from.Currency); err != nil {
		return nil, err
	}
	if err := s.verifyPin(ctx, from.ID, cmd.UserID, cmd.Pin); err != nil {
		return nil, err
	}
	in := transferInput{
		idempotencyKey: cmd.IdempotencyKey,
		userID:         cmd.UserID,
		fromWalletID:   from.ID,
		toWalletID:     toID,
		amount:         cmd.Amount,
		description:    cmd.Description,
	}
	var out *model.Transaction
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		out, err = s.transferWithinTx(ctx, in)
		return err
	})
	if err != nil {
		return s.resolveDuplicate(ctx, err, cmd.IdempotencyKey, cmd.FromWalletID, model.TransactionTypeTransferOut)
	}
	s.invalidateWallets(ctx, from.ID, toID)
	s.Logger.Info("transfer completed", zap.String("tx_id", out.ID.String()), zap.String("to_wallet", toID.String()))
	return out, nil
}

func (s *transferService) resolveRecipient(ctx context.Context, from *model.Wallet, cmd port.TransferCommand) (uuid.UUID, error) {
	toID := cmd.ToWalletID
	if toID == uuid.Nil {
		if cmd.ToUserID == uuid.Nil {
			return uuid.Nil, model.Invalid("to_wallet_id or to_user_id is required")
		}
		to, err := s.Wallets.GetByUserAndCurrency(ctx, cmd.ToUserID, from.Currency)
		if err != nil {
			return uuid.Nil, err
		}
		toID = to.ID
	}
	if toID == from.ID {
		return uuid.Nil, model.ErrSameWallet
	}
	return toID, nil
}

func (c *core) transferWithinTx(ctx context.Context, in transferInput) (*model.Transaction, error) {
	if in.fromWalletID == in.toWalletID {
		return nil, model.ErrSameWallet
	}
	wallets, err := c.lockWallets(ctx, in.fromWalletID, in.toWalletID)
	if err != nil {
		return nil, err
	}
	src, dst := wallets[in.fromWalletID], wallets[in.toWalletID]
	if err := src.EnsureOwner(in.userID); err != nil {
		return nil, err
	}
	if err := src.EnsureActive(); err != nil {
		return nil, err
	}
	if err := dst.EnsureActive(); err != nil {
		return nil, err
	}
	if src.Currency != dst.Currency {
		return nil, model.ErrCurrencyMismatch
	}
	now := c.now()
	out := model.NewTransaction(src, model.TransactionTypeTransferOut, model.TransactionStatusCompleted, in.amount, now)
	out.IdempotencyKey = in.idempotencyKey
	out.Fee = c.Fees.Transfer(in.amount, src.Currency)
	out.CounterpartyWalletID = model.NullID(dst.ID)
	out.Description = in.description
	out.Metadata = in.metadata
	total := out.TotalDebit()
	if err := c.checkLimits(ctx, src, total); err != nil {
		return nil, err
	}
	if err := src.Debit(total, now); err != nil {
		return nil, err
	}
	if err := dst.Credit(in.amount, now); err != nil {
		return nil, err
	}

	inbound := model.NewTransaction(dst, model.TransactionTypeTransferIn, model.TransactionStatusCompleted, in.amount, now)
	inbound.CounterpartyWalletID = model.NullID(src.ID)
	inbound.ReferenceID = model.NullID(out.ID)
	inbound.Description = in.description
	inbound.Metadata = in.metadata

	if err := c.persistDebit(ctx, src, out, total, now); err != nil {
		return nil, err
	}
	if err := c.persistCredit(ctx, dst, inbound, in.amount, now); err != nil {
		return nil, err
	}
	out.ReferenceID = model.NullID(inbound.ID)
	if err := c.Transactions.Update(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}
