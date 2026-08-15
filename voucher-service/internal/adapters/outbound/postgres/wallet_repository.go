package postgres

import (
	"context"
	"time"

	walletapp "github.com/JIeeiroSst/voucher-service/internal/application/wallet"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/wallet"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"gorm.io/gorm"
)

type walletModel struct {
	ID        string    `gorm:"column:id;primaryKey"`
	OwnerType string    `gorm:"column:owner_type"`
	OwnerID   string    `gorm:"column:owner_id"`
	Balance   float64   `gorm:"column:balance"`
	Currency  string    `gorm:"column:currency"`
	Version   int       `gorm:"column:version"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (walletModel) TableName() string { return "wallets" }

func walletToModel(w *wallet.Wallet) *walletModel {
	return &walletModel{
		ID:        w.ID.String(),
		OwnerType: string(w.OwnerType),
		OwnerID:   w.OwnerID,
		Balance:   float64(w.Balance.Amount),
		Currency:  w.Balance.Currency,
		Version:   w.Version,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func walletFromModel(m *walletModel) (*wallet.Wallet, error) {
	id, err := shared.ParseWalletID(m.ID)
	if err != nil {
		return nil, err
	}
	return &wallet.Wallet{
		ID:               id,
		OwnerType:        wallet.OwnerType(m.OwnerType),
		OwnerID:          m.OwnerID,
		Balance:          shared.NewMoney(int64(m.Balance), m.Currency),
		Version:          m.Version,
		PersistedVersion: m.Version,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}, nil
}

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) walletapp.WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) find(ctx context.Context, ownerType wallet.OwnerType, ownerID string, forUpdate bool) (*wallet.Wallet, error) {
	db := txmanager.DBFromContext(ctx, r.db)
	if forUpdate {
		db = lockForUpdate(db)
	}
	var m walletModel
	err := db.First(&m, "owner_type = ? AND owner_id = ?", string(ownerType), ownerID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, wallet.ErrWalletNotFound
		}
		return nil, err
	}
	return walletFromModel(&m)
}

func (r *WalletRepository) FindByOwner(ctx context.Context, ownerType wallet.OwnerType, ownerID string) (*wallet.Wallet, error) {
	return r.find(ctx, ownerType, ownerID, false)
}

func (r *WalletRepository) FindByOwnerForUpdate(ctx context.Context, ownerType wallet.OwnerType, ownerID string) (*wallet.Wallet, error) {
	return r.find(ctx, ownerType, ownerID, true)
}

func (r *WalletRepository) Create(ctx context.Context, w *wallet.Wallet) error {
	return txmanager.DBFromContext(ctx, r.db).Create(walletToModel(w)).Error
}

func (r *WalletRepository) Save(ctx context.Context, w *wallet.Wallet) error {
	model := walletToModel(w)
	tx := txmanager.DBFromContext(ctx, r.db).Model(&walletModel{}).
		Where("id = ? AND version = ?", model.ID, w.PersistedVersion).
		Updates(map[string]any{
			"balance":    model.Balance,
			"version":    model.Version,
			"updated_at": model.UpdatedAt,
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return wallet.ErrVersionConflict
	}
	return nil
}

// ---- Ledger (append-only) ----

type walletTransactionModel struct {
	ID             string    `gorm:"column:id;primaryKey"`
	WalletID       string    `gorm:"column:wallet_id"`
	Type           string    `gorm:"column:type"`
	Amount         float64   `gorm:"column:amount"`
	BalanceAfter   float64   `gorm:"column:balance_after"`
	RefType        string    `gorm:"column:ref_type"`
	RefID          string    `gorm:"column:ref_id"`
	IdempotencyKey *string   `gorm:"column:idempotency_key"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (walletTransactionModel) TableName() string { return "wallet_transactions" }

type LedgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) walletapp.LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) Append(ctx context.Context, entry walletapp.LedgerEntry) error {
	model := walletTransactionModel{
		ID:           shared.NewWalletID().String(),
		WalletID:     entry.WalletID.String(),
		Type:         entry.Type,
		Amount:       float64(entry.Amount.Amount),
		BalanceAfter: float64(entry.BalanceAfter.Amount),
		RefType:      entry.RefType,
		RefID:        entry.RefID,
		CreatedAt:    time.Now().UTC(),
	}
	if entry.IdempotencyKey != "" {
		model.IdempotencyKey = &entry.IdempotencyKey
	}
	return txmanager.DBFromContext(ctx, r.db).Create(&model).Error
}
