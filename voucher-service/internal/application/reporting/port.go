package reporting

import (
	"context"
	"time"
)

type RedemptionReport struct {
	MerchantID     string    `json:"merchant_id"`
	TotalIssued    int       `json:"total_issued"`
	TotalRedeemed  int       `json:"total_redeemed"`
	RedemptionRate float64   `json:"redemption_rate"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
}

type CorporateSpendReport struct {
	CorporateID  string `json:"corporate_id"`
	TotalSpent   int64  `json:"total_spent"`
	Currency     string `json:"currency"`
	VouchersSent int    `json:"vouchers_sent"`
}

type ReportingService interface {
	RedemptionRateByMerchant(ctx context.Context, from, to time.Time) ([]RedemptionReport, error)
	CorporateSpend(ctx context.Context, corporateID string, from, to time.Time) (*CorporateSpendReport, error)
}

type ReportingRepository interface {
	RedemptionRateByMerchant(ctx context.Context, from, to time.Time) ([]RedemptionReport, error)
	CorporateSpend(ctx context.Context, corporateID string, from, to time.Time) (*CorporateSpendReport, error)
}
