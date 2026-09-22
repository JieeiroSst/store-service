package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
)

// fakeState is the shared in-memory ledger behind fakeWalletRepo,
// fakeTransactionRepo and fakeTransferRepo, standing in for one Postgres
// database across the three real repository ports.
type fakeState struct {
	wallets         map[string]*model.Wallet
	transactions    map[string]*model.Transaction
	transfers       map[string]*model.Transfer
	pockets         map[string]*model.Pocket
	paymentRequests map[string]*model.PaymentRequest
}

func newFakeState() *fakeState {
	return &fakeState{
		wallets:         map[string]*model.Wallet{},
		transactions:    map[string]*model.Transaction{},
		transfers:       map[string]*model.Transfer{},
		pockets:         map[string]*model.Pocket{},
		paymentRequests: map[string]*model.PaymentRequest{},
	}
}

type fakeWalletRepo struct{ s *fakeState }

func (r *fakeWalletRepo) Create(_ context.Context, w *model.Wallet) error {
	r.s.wallets[w.WalletID] = w
	return nil
}

func (r *fakeWalletRepo) GetByID(_ context.Context, id string) (*model.Wallet, error) {
	w, ok := r.s.wallets[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *w
	return &cp, nil
}

func (r *fakeWalletRepo) GetByUserID(_ context.Context, userID string) (*model.Wallet, error) {
	for _, w := range r.s.wallets {
		if w.UserID == userID {
			cp := *w
			return &cp, nil
		}
	}
	return nil, port.ErrNotFound
}

func (r *fakeWalletRepo) Deposit(_ context.Context, walletID string, txn *model.Transaction) (*model.Wallet, error) {
	w, ok := r.s.wallets[walletID]
	if !ok {
		return nil, port.ErrNotFound
	}
	if w.Status != model.WalletActive {
		return nil, port.ErrWalletNotActive
	}
	w.Balance += txn.Amount
	txn.CreatedAt = time.Now()
	r.s.transactions[txn.TransactionID] = txn
	cp := *w
	return &cp, nil
}

func (r *fakeWalletRepo) Withdraw(_ context.Context, walletID string, txn *model.Transaction) (*model.Wallet, error) {
	w, ok := r.s.wallets[walletID]
	if !ok {
		return nil, port.ErrNotFound
	}
	if w.Status != model.WalletActive {
		return nil, port.ErrWalletNotActive
	}
	if w.Balance < txn.Amount {
		return nil, port.ErrInsufficientBalance
	}
	w.Balance -= txn.Amount
	txn.CreatedAt = time.Now()
	r.s.transactions[txn.TransactionID] = txn
	cp := *w
	return &cp, nil
}

func (r *fakeWalletRepo) Transfer(_ context.Context, senderID, receiverID string, transfer *model.Transfer, outTxn, inTxn *model.Transaction) (*model.Wallet, *model.Wallet, error) {
	sender, ok := r.s.wallets[senderID]
	if !ok {
		return nil, nil, port.ErrNotFound
	}
	receiver, ok := r.s.wallets[receiverID]
	if !ok {
		return nil, nil, port.ErrNotFound
	}
	if sender.Status != model.WalletActive || receiver.Status != model.WalletActive {
		return nil, nil, port.ErrWalletNotActive
	}
	if sender.Balance < transfer.Amount {
		return nil, nil, port.ErrInsufficientBalance
	}

	sender.Balance -= transfer.Amount
	receiver.Balance += transfer.Amount
	now := time.Now()
	outTxn.CreatedAt, inTxn.CreatedAt, transfer.CreatedAt = now, now, now
	r.s.transactions[outTxn.TransactionID] = outTxn
	r.s.transactions[inTxn.TransactionID] = inTxn
	r.s.transfers[transfer.TransferID] = transfer

	cs, cr := *sender, *receiver
	return &cs, &cr, nil
}

func (r *fakeWalletRepo) Reverse(_ context.Context, walletID string, original *model.Transaction, reversal *model.Transaction) (*model.Wallet, error) {
	w, ok := r.s.wallets[walletID]
	if !ok {
		return nil, port.ErrNotFound
	}
	original, ok = r.s.transactions[original.TransactionID]
	if !ok || original.Status != model.TxnCompleted {
		return nil, port.ErrTransactionNotReversible
	}

	switch original.Type {
	case model.TxnDeposit:
		if w.Balance < original.Amount {
			return nil, port.ErrInsufficientBalance
		}
		w.Balance -= original.Amount
	case model.TxnWithdraw:
		w.Balance += original.Amount
	default:
		return nil, port.ErrTransactionNotReversible
	}

	original.Status = model.TxnReversed
	r.s.transactions[reversal.TransactionID] = reversal
	cp := *w
	return &cp, nil
}

func (r *fakeWalletRepo) ReverseTransfer(_ context.Context, transfer *model.Transfer, reversalToSender, reversalFromReceiver *model.Transaction) (*model.Wallet, *model.Wallet, error) {
	sender, ok := r.s.wallets[transfer.SenderWalletID]
	if !ok {
		return nil, nil, port.ErrNotFound
	}
	receiver, ok := r.s.wallets[transfer.ReceiverWalletID]
	if !ok {
		return nil, nil, port.ErrNotFound
	}
	stored, ok := r.s.transfers[transfer.TransferID]
	if !ok || stored.Status != model.TxnCompleted {
		return nil, nil, port.ErrTransferNotReversible
	}
	if receiver.Balance < transfer.Amount {
		return nil, nil, port.ErrInsufficientBalance
	}

	sender.Balance += transfer.Amount
	receiver.Balance -= transfer.Amount
	stored.Status = model.TxnReversed
	r.s.transactions[reversalToSender.TransactionID] = reversalToSender
	r.s.transactions[reversalFromReceiver.TransactionID] = reversalFromReceiver

	cs, cr := *sender, *receiver
	return &cs, &cr, nil
}

func (r *fakeWalletRepo) UpdateStatus(_ context.Context, walletID string, status model.WalletStatus, reason *string) (*model.Wallet, error) {
	w, ok := r.s.wallets[walletID]
	if !ok {
		return nil, port.ErrNotFound
	}
	w.Status = status
	w.FrozenReason = reason
	cp := *w
	return &cp, nil
}

func (r *fakeWalletRepo) Close(_ context.Context, walletID string) (*model.Wallet, error) {
	w, ok := r.s.wallets[walletID]
	if !ok {
		return nil, port.ErrNotFound
	}
	if w.Balance != 0 {
		return nil, port.ErrWalletNotEmpty
	}
	w.Status = model.WalletClosed
	cp := *w
	return &cp, nil
}

func (r *fakeWalletRepo) UpdateLimits(_ context.Context, walletID string, dailyLimit, perTransactionLimit int64) (*model.Wallet, error) {
	w, ok := r.s.wallets[walletID]
	if !ok {
		return nil, port.ErrNotFound
	}
	w.DailyLimit = dailyLimit
	w.PerTransactionLimit = perTransactionLimit
	cp := *w
	return &cp, nil
}

type fakeTransactionRepo struct{ s *fakeState }

func (r *fakeTransactionRepo) GetByID(_ context.Context, id string) (*model.Transaction, error) {
	t, ok := r.s.transactions[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *fakeTransactionRepo) GetByReferenceID(_ context.Context, walletID, referenceID string) (*model.Transaction, error) {
	if referenceID == "" {
		return nil, port.ErrNotFound
	}
	for _, t := range r.s.transactions {
		if t.WalletID == walletID && t.ReferenceID == referenceID {
			cp := *t
			return &cp, nil
		}
	}
	return nil, port.ErrNotFound
}

func (r *fakeTransactionRepo) ListByWallet(_ context.Context, walletID string, _, _ int) ([]model.Transaction, error) {
	var out []model.Transaction
	for _, t := range r.s.transactions {
		if t.WalletID == walletID {
			out = append(out, *t)
		}
	}
	return out, nil
}

func (r *fakeTransactionRepo) ListByWalletAndDateRange(_ context.Context, walletID string, from, to time.Time) ([]model.Transaction, error) {
	var out []model.Transaction
	for _, t := range r.s.transactions {
		if t.WalletID == walletID && !t.CreatedAt.Before(from) && t.CreatedAt.Before(to) {
			out = append(out, *t)
		}
	}
	return out, nil
}

func (r *fakeTransactionRepo) SumOutgoingSince(_ context.Context, walletID string, since time.Time) (int64, error) {
	var total int64
	for _, t := range r.s.transactions {
		if t.WalletID != walletID || t.Status != model.TxnCompleted {
			continue
		}
		if t.Type != model.TxnWithdraw && t.Type != model.TxnTransferOut {
			continue
		}
		if t.CreatedAt.Before(since) {
			continue
		}
		total += t.Amount
	}
	return total, nil
}

type fakeTransferRepo struct{ s *fakeState }

func (r *fakeTransferRepo) GetByID(_ context.Context, transferID string) (*model.Transfer, error) {
	t, ok := r.s.transfers[transferID]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *fakeTransferRepo) GetByReferenceID(_ context.Context, referenceID string) (*model.Transfer, error) {
	if referenceID == "" {
		return nil, port.ErrNotFound
	}
	for _, t := range r.s.transfers {
		if t.ReferenceID == referenceID {
			cp := *t
			return &cp, nil
		}
	}
	return nil, port.ErrNotFound
}

type fakePocketRepo struct{ s *fakeState }

func (r *fakePocketRepo) Create(_ context.Context, p *model.Pocket) error {
	r.s.pockets[p.PocketID] = p
	return nil
}

func (r *fakePocketRepo) GetByID(_ context.Context, id string) (*model.Pocket, error) {
	p, ok := r.s.pockets[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *fakePocketRepo) ListByWallet(_ context.Context, walletID string) ([]model.Pocket, error) {
	var out []model.Pocket
	for _, p := range r.s.pockets {
		if p.WalletID == walletID {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (r *fakePocketRepo) MoveToPocket(_ context.Context, walletID, pocketID string, txn *model.Transaction) (*model.Wallet, *model.Pocket, error) {
	w, ok := r.s.wallets[walletID]
	if !ok {
		return nil, nil, port.ErrNotFound
	}
	p, ok := r.s.pockets[pocketID]
	if !ok || p.WalletID != walletID {
		return nil, nil, port.ErrNotFound
	}
	if w.Status != model.WalletActive {
		return nil, nil, port.ErrWalletNotActive
	}
	if w.Balance < txn.Amount {
		return nil, nil, port.ErrInsufficientBalance
	}

	w.Balance -= txn.Amount
	p.Balance += txn.Amount
	r.s.transactions[txn.TransactionID] = txn
	cw, cp := *w, *p
	return &cw, &cp, nil
}

func (r *fakePocketRepo) MoveFromPocket(_ context.Context, walletID, pocketID string, txn *model.Transaction) (*model.Wallet, *model.Pocket, error) {
	w, ok := r.s.wallets[walletID]
	if !ok {
		return nil, nil, port.ErrNotFound
	}
	p, ok := r.s.pockets[pocketID]
	if !ok || p.WalletID != walletID {
		return nil, nil, port.ErrNotFound
	}
	if p.Balance < txn.Amount {
		return nil, nil, port.ErrInsufficientBalance
	}

	w.Balance += txn.Amount
	p.Balance -= txn.Amount
	r.s.transactions[txn.TransactionID] = txn
	cw, cp := *w, *p
	return &cw, &cp, nil
}

func (r *fakePocketRepo) Close(_ context.Context, walletID, pocketID string) (*model.Wallet, error) {
	w, ok := r.s.wallets[walletID]
	if !ok {
		return nil, port.ErrNotFound
	}
	p, ok := r.s.pockets[pocketID]
	if !ok || p.WalletID != walletID {
		return nil, port.ErrNotFound
	}

	w.Balance += p.Balance
	delete(r.s.pockets, pocketID)
	cp := *w
	return &cp, nil
}

type fakePaymentRequestRepo struct{ s *fakeState }

func (r *fakePaymentRequestRepo) Create(_ context.Context, req *model.PaymentRequest) error {
	r.s.paymentRequests[req.PaymentRequestID] = req
	return nil
}

func (r *fakePaymentRequestRepo) GetByID(_ context.Context, id string) (*model.PaymentRequest, error) {
	req, ok := r.s.paymentRequests[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *req
	return &cp, nil
}

func (r *fakePaymentRequestRepo) ListByWallet(_ context.Context, walletID string) ([]model.PaymentRequest, error) {
	var out []model.PaymentRequest
	for _, req := range r.s.paymentRequests {
		if req.RequesterWalletID == walletID || (req.PayerWalletID != nil && *req.PayerWalletID == walletID) {
			out = append(out, *req)
		}
	}
	return out, nil
}

func (r *fakePaymentRequestRepo) UpdateStatus(_ context.Context, id string, status model.PaymentRequestStatus, transferID *string) error {
	req, ok := r.s.paymentRequests[id]
	if !ok || req.Status != model.PaymentRequestPending {
		return port.ErrPaymentRequestNotPending
	}
	req.Status = status
	if transferID != nil {
		req.TransferID = transferID
	}
	return nil
}
