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

type pocketRepository struct {
	db *gorm.DB
}

func NewPocketRepository(db *gorm.DB) port.PocketRepository {
	return &pocketRepository{db: db}
}

func (r *pocketRepository) Create(ctx context.Context, pocket *model.Pocket) error {
	now := time.Now()
	pocket.CreatedAt, pocket.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(pocket).Error
}

func (r *pocketRepository) GetByID(ctx context.Context, pocketID string) (*model.Pocket, error) {
	var pocket model.Pocket
	if err := r.db.WithContext(ctx).Where("pocket_id = ?", pocketID).First(&pocket).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &pocket, nil
}

func (r *pocketRepository) ListByWallet(ctx context.Context, walletID string) ([]model.Pocket, error) {
	var pockets []model.Pocket
	err := r.db.WithContext(ctx).Where("wallet_id = ?", walletID).Order("created_at ASC").Find(&pockets).Error
	return pockets, err
}

func lockPocket(tx *gorm.DB, pocketID string) (*model.Pocket, error) {
	var pocket model.Pocket
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("pocket_id = ?", pocketID).First(&pocket).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, port.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &pocket, nil
}

func updatePocketBalance(tx *gorm.DB, pocketID string, balance int64) error {
	return tx.Model(&model.Pocket{}).Where("pocket_id = ?", pocketID).
		Updates(map[string]interface{}{"balance": balance, "updated_at": time.Now()}).Error
}

func (r *pocketRepository) MoveToPocket(ctx context.Context, walletID, pocketID string, txn *model.Transaction) (*model.Wallet, *model.Pocket, error) {
	var wallet *model.Wallet
	var pocket *model.Pocket
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
		p, err := lockPocket(tx, pocketID)
		if err != nil {
			return err
		}
		if p.WalletID != walletID {
			return port.ErrNotFound
		}

		w.Balance -= txn.Amount
		p.Balance += txn.Amount
		if err := updateBalance(tx, walletID, w.Balance); err != nil {
			return err
		}
		if err := updatePocketBalance(tx, pocketID, p.Balance); err != nil {
			return err
		}
		txn.CreatedAt = time.Now()
		if err := tx.Create(txn).Error; err != nil {
			return err
		}
		wallet, pocket = w, p
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return wallet, pocket, nil
}

func (r *pocketRepository) MoveFromPocket(ctx context.Context, walletID, pocketID string, txn *model.Transaction) (*model.Wallet, *model.Pocket, error) {
	var wallet *model.Wallet
	var pocket *model.Pocket
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		w, err := lockWallet(tx, walletID)
		if err != nil {
			return err
		}
		p, err := lockPocket(tx, pocketID)
		if err != nil {
			return err
		}
		if p.WalletID != walletID {
			return port.ErrNotFound
		}
		if p.Balance < txn.Amount {
			return port.ErrInsufficientBalance
		}

		w.Balance += txn.Amount
		p.Balance -= txn.Amount
		if err := updateBalance(tx, walletID, w.Balance); err != nil {
			return err
		}
		if err := updatePocketBalance(tx, pocketID, p.Balance); err != nil {
			return err
		}
		txn.CreatedAt = time.Now()
		if err := tx.Create(txn).Error; err != nil {
			return err
		}
		wallet, pocket = w, p
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return wallet, pocket, nil
}

func (r *pocketRepository) Close(ctx context.Context, walletID, pocketID string) (*model.Wallet, error) {
	var wallet *model.Wallet
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		w, err := lockWallet(tx, walletID)
		if err != nil {
			return err
		}
		p, err := lockPocket(tx, pocketID)
		if err != nil {
			return err
		}
		if p.WalletID != walletID {
			return port.ErrNotFound
		}

		w.Balance += p.Balance
		if err := updateBalance(tx, walletID, w.Balance); err != nil {
			return err
		}
		if err := tx.Where("pocket_id = ?", pocketID).Delete(&model.Pocket{}).Error; err != nil {
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
