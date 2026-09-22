package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"github.com/google/uuid"
)

type transactionService struct {
	wallets      port.WalletRepository
	transactions port.TransactionRepository
	transfers    port.TransferRepository
}

func NewTransactionService(
	wallets port.WalletRepository,
	transactions port.TransactionRepository,
	transfers port.TransferRepository,
) port.TransactionUsecase {
	return &transactionService{wallets: wallets, transactions: transactions, transfers: transfers}
}

func (s *transactionService) Deposit(ctx context.Context, walletID string, amount int64, referenceID, description string) (*model.Transaction, error) {
	if amount <= 0 {
		return nil, port.ErrInvalidAmount
	}
	if existing, ok, err := s.existingTransaction(ctx, walletID, referenceID); err != nil {
		return nil, err
	} else if ok {
		return existing, nil
	}

	wallet, err := s.wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}

	txn := &model.Transaction{
		TransactionID: uuid.NewString(),
		WalletID:      walletID,
		Type:          model.TxnDeposit,
		Amount:        amount,
		Currency:      wallet.Currency,
		Status:        model.TxnCompleted,
		ReferenceID:   referenceID,
		Description:   description,
	}
	if _, err := s.wallets.Deposit(ctx, walletID, txn); err != nil {
		return nil, fmt.Errorf("deposit: %w", err)
	}
	return txn, nil
}

func (s *transactionService) Withdraw(ctx context.Context, walletID string, amount int64, referenceID, description string) (*model.Transaction, error) {
	if amount <= 0 {
		return nil, port.ErrInvalidAmount
	}
	if existing, ok, err := s.existingTransaction(ctx, walletID, referenceID); err != nil {
		return nil, err
	} else if ok {
		return existing, nil
	}

	wallet, err := s.wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}
	if err := s.checkLimits(ctx, wallet, amount); err != nil {
		return nil, err
	}

	txn := &model.Transaction{
		TransactionID: uuid.NewString(),
		WalletID:      walletID,
		Type:          model.TxnWithdraw,
		Amount:        amount,
		Currency:      wallet.Currency,
		Status:        model.TxnCompleted,
		ReferenceID:   referenceID,
		Description:   description,
	}
	if _, err := s.wallets.Withdraw(ctx, walletID, txn); err != nil {
		return nil, fmt.Errorf("withdraw: %w", err)
	}
	return txn, nil
}

func (s *transactionService) Transfer(ctx context.Context, senderWalletID, receiverWalletID string, amount int64, referenceID, description string) (*model.Transfer, error) {
	if amount <= 0 {
		return nil, port.ErrInvalidAmount
	}
	if senderWalletID == receiverWalletID {
		return nil, port.ErrSameWallet
	}
	if referenceID != "" {
		existing, err := s.transfers.GetByReferenceID(ctx, referenceID)
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, port.ErrNotFound) {
			return nil, fmt.Errorf("check existing transfer: %w", err)
		}
	}

	sender, err := s.wallets.GetByID(ctx, senderWalletID)
	if err != nil {
		return nil, fmt.Errorf("get sender wallet: %w", err)
	}
	receiver, err := s.wallets.GetByID(ctx, receiverWalletID)
	if err != nil {
		return nil, fmt.Errorf("get receiver wallet: %w", err)
	}
	if sender.Currency != receiver.Currency {
		return nil, port.ErrCurrencyMismatch
	}
	if err := s.checkLimits(ctx, sender, amount); err != nil {
		return nil, err
	}

	outTxn := &model.Transaction{
		TransactionID:        uuid.NewString(),
		WalletID:             senderWalletID,
		Type:                 model.TxnTransferOut,
		Amount:               amount,
		Currency:             sender.Currency,
		Status:               model.TxnCompleted,
		ReferenceID:          referenceID,
		CounterpartyWalletID: &receiverWalletID,
		Description:          description,
	}
	inTxn := &model.Transaction{
		TransactionID:        uuid.NewString(),
		WalletID:             receiverWalletID,
		Type:                 model.TxnTransferIn,
		Amount:               amount,
		Currency:             receiver.Currency,
		Status:               model.TxnCompleted,
		ReferenceID:          referenceID,
		CounterpartyWalletID: &senderWalletID,
		Description:          description,
	}
	transfer := &model.Transfer{
		TransferID:       uuid.NewString(),
		SenderWalletID:   senderWalletID,
		ReceiverWalletID: receiverWalletID,
		Amount:           amount,
		Currency:         sender.Currency,
		Status:           model.TxnCompleted,
		ReferenceID:      referenceID,
		OutTransactionID: outTxn.TransactionID,
		InTransactionID:  inTxn.TransactionID,
		Description:      description,
	}

	if _, _, err := s.wallets.Transfer(ctx, senderWalletID, receiverWalletID, transfer, outTxn, inTxn); err != nil {
		return nil, fmt.Errorf("transfer: %w", err)
	}
	return transfer, nil
}

