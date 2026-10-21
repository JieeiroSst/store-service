package application

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type memDB struct {
	mu              sync.Mutex
	wallets         map[uuid.UUID]model.Wallet
	txs             map[uuid.UUID]model.Transaction
	ledger          []model.LedgerEntry
	cards           map[uuid.UUID]model.Card
	banks           map[uuid.UUID]model.Bank
	merchants       map[uuid.UUID]model.Merchant
	accounts        map[uuid.UUID]model.BankAccount
	requests        map[uuid.UUID]model.PaymentRequest
	batches         map[uuid.UUID]model.SettlementBatch
	clearing        []model.ClearingRecord
	commits         int
	rollbacks       int
	createTxFailure error
}

func newMemDB() *memDB {
	return &memDB{
		wallets:   map[uuid.UUID]model.Wallet{},
		txs:       map[uuid.UUID]model.Transaction{},
		cards:     map[uuid.UUID]model.Card{},
		banks:     map[uuid.UUID]model.Bank{},
		merchants: map[uuid.UUID]model.Merchant{},
		accounts:  map[uuid.UUID]model.BankAccount{},
		requests:  map[uuid.UUID]model.PaymentRequest{},
		batches:   map[uuid.UUID]model.SettlementBatch{},
	}
}

func cloneMap[K comparable, V any](m map[K]V) map[K]V {
	out := make(map[K]V, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

type snapshot struct {
	wallets  map[uuid.UUID]model.Wallet
	txs      map[uuid.UUID]model.Transaction
	ledger   []model.LedgerEntry
	cards    map[uuid.UUID]model.Card
	accounts map[uuid.UUID]model.BankAccount
	requests map[uuid.UUID]model.PaymentRequest
	batches  map[uuid.UUID]model.SettlementBatch
	clearing []model.ClearingRecord
}

func (db *memDB) snapshot() snapshot {
	return snapshot{
		wallets:  cloneMap(db.wallets),
		txs:      cloneMap(db.txs),
		ledger:   append([]model.LedgerEntry(nil), db.ledger...),
		cards:    cloneMap(db.cards),
		accounts: cloneMap(db.accounts),
		requests: cloneMap(db.requests),
		batches:  cloneMap(db.batches),
		clearing: append([]model.ClearingRecord(nil), db.clearing...),
	}
}

func (db *memDB) restore(s snapshot) {
	db.wallets, db.txs, db.ledger = s.wallets, s.txs, s.ledger
	db.cards, db.accounts, db.requests = s.cards, s.accounts, s.requests
	db.batches, db.clearing = s.batches, s.clearing
}

type txCtxKey struct{}

type fakeTx struct{ db *memDB }

func (f fakeTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	if ctx.Value(txCtxKey{}) != nil {
		return fn(ctx)
	}
	f.db.mu.Lock()
	snap := f.db.snapshot()
	f.db.mu.Unlock()
	if err := fn(context.WithValue(ctx, txCtxKey{}, true)); err != nil {
		f.db.mu.Lock()
		f.db.restore(snap)
		f.db.rollbacks++
		f.db.mu.Unlock()
		return err
	}
	f.db.mu.Lock()
	f.db.commits++
	f.db.mu.Unlock()
	return nil
}

type walletRepo struct{ db *memDB }

func (r walletRepo) Create(_ context.Context, w *model.Wallet) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	for _, existing := range r.db.wallets {
		if existing.UserID == w.UserID && existing.Currency == w.Currency && existing.Status != model.WalletStatusClosed {
			return model.ErrWalletExists
		}
	}
	r.db.wallets[w.ID] = *w
	return nil
}

func (r walletRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Wallet, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	w, ok := r.db.wallets[id]
	if !ok {
		return nil, model.ErrWalletNotFound
	}
	return &w, nil
}

func (r walletRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	return r.GetByID(ctx, id)
}

func (r walletRepo) GetByUserAndCurrency(_ context.Context, userID uuid.UUID, currency model.Currency) (*model.Wallet, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	for _, w := range r.db.wallets {
		if w.UserID == userID && w.Currency == currency && w.Status != model.WalletStatusClosed {
			return &w, nil
		}
	}
	return nil, model.ErrWalletNotFound
}

func (r walletRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]*model.Wallet, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := []*model.Wallet{}
	for _, w := range r.db.wallets {
		if w.UserID == userID {
			w := w
			out = append(out, &w)
		}
	}
	return out, nil
}

