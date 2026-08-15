package voucher

import (
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type StockStatus string

const (
	StockStatusAvailable StockStatus = "available"
	StockStatusClaimed   StockStatus = "claimed"
	StockStatusVoid      StockStatus = "void"
)

type StockCode struct {
	ID                 shared.VoucherID
	MerchantID         shared.MerchantID
	ProductSKU         string
	Code               string
	PIN                string
	Status             StockStatus
	ClaimedByVoucherID *shared.VoucherID
	BatchID            string
	ImportedAt         time.Time
	ClaimedAt          *time.Time
}

var ErrStockCodeAlreadyClaimed = errStockAlreadyClaimed{}

type errStockAlreadyClaimed struct{}

func (errStockAlreadyClaimed) Error() string { return "stock code already claimed" }

func (s *StockCode) Claim(voucherID shared.VoucherID, now time.Time) error {
	if s.Status != StockStatusAvailable {
		return ErrStockCodeAlreadyClaimed
	}
	s.Status = StockStatusClaimed
	s.ClaimedByVoucherID = &voucherID
	s.ClaimedAt = &now
	return nil
}
