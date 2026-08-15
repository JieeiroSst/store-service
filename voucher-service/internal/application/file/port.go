package file

import (
	"context"
	"io"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type ImportStockInput struct {
	MerchantID shared.MerchantID
	ProductSKU string
	Reader     io.Reader // CSV: code,pin
}

type ImportResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

type ExportReportInput struct {
	ReportType string // "redemptions" | "corporate_spend"
	From, To   string
}

type FileService interface {
	ImportStockCSV(ctx context.Context, in ImportStockInput) (*ImportResult, error)
	ExportReportCSV(ctx context.Context, in ExportReportInput, w io.Writer) error
}

type StockImporter interface {
	ImportCodes(ctx context.Context, merchantID shared.MerchantID, productSKU string, codes [][2]string, batchID string) (int, error)
}

type ReportSource interface {
	RedemptionRows(ctx context.Context, from, to string) ([][]string, error)
}
