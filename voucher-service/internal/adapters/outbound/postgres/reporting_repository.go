package postgres

import (
	"context"
	"time"

	reportingapp "github.com/JIeeiroSst/voucher-service/internal/application/reporting"
	"gorm.io/gorm"
)

type ReportingRepository struct {
	db *gorm.DB
}

func NewReportingRepository(db *gorm.DB) reportingapp.ReportingRepository {
	return &ReportingRepository{db: db}
}

func (r *ReportingRepository) RedemptionRateByMerchant(ctx context.Context, from, to time.Time) ([]reportingapp.RedemptionReport, error) {
	type row struct {
		MerchantID    string
		TotalIssued   int
		TotalRedeemed int
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&voucherModel{}).
		Select("merchant_id, count(*) as total_issued, count(*) FILTER (WHERE status = 'redeemed') as total_redeemed").
		Where("created_at BETWEEN ? AND ?", from, to).
		Group("merchant_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]reportingapp.RedemptionReport, 0, len(rows))
	for _, row := range rows {
		rate := 0.0
		if row.TotalIssued > 0 {
			rate = float64(row.TotalRedeemed) / float64(row.TotalIssued)
		}
		out = append(out, reportingapp.RedemptionReport{
			MerchantID:     row.MerchantID,
			TotalIssued:    row.TotalIssued,
			TotalRedeemed:  row.TotalRedeemed,
			RedemptionRate: rate,
			PeriodStart:    from,
			PeriodEnd:      to,
		})
	}
	return out, nil
}

func (r *ReportingRepository) CorporateSpend(ctx context.Context, corporateID string, from, to time.Time) (*reportingapp.CorporateSpendReport, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Table("wallet_transactions wt").
		Joins("JOIN wallets w ON w.id = wt.wallet_id").
		Where("w.owner_type = 'corporate' AND w.owner_id = ? AND wt.type = 'debit' AND wt.created_at BETWEEN ? AND ?", corporateID, from, to).
		Select("COALESCE(SUM(wt.amount), 0)").
		Scan(&total).Error
	if err != nil {
		return nil, err
	}

	var voucherCount int64
	err = r.db.WithContext(ctx).Model(&voucherModel{}).
		Where("owner_type = 'corporate' AND owner_id = ? AND created_at BETWEEN ? AND ?", corporateID, from, to).
		Count(&voucherCount).Error
	if err != nil {
		return nil, err
	}

	return &reportingapp.CorporateSpendReport{
		CorporateID:  corporateID,
		TotalSpent:   int64(total),
		Currency:     "VND",
		VouchersSent: int(voucherCount),
	}, nil
}
