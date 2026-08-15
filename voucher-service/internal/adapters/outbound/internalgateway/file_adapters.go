package internalgateway

import (
	"context"
	"time"

	fileapp "github.com/JIeeiroSst/voucher-service/internal/application/file"
	inventoryapp "github.com/JIeeiroSst/voucher-service/internal/application/inventory"
	reportingapp "github.com/JIeeiroSst/voucher-service/internal/application/reporting"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type StockImporterAdapter struct {
	inventorySvc inventoryapp.InventoryService
}

func NewStockImporterAdapter(inventorySvc inventoryapp.InventoryService) fileapp.StockImporter {
	return &StockImporterAdapter{inventorySvc: inventorySvc}
}

func (a *StockImporterAdapter) ImportCodes(ctx context.Context, merchantID shared.MerchantID, productSKU string, codes [][2]string, batchID string) (int, error) {
	entries := make([]inventoryapp.CodeEntry, len(codes))
	for i, c := range codes {
		entries[i] = inventoryapp.CodeEntry{Code: c[0], PIN: c[1]}
	}
	return a.inventorySvc.ImportCodes(ctx, inventoryapp.ImportCodesInput{
		MerchantID: merchantID,
		ProductSKU: productSKU,
		Codes:      entries,
		BatchID:    batchID,
	})
}

type ReportSourceAdapter struct {
	reportingSvc reportingapp.ReportingService
}

func NewReportSourceAdapter(reportingSvc reportingapp.ReportingService) fileapp.ReportSource {
	return &ReportSourceAdapter{reportingSvc: reportingSvc}
}

func (a *ReportSourceAdapter) RedemptionRows(ctx context.Context, from, to string) ([][]string, error) {
	fromT, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return nil, err
	}
	toT, err := time.Parse(time.RFC3339, to)
	if err != nil {
		return nil, err
	}
	reports, err := a.reportingSvc.RedemptionRateByMerchant(ctx, fromT, toT)
	if err != nil {
		return nil, err
	}
	rows := [][]string{{"merchant_id", "total_issued", "total_redeemed", "redemption_rate"}}
	for _, r := range reports {
		rows = append(rows, []string{
			r.MerchantID,
			itoa(r.TotalIssued),
			itoa(r.TotalRedeemed),
			ftoa(r.RedemptionRate),
		})
	}
	return rows, nil
}
