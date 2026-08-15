package inventory

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type StockLevel struct {
	MerchantID     shared.MerchantID
	ProductSKU     string
	AvailableCount int
}

type ImportCodesInput struct {
	MerchantID shared.MerchantID
	ProductSKU string
	Codes      []CodeEntry
	BatchID    string
}

type CodeEntry struct {
	Code string
	PIN  string
}

type InventoryService interface {
	ImportCodes(ctx context.Context, in ImportCodesInput) (imported int, err error)
	GetStockLevel(ctx context.Context, merchantID shared.MerchantID, productSKU string) (StockLevel, error)
	ListLowStock(ctx context.Context, threshold int) ([]StockLevel, error)
}

type StockRepository interface {
	BulkInsert(ctx context.Context, merchantID shared.MerchantID, productSKU string, codes []CodeEntry, batchID string) (int, error)
	CountAvailable(ctx context.Context, merchantID shared.MerchantID, productSKU string) (int, error)
	ListAvailableBelowThreshold(ctx context.Context, threshold int) ([]StockLevel, error)
}

type StockClaimer interface {
	ClaimCode(ctx context.Context, merchantID shared.MerchantID, productSKU string, voucherID shared.VoucherID) (code, pin string, err error)
}