func (r walletRepo) Update(_ context.Context, w *model.Wallet) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	current, ok := r.db.wallets[w.ID]
	if !ok {
		return model.ErrWalletNotFound
	}
	if current.Version != w.Version {
		return model.ErrVersionConflict
	}
	if w.Balance.IsNegative() || w.FrozenAmount.IsNegative() || w.Balance.LessThan(w.FrozenAmount) {
		return errors.New("check constraint violated")
	}
	w.Version++
	r.db.wallets[w.ID] = *w
	return nil
}

type txRepo struct{ db *memDB }

func (r txRepo) Create(_ context.Context, tx *model.Transaction) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	if r.db.createTxFailure != nil {
		return r.db.createTxFailure
	}
	if tx.IdempotencyKey != "" {
		for _, existing := range r.db.txs {
			if existing.IdempotencyKey == tx.IdempotencyKey {
				return model.ErrDuplicateTransaction
			}
		}
	}
	r.db.txs[tx.ID] = *tx
	return nil
}

func (r txRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Transaction, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	tx, ok := r.db.txs[id]
	if !ok {
		return nil, model.ErrTransactionNotFound
	}
	return &tx, nil
}

func (r txRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	return r.GetByID(ctx, id)
}

func (r txRepo) GetByIdempotencyKey(_ context.Context, key string) (*model.Transaction, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	for _, tx := range r.db.txs {
		if tx.IdempotencyKey == key {
			return &tx, nil
		}
	}
	return nil, model.ErrTransactionNotFound
}

func (r txRepo) Update(_ context.Context, tx *model.Transaction) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	if _, ok := r.db.txs[tx.ID]; !ok {
		return model.ErrTransactionNotFound
	}
	r.db.txs[tx.ID] = *tx
	return nil
}

func (r txRepo) List(_ context.Context, f model.TransactionFilter) ([]*model.Transaction, int64, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := []*model.Transaction{}
	for _, tx := range r.db.txs {
		if tx.WalletID != f.WalletID {
			continue
		}
		if len(f.Types) > 0 && !contains(f.Types, tx.Type) {
			continue
		}
		if len(f.Statuses) > 0 && !contains(f.Statuses, tx.Status) {
			continue
		}
		tx := tx
		out = append(out, &tx)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, int64(len(out)), nil
}

func contains[T comparable](items []T, v T) bool {
	for _, it := range items {
		if it == v {
			return true
		}
	}
	return false
}

func (r txRepo) ListUnbatchedCaptures(_ context.Context, merchantID uuid.UUID) ([]*model.Transaction, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := []*model.Transaction{}
	for _, tx := range r.db.txs {
		if tx.MerchantID.UUID == merchantID && tx.Status == model.TransactionStatusCaptured && tx.CardID.Valid && !tx.BatchID.Valid {
			tx := tx
			out = append(out, &tx)
		}
	}
	return out, nil
}

func (r txRepo) ListByBatch(_ context.Context, batchID uuid.UUID) ([]*model.Transaction, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := []*model.Transaction{}
	for _, tx := range r.db.txs {
		if tx.BatchID.Valid && tx.BatchID.UUID == batchID {
			tx := tx
			out = append(out, &tx)
		}
	}
	return out, nil
}

func (r txRepo) AssignBatch(_ context.Context, ids []uuid.UUID, status model.TransactionStatus, batchID uuid.UUID) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	for _, id := range ids {
		tx := r.db.txs[id]
		tx.Status = status
		tx.BatchID = model.NullID(batchID)
		r.db.txs[id] = tx
	}
	return nil
}

type ledgerRepo struct{ db *memDB }

