package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) port.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(ctx context.Context, wallet *model.Wallet) error {
	now := time.Now()
	wallet.CreatedAt, wallet.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *walletRepository) GetByID(ctx context.Context, walletID string) (*model.Wallet, error) {
	var wallet model.Wallet
	if err := r.db.WithContext(ctx).Where("wallet_id = ?", walletID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	var wallet model.Wallet
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &wallet, nil
}

func lockWallet(tx *gorm.DB, walletID string) (*model.Wallet, error) {
	var wallet model.Wallet
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("wallet_id = ?", walletID).First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, port.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func lockWalletPair(tx *gorm.DB, aID, bID string) (a *model.Wallet, b *model.Wallet, err error) {
	first, second := aID, bID
	aFirst := true
	if second < first {
		first, second = second, first
		aFirst = false
	}
	w1, err := lockWallet(tx, first)
	if err != nil {
		return nil, nil, err
	}
	w2, err := lockWallet(tx, second)
	if err != nil {
		return nil, nil, err
	}
	if aFirst {
		return w1, w2, nil
	}
	return w2, w1, nil
}

func updateBalance(tx *gorm.DB, walletID string, balance int64) error {
	return tx.Model(&model.Wallet{}).Where("wallet_id = ?", walletID).
		Updates(map[string]interface{}{"balance": balance, "updated_at": time.Now()}).Error
}

func (r *walletRepository) Deposit(ctx context.Context, walletID string, txn *model.Transaction) (*model.Wallet, error) {
	var wallet *model.Wallet
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		w, err := lockWallet(tx, walletID)
		if err != nil {
			return err
		}
		if w.Status != model.WalletActive {
			return port.ErrWalletNotActive
		}

		w.Balance += txn.Amount
		if err := updateBalance(tx, walletID, w.Balance); err != nil {
			return err
		}
		txn.CreatedAt = time.Now()
		if err := tx.Create(txn).Error; err != nil {
			return err
		}
		wallet = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (r *walletRepository) Withdraw(ctx context.Context, walletID string, txn *model.Transaction) (*model.Wallet, error) {
	var wallet *model.Wallet
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		w, err := lockWallet(tx, walletID)
		if err != nil {
			return err
		}
		if w.Status != model.WalletActive {
			return port.ErrWalletNotActive
		}
		if w.Balance < txn.Amount {
			return port.ErrInsufficientBalance
		}

		w.Balance -= txn.Amount
		if err := updateBalance(tx, walletID, w.Balance); err != nil {
			return err
		}
		txn.CreatedAt = time.Now()
		if err := tx.Create(txn).Error; err != nil {
			return err
		}
		wallet = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (r *walletRepository) Transfer(ctx context.Context, senderWalletID, receiverWalletID string, transfer *model.Transfer, outTxn, inTxn *model.Transaction) (*model.Wallet, *model.Wallet, error) {
	var sender, receiver *model.Wallet
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		s, rcv, err := lockWalletPair(tx, senderWalletID, receiverWalletID)
		if err != nil {
			return err
		}

		if s.Status != model.WalletActive || rcv.Status != model.WalletActive {
			return port.ErrWalletNotActive
		}
		if s.Balance < transfer.Amount {
			return port.ErrInsufficientBalance
		}

		s.Balance -= transfer.Amount
		rcv.Balance += transfer.Amount
		if err := updateBalance(tx, s.WalletID, s.Balance); err != nil {
			return err
		}
		if err := updateBalance(tx, rcv.WalletID, rcv.Balance); err != nil {
			return err
		}

		now := time.Now()
		outTxn.CreatedAt, inTxn.CreatedAt = now, now
		if err := tx.Create(outTxn).Error; err != nil {
			return err
		}
		if err := tx.Create(inTxn).Error; err != nil {
			return err
		}
		transfer.CreatedAt = now
		if err := tx.Create(transfer).Error; err != nil {
			return err
		}
		sender, receiver = s, rcv
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return sender, receiver, nil
}

func (r *walletRepository) Reverse(ctx context.Context, walletID string, original *model.Transaction, reversal *model.Transaction) (*model.Wallet, error) {
	var wallet *model.Wallet
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		w, err := lockWallet(tx, walletID)
		if err != nil {
			return err
		}

		switch original.Type {
		case model.TxnDeposit:
			if w.Balance < original.Amount {
				return port.ErrInsufficientBalance
			}
			w.Balance -= original.Amount
		case model.TxnWithdraw:
			w.Balance += original.Amount
		default:
			return port.ErrTransactionNotReversible
		}
		if err := updateBalance(tx, walletID, w.Balance); err != nil {
			return err
		}

		res := tx.Model(&model.Transaction{}).
			Where("transaction_id = ? AND status = ?", original.TransactionID, model.TxnCompleted).
			Update("status", model.TxnReversed)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return port.ErrTransactionNotReversible
		}

		reversal.CreatedAt = time.Now()
		if err := tx.Create(reversal).Error; err != nil {
			return err
		}
		wallet = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (r *walletRepository) ReverseTransfer(ctx context.Context, transfer *model.Transfer, reversalToSender, reversalFromReceiver *model.Transaction) (*model.Wallet, *model.Wallet, error) {
	var sender, receiver *model.Wallet
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		s, rcv, err := lockWalletPair(tx, transfer.SenderWalletID, transfer.ReceiverWalletID)
		if err != nil {
			return err
		}
		if rcv.Balance < transfer.Amount {
			return port.ErrInsufficientBalance
		}

		s.Balance += transfer.Amount
		rcv.Balance -= transfer.Amount
		if err := updateBalance(tx, s.WalletID, s.Balance); err != nil {
			return err
		}
		if err := updateBalance(tx, rcv.WalletID, rcv.Balance); err != nil {
			return err
		}

		res := tx.Model(&model.Transfer{}).
			Where("transfer_id = ? AND status = ?", transfer.TransferID, model.TxnCompleted).
			Update("status", model.TxnReversed)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return port.ErrTransferNotReversible
		}

		if err := tx.Model(&model.Transaction{}).
			Where("transaction_id IN ? AND status = ?", []string{transfer.OutTransactionID, transfer.InTransactionID}, model.TxnCompleted).
			Update("status", model.TxnReversed).Error; err != nil {
			return err
		}

		now := time.Now()
		reversalToSender.CreatedAt, reversalFromReceiver.CreatedAt = now, now
		if err := tx.Create(reversalToSender).Error; err != nil {
			return err
		}
		if err := tx.Create(reversalFromReceiver).Error; err != nil {
			return err
		}
		sender, receiver = s, rcv
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return sender, receiver, nil
}

func (r *walletRepository) UpdateStatus(ctx context.Context, walletID string, status model.WalletStatus, reason *string) (*model.Wallet, error) {
	var wallet *model.Wallet
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		w, err := lockWallet(tx, walletID)
		if err != nil {
			return err
		}
		if err := tx.Model(&model.Wallet{}).Where("wallet_id = ?", walletID).
			Updates(map[string]interface{}{"status": status, "frozen_reason": reason, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		w.Status = status
		w.FrozenReason = reason
		wallet = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (r *walletRepository) Close(ctx context.Context, walletID string) (*model.Wallet, error) {
	var wallet *model.Wallet
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		w, err := lockWallet(tx, walletID)
		if err != nil {
			return err
		}
		if w.Balance != 0 {
			return port.ErrWalletNotEmpty
		}
		if err := tx.Model(&model.Wallet{}).Where("wallet_id = ?", walletID).
			Updates(map[string]interface{}{"status": model.WalletClosed, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		w.Status = model.WalletClosed
		wallet = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (r *walletRepository) UpdateLimits(ctx context.Context, walletID string, dailyLimit, perTransactionLimit int64) (*model.Wallet, error) {
	res := r.db.WithContext(ctx).Model(&model.Wallet{}).Where("wallet_id = ?", walletID).
		Updates(map[string]interface{}{"daily_limit": dailyLimit, "per_transaction_limit": perTransactionLimit, "updated_at": time.Now()})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, port.ErrNotFound
	}
	return r.GetByID(ctx, walletID)
}
