package postgres

import (
	"context"
	"errors"
	"time"

	inventoryapp "github.com/JIeeiroSst/voucher-service/internal/application/inventory"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type voucherStockModel struct {
	ID                 string     `gorm:"column:id;primaryKey"`
	MerchantID         string     `gorm:"column:merchant_id"`
	ProductSKU         string     `gorm:"column:product_sku"`
	Code               string     `gorm:"column:code"`
	PIN                string     `gorm:"column:pin"`
	Status             string     `gorm:"column:status"`
	ClaimedByVoucherID *string    `gorm:"column:claimed_by_voucher_id"`
	BatchID            *string    `gorm:"column:batch_id"`
	ImportedAt         time.Time  `gorm:"column:imported_at"`
	ClaimedAt          *time.Time `gorm:"column:claimed_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
}

func (voucherStockModel) TableName() string { return "voucher_stock" }

var ErrNoStockAvailable = errors.New("no stock codes available for this merchant/sku")

type VoucherStockRepository struct {
	db *gorm.DB
}

func NewVoucherStockRepository(db *gorm.DB) *VoucherStockRepository {
	return &VoucherStockRepository{db: db}
}

func (r *VoucherStockRepository) BulkInsert(ctx context.Context, merchantID shared.MerchantID, productSKU string, codes []inventoryapp.CodeEntry, batchID string) (int, error) {
	if len(codes) == 0 {
		return 0, nil
	}
	now := time.Now().UTC()
	models := make([]voucherStockModel, 0, len(codes))
	for _, c := range codes {
		models = append(models, voucherStockModel{
			ID:         shared.NewVoucherID().String(),
			MerchantID: merchantID.String(),
			ProductSKU: productSKU,
			Code:       c.Code,
			PIN:        c.PIN,
			Status:     "available",
			BatchID:    &batchID,
			ImportedAt: now,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&models)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return int(tx.RowsAffected), nil
}

func (r *VoucherStockRepository) CountAvailable(ctx context.Context, merchantID shared.MerchantID, productSKU string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&voucherStockModel{}).
		Where("merchant_id = ? AND product_sku = ? AND status = 'available'", merchantID.String(), productSKU).
		Count(&count).Error
	return int(count), err
}

func (r *VoucherStockRepository) ListAvailableBelowThreshold(ctx context.Context, threshold int) ([]inventoryapp.StockLevel, error) {
	type row struct {
		MerchantID string
		ProductSKU string
		Count      int
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&voucherStockModel{}).
		Select("merchant_id, product_sku, count(*) as count").
		Where("status = 'available'").
		Group("merchant_id, product_sku").
		Having("count(*) < ?", threshold).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]inventoryapp.StockLevel, 0, len(rows))
	for _, r := range rows {
		merchantID, err := shared.ParseMerchantID(r.MerchantID)
		if err != nil {
			return nil, err
		}
		out = append(out, inventoryapp.StockLevel{MerchantID: merchantID, ProductSKU: r.ProductSKU, AvailableCount: r.Count})
	}
	return out, nil
}

// ClaimCode atomically claims one available code for a voucher, using
// SELECT ... FOR UPDATE SKIP LOCKED so concurrent claims for the same
// merchant/sku don't serialize on each other. This method has no interface
// in this package; adapters/outbound/provider/stock_provider.go declares
// its own narrow "stockStore" interface that this type satisfies
// structurally (dependency inversion applied one level below the port).
func (r *VoucherStockRepository) ClaimCode(ctx context.Context, merchantID shared.MerchantID, productSKU string, voucherID shared.VoucherID) (code, pin string, err error) {
	txErr := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m voucherStockModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("merchant_id = ? AND product_sku = ? AND status = 'available'", merchantID.String(), productSKU).
			Order("created_at").
			Limit(1).
			First(&m).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrNoStockAvailable
			}
			return err
		}

		now := time.Now().UTC()
		vid := voucherID.String()
		res := tx.Model(&voucherStockModel{}).
			Where("id = ? AND status = 'available'", m.ID).
			Updates(map[string]any{
				"status":                "claimed",
				"claimed_by_voucher_id": vid,
				"claimed_at":            now,
				"updated_at":            now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNoStockAvailable
		}

		code = m.Code
		pin = m.PIN
		return nil
	})
	return code, pin, txErr
}