func (s *transactionService) GetTransaction(ctx context.Context, transactionID string) (*model.Transaction, error) {
	return s.transactions.GetByID(ctx, transactionID)
}

func (s *transactionService) ListTransactions(ctx context.Context, walletID string, limit, offset int) ([]model.Transaction, error) {
	return s.transactions.ListByWallet(ctx, walletID, limit, offset)
}

func (s *transactionService) Statement(ctx context.Context, walletID string, from, to time.Time) ([]model.Transaction, error) {
	return s.transactions.ListByWalletAndDateRange(ctx, walletID, from, to)
}

func (s *transactionService) GetTransfer(ctx context.Context, transferID string) (*model.Transfer, error) {
	return s.transfers.GetByID(ctx, transferID)
}

func (s *transactionService) Reverse(ctx context.Context, transactionID, reason string) (*model.Transaction, error) {
	original, err := s.transactions.GetByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w", err)
	}
	if original.Status != model.TxnCompleted {
		return nil, port.ErrTransactionNotReversible
	}
	if original.Type != model.TxnDeposit && original.Type != model.TxnWithdraw {
		return nil, port.ErrTransactionNotReversible
	}

	reversal := &model.Transaction{
		TransactionID: uuid.NewString(),
		WalletID:      original.WalletID,
		Type:          model.TxnReversal,
		Amount:        original.Amount,
		Currency:      original.Currency,
		Status:        model.TxnCompleted,
		Description:   fmt.Sprintf("reversal of %s: %s", original.TransactionID, reason),
	}
	if _, err := s.wallets.Reverse(ctx, original.WalletID, original, reversal); err != nil {
		return nil, fmt.Errorf("reverse: %w", err)
	}
	return reversal, nil
}

func (s *transactionService) ReverseTransfer(ctx context.Context, transferID, reason string) (*model.Transfer, error) {
	transfer, err := s.transfers.GetByID(ctx, transferID)
	if err != nil {
		return nil, fmt.Errorf("get transfer: %w", err)
	}
	if transfer.Status != model.TxnCompleted {
		return nil, port.ErrTransferNotReversible
	}

	reversalToSender := &model.Transaction{
		TransactionID:        uuid.NewString(),
		WalletID:             transfer.SenderWalletID,
		Type:                 model.TxnReversal,
		Amount:               transfer.Amount,
		Currency:             transfer.Currency,
		Status:               model.TxnCompleted,
		CounterpartyWalletID: &transfer.ReceiverWalletID,
		Description:          fmt.Sprintf("reversal of transfer %s: %s", transfer.TransferID, reason),
	}
	reversalFromReceiver := &model.Transaction{
		TransactionID:        uuid.NewString(),
		WalletID:             transfer.ReceiverWalletID,
		Type:                 model.TxnReversal,
		Amount:               transfer.Amount,
		Currency:             transfer.Currency,
		Status:               model.TxnCompleted,
		CounterpartyWalletID: &transfer.SenderWalletID,
		Description:          fmt.Sprintf("reversal of transfer %s: %s", transfer.TransferID, reason),
	}

	if _, _, err := s.wallets.ReverseTransfer(ctx, transfer, reversalToSender, reversalFromReceiver); err != nil {
		return nil, fmt.Errorf("reverse transfer: %w", err)
	}
	transfer.Status = model.TxnReversed
	return transfer, nil
}

func (s *transactionService) checkLimits(ctx context.Context, wallet *model.Wallet, amount int64) error {
	if wallet.PerTransactionLimit > 0 && amount > wallet.PerTransactionLimit {
		return port.ErrPerTransactionLimitExceeded
	}
	if wallet.DailyLimit > 0 {
		since := time.Now().Truncate(24 * time.Hour)
		spent, err := s.transactions.SumOutgoingSince(ctx, wallet.WalletID, since)
		if err != nil {
			return fmt.Errorf("sum daily outgoing: %w", err)
		}
		if spent+amount > wallet.DailyLimit {
			return port.ErrDailyLimitExceeded
		}
	}
	return nil
}

func (s *transactionService) existingTransaction(ctx context.Context, walletID, referenceID string) (*model.Transaction, bool, error) {
	if referenceID == "" {
		return nil, false, nil
	}
	existing, err := s.transactions.GetByReferenceID(ctx, walletID, referenceID)
	if err == nil {
		return existing, true, nil
	}
	if errors.Is(err, port.ErrNotFound) {
		return nil, false, nil
	}
	return nil, false, fmt.Errorf("check existing transaction: %w", err)
}