func (r ledgerRepo) Append(_ context.Context, entries ...*model.LedgerEntry) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	for _, e := range entries {
		r.db.ledger = append(r.db.ledger, *e)
	}
	return nil
}

func (r ledgerRepo) List(_ context.Context, walletID uuid.UUID, _, _ *time.Time, _, _ int) ([]*model.LedgerEntry, int64, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := []*model.LedgerEntry{}
	for i := len(r.db.ledger) - 1; i >= 0; i-- {
		if e := r.db.ledger[i]; e.WalletID == walletID {
			out = append(out, &e)
		}
	}
	return out, int64(len(out)), nil
}

func (r ledgerRepo) SumDebitsSince(_ context.Context, walletID uuid.UUID, since time.Time) (decimal.Decimal, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	sum := decimal.Zero
	for _, e := range r.db.ledger {
		if e.WalletID == walletID && e.Direction == model.EntryDebit && !e.CreatedAt.Before(since) {
			sum = sum.Add(e.Amount)
		}
	}
	return sum, nil
}

func (r ledgerRepo) Summarize(_ context.Context, walletID uuid.UUID, from, to time.Time) (model.LedgerSummary, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	s := model.LedgerSummary{}
	for _, e := range r.db.ledger {
		if e.WalletID != walletID {
			continue
		}
		if e.CreatedAt.Before(from) {
			s.OpeningBalance = e.BalanceAfter
			continue
		}
		if !e.CreatedAt.Before(to) {
			continue
		}
		if e.Direction == model.EntryCredit {
			s.TotalCredit = s.TotalCredit.Add(e.Amount)
			s.CreditCount++
		} else {
			s.TotalDebit = s.TotalDebit.Add(e.Amount)
			s.DebitCount++
		}
	}
	s.ClosingBalance = s.OpeningBalance.Add(s.TotalCredit).Sub(s.TotalDebit)
	return s, nil
}

type cardRepo struct{ db *memDB }

func (r cardRepo) Create(_ context.Context, c *model.Card) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	r.db.cards[c.ID] = *c
	return nil
}

func (r cardRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Card, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	c, ok := r.db.cards[id]
	if !ok {
		return nil, model.ErrCardNotFound
	}
	return &c, nil
}

func (r cardRepo) ListByWallet(_ context.Context, walletID uuid.UUID) ([]*model.Card, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := []*model.Card{}
	for _, c := range r.db.cards {
		if c.WalletID == walletID {
			c := c
			out = append(out, &c)
		}
	}
	return out, nil
}

func (r cardRepo) Update(_ context.Context, c *model.Card) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	r.db.cards[c.ID] = *c
	return nil
}

type bankRepo struct{ db *memDB }

func (r bankRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Bank, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	b, ok := r.db.banks[id]
	if !ok {
		return nil, model.ErrBankNotFound
	}
	return &b, nil
}

type merchantRepo struct{ db *memDB }

func (r merchantRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Merchant, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	m, ok := r.db.merchants[id]
	if !ok {
		return nil, model.ErrMerchantNotFound
	}
	return &m, nil
}

type accountRepo struct{ db *memDB }

func (r accountRepo) Create(_ context.Context, a *model.BankAccount) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	for _, existing := range r.db.accounts {
		if existing.UserID == a.UserID && existing.AccountNumberHash == a.AccountNumberHash && existing.Status == model.BankAccountActive {
			return model.ErrBankAccountFound
		}
	}
	r.db.accounts[a.ID] = *a
	return nil
}

func (r accountRepo) GetByID(_ context.Context, id uuid.UUID) (*model.BankAccount, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	a, ok := r.db.accounts[id]
	if !ok {
		return nil, model.ErrBankAccountAbsent
	}
	return &a, nil
}

