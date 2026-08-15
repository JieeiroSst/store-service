package inventory

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Service struct {
	repo StockRepository
}

func NewService(repo StockRepository) InventoryService {
	return &Service{repo: repo}
}

func (s *Service) ImportCodes(ctx context.Context, in ImportCodesInput) (int, error) {
	batchID := in.BatchID
	if batchID == "" {
		batchID = shared.NewVoucherID().String()
	}
	return s.repo.BulkInsert(ctx, in.MerchantID, in.ProductSKU, in.Codes, batchID)
}

func (s *Service) GetStockLevel(ctx context.Context, merchantID shared.MerchantID, productSKU string) (StockLevel, error) {
	count, err := s.repo.CountAvailable(ctx, merchantID, productSKU)
	if err != nil {
		return StockLevel{}, err
	}
	return StockLevel{MerchantID: merchantID, ProductSKU: productSKU, AvailableCount: count}, nil
}

func (s *Service) ListLowStock(ctx context.Context, threshold int) ([]StockLevel, error) {
	return s.repo.ListAvailableBelowThreshold(ctx, threshold)
}
