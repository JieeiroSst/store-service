package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const (
	walletCacheKey      = "wallet:id:%s"
	transactionCacheKey = "tx:%s"
	walletCacheTTL      = time.Minute
	transactionCacheTTL = 30 * time.Minute
	maxIdempotencyKey   = 64
)

type Deps struct {
	Tx              port.TxManager
	Wallets         port.WalletRepository
	Transactions    port.TransactionRepository
	Ledger          port.LedgerRepository
	Cards           port.CardRepository
	Banks           port.BankRepository
	Merchants       port.MerchantRepository
	BankAccounts    port.BankAccountRepository
	PaymentRequests port.PaymentRequestRepository
	Batches         port.SettlementBatchRepository
	Clearing        port.ClearingRepository
	Cache           port.Cache
	Pins            port.PinHasher
	BankGateway     port.BankGateway
	Fees            model.FeeSchedule
	PaymentRequest  PaymentRequestPolicy
	Clock           func() time.Time
	Logger          *zap.Logger
}

type PaymentRequestPolicy struct {
	TTL time.Duration
}

type core struct {
	Deps
}

func newCore(d Deps) core {
	if d.Clock == nil {
		d.Clock = time.Now
	}
	if d.Logger == nil {
		d.Logger = zap.NewNop()
	}
	if d.PaymentRequest.TTL <= 0 {
		d.PaymentRequest.TTL = 7 * 24 * time.Hour
	}
	return core{Deps: d}
}

func (c *core) now() time.Time { return c.Clock().UTC() }

func validateIdempotencyKey(key string) error {
	if key == "" {
		return model.ErrIdempotencyKeyMissing
	}
	if len(key) > maxIdempotencyKey {
		return model.Invalid(fmt.Sprintf("idempotency key must be at most %d characters", maxIdempotencyKey))
	}
	return nil
}

func (c *core) replay(ctx context.Context, key string, walletID uuid.UUID, types ...model.TransactionType) (*model.Transaction, error) {
	if key == "" {
		return nil, nil
	}
	tx, err := c.Transactions.GetByIdempotencyKey(ctx, key)
	if errors.Is(err, model.ErrTransactionNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if tx.WalletID != walletID {
		return nil, model.ErrIdempotencyKeyReused
	}
	if len(types) == 0 {
		return tx, nil
	}
	for _, t := range types {
		if tx.Type == t {
			return tx, nil
		}
	}
	return nil, model.ErrIdempotencyKeyReused
}

func (c *core) resolveDuplicate(ctx context.Context, err error, key string, walletID uuid.UUID, types ...model.TransactionType) (*model.Transaction, error) {
	if !errors.Is(err, model.ErrDuplicateTransaction) {
		return nil, err
	}
	tx, rerr := c.replay(ctx, key, walletID, types...)
	if rerr != nil {
		return nil, rerr
	}
	if tx == nil {
		return nil, err
	}
	return tx, nil
}

func (c *core) ownedWallet(ctx context.Context, walletID, userID uuid.UUID) (*model.Wallet, error) {
	w, err := c.Wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if err := w.EnsureOwner(userID); err != nil {
		return nil, err
	}
	return w, nil
}

func (c *core) verifyPin(ctx context.Context, walletID, userID uuid.UUID, pin string) error {
	var verdict error
	err := c.Tx.WithinTx(ctx, func(ctx context.Context) error {
		verdict = nil
		w, err := c.Wallets.GetByIDForUpdate(ctx, walletID)
		if err != nil {
			return err
		}
		if err := w.EnsureOwner(userID); err != nil {
			return err
		}
		now := c.now()
		if !w.HasPin {
			return model.ErrPinNotSet
		}
		if w.PinLocked(now) {
			return model.ErrPinLocked
		}
		if c.Pins.Compare(w.PinHash, pin) {
			if w.ResetPinFailures(now) {
				return c.Wallets.Update(ctx, w)
			}
			return nil
		}
		if w.RegisterPinFailure(now) {
			verdict = model.ErrPinLocked
		} else {
			verdict = model.ErrPinInvalid
		}
		return c.Wallets.Update(ctx, w)
	})
	if err != nil {
		return err
	}
	return verdict
}

func (c *core) checkLimits(ctx context.Context, w *model.Wallet, amount decimal.Decimal) error {
	used, err := c.Ledger.SumDebitsSince(ctx, w.ID, model.StartOfDay(c.now()))
	if err != nil {
		return err
	}
	return model.LimitsFor(w.Tier, w.Currency).Check(amount, used)
}

func (c *core) lockWallets(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID]*model.Wallet, error) {
	ordered := append([]uuid.UUID(nil), ids...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].String() < ordered[j].String() })
	out := make(map[uuid.UUID]*model.Wallet, len(ordered))
	for _, id := range ordered {
		if _, ok := out[id]; ok {
			continue
		}
		w, err := c.Wallets.GetByIDForUpdate(ctx, id)
		if err != nil {
			return nil, err
		}
		out[id] = w
	}
	return out, nil
}

func (c *core) invalidateWallets(ctx context.Context, ids ...uuid.UUID) {
	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, fmt.Sprintf(walletCacheKey, id))
	}
	c.invalidate(ctx, keys...)
}

func (c *core) invalidateTransactions(ctx context.Context, ids ...uuid.UUID) {
	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, fmt.Sprintf(transactionCacheKey, id))
	}
	c.invalidate(ctx, keys...)
}

func (c *core) invalidate(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	if err := c.Cache.Delete(ctx, keys...); err != nil {
		c.Logger.Warn("cache invalidation failed", zap.Strings("keys", keys), zap.Error(err))
	}
}

func cacheGet[T any](ctx context.Context, c *core, key string) (*T, bool) {
	raw, err := c.Cache.Get(ctx, key)
	if err != nil || raw == "" {
		return nil, false
	}
	var v T
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, false
	}
	return &v, true
}

func cacheSet(ctx context.Context, c *core, key string, v any, ttl time.Duration) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	if err := c.Cache.Set(ctx, key, string(b), ttl); err != nil {
		c.Logger.Debug("cache set failed", zap.String("key", key), zap.Error(err))
	}
}

func (c *core) persistDebit(ctx context.Context, w *model.Wallet, tx *model.Transaction, amount decimal.Decimal, now time.Time) error {
	if err := c.Transactions.Create(ctx, tx); err != nil {
		return err
	}
	if err := c.Ledger.Append(ctx, model.NewLedgerEntry(w, tx.ID, model.EntryDebit, amount, now)); err != nil {
		return err
	}
	return c.Wallets.Update(ctx, w)
}

func (c *core) persistCredit(ctx context.Context, w *model.Wallet, tx *model.Transaction, amount decimal.Decimal, now time.Time) error {
	if err := c.Transactions.Create(ctx, tx); err != nil {
		return err
	}
	if err := c.Ledger.Append(ctx, model.NewLedgerEntry(w, tx.ID, model.EntryCredit, amount, now)); err != nil {
		return err
	}
	return c.Wallets.Update(ctx, w)
}