func (r accountRepo) ListActiveByUser(_ context.Context, userID uuid.UUID) ([]*model.BankAccount, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := []*model.BankAccount{}
	for _, a := range r.db.accounts {
		if a.UserID == userID && a.Status == model.BankAccountActive {
			a := a
			out = append(out, &a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r accountRepo) Update(_ context.Context, a *model.BankAccount) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	r.db.accounts[a.ID] = *a
	return nil
}

func (r accountRepo) ClearDefault(_ context.Context, userID uuid.UUID) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	for id, a := range r.db.accounts {
		if a.UserID == userID {
			a.IsDefault = false
			r.db.accounts[id] = a
		}
	}
	return nil
}

type requestRepo struct{ db *memDB }

func (r requestRepo) Create(_ context.Context, p *model.PaymentRequest) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	r.db.requests[p.ID] = *p
	return nil
}

func (r requestRepo) GetByID(_ context.Context, id uuid.UUID) (*model.PaymentRequest, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	p, ok := r.db.requests[id]
	if !ok {
		return nil, model.ErrPaymentRequestNotFound
	}
	return &p, nil
}

func (r requestRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.PaymentRequest, error) {
	return r.GetByID(ctx, id)
}

func (r requestRepo) Update(_ context.Context, p *model.PaymentRequest) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	r.db.requests[p.ID] = *p
	return nil
}

func (r requestRepo) List(_ context.Context, f model.PaymentRequestFilter) ([]*model.PaymentRequest, int64, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := []*model.PaymentRequest{}
	for _, p := range r.db.requests {
		owner := p.PayerUserID
		if f.Role == model.RoleOutgoing {
			owner = p.RequesterUserID
		}
		if owner == f.UserID {
			p := p
			out = append(out, &p)
		}
	}
	return out, int64(len(out)), nil
}

type batchRepo struct{ db *memDB }

func (r batchRepo) Create(_ context.Context, b *model.SettlementBatch) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	stored := *b
	stored.Transactions = nil
	r.db.batches[b.ID] = stored
	return nil
}

func (r batchRepo) GetByIDForUpdate(_ context.Context, id uuid.UUID) (*model.SettlementBatch, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	b, ok := r.db.batches[id]
	if !ok {
		return nil, model.ErrBatchNotFound
	}
	return &b, nil
}

func (r batchRepo) Update(_ context.Context, b *model.SettlementBatch) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	r.db.batches[b.ID] = *b
	return nil
}

type clearingRepo struct{ db *memDB }

func (r clearingRepo) CreateBulk(_ context.Context, records []*model.ClearingRecord) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	for _, rec := range records {
		r.db.clearing = append(r.db.clearing, *rec)
	}
	return nil
}

func (r clearingRepo) ListByBatch(_ context.Context, batchID uuid.UUID) ([]*model.ClearingRecord, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()
	out := []*model.ClearingRecord{}
	for _, rec := range r.db.clearing {
		if rec.BatchID == batchID {
			rec := rec
			out = append(out, &rec)
		}
	}
	return out, nil
}

type memCache struct {
	mu   sync.Mutex
	data map[string]string
}

func (c *memCache) Get(_ context.Context, key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[key], nil
}

func (c *memCache) Set(_ context.Context, key, value string, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return nil
}

func (c *memCache) Delete(_ context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, k := range keys {
		delete(c.data, k)
	}
	return nil
}

type plainPins struct{}

func (plainPins) Hash(pin string) (string, error) { return "hashed:" + pin, nil }
func (plainPins) Compare(hash, pin string) bool   { return hash == "hashed:"+pin }

type fakeBank struct {
	mu         sync.Mutex
	collectErr error
	payoutErr  error
	collects   []port.BankTransfer
	payouts    []port.BankTransfer
}

func (b *fakeBank) Collect(_ context.Context, t port.BankTransfer) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.collects = append(b.collects, t)
	if b.collectErr != nil {
		return "", b.collectErr
	}
	return "BANK-C-" + t.Reference.String()[:8], nil
}

func (b *fakeBank) Payout(_ context.Context, t port.BankTransfer) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.payouts = append(b.payouts, t)
	if b.payoutErr != nil {
		return "", b.payoutErr
	}
	return "BANK-P-" + t.Reference.String()[:8], nil
}
