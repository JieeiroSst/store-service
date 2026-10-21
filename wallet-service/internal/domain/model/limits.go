package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Limits struct {
	PerTransaction decimal.Decimal `json:"per_transaction"`
	Daily          decimal.Decimal `json:"daily"`
}

type LimitUsage struct {
	WalletID       uuid.UUID       `json:"wallet_id"`
	Tier           Tier            `json:"tier"`
	Currency       Currency        `json:"currency"`
	Limits         Limits          `json:"limits"`
	UsedToday      decimal.Decimal `json:"used_today"`
	RemainingToday decimal.Decimal `json:"remaining_today"`
	ResetsAt       time.Time       `json:"resets_at"`
}

var tierLimitsUSD = map[Tier]Limits{
	TierBasic:    {PerTransaction: decimal.NewFromInt(2000), Daily: decimal.NewFromInt(5000)},
	TierVerified: {PerTransaction: decimal.NewFromInt(10000), Daily: decimal.NewFromInt(25000)},
	TierPremium:  {PerTransaction: decimal.NewFromInt(50000), Daily: decimal.NewFromInt(100000)},
}

func LimitsFor(tier Tier, currency Currency) Limits {
	base, ok := tierLimitsUSD[tier]
	if !ok {
		base = tierLimitsUSD[TierBasic]
	}
	return Limits{
		PerTransaction: currency.fromUSD(base.PerTransaction),
		Daily:          currency.fromUSD(base.Daily),
	}
}

func (l Limits) Check(amount, usedToday decimal.Decimal) error {
	if amount.GreaterThan(l.PerTransaction) {
		return ErrPerTransactionLimit
	}
	if usedToday.Add(amount).GreaterThan(l.Daily) {
		return ErrDailyLimit
	}
	return nil
}

func StartOfDay(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

func NewLimitUsage(w *Wallet, usedToday decimal.Decimal, now time.Time) LimitUsage {
	limits := LimitsFor(w.Tier, w.Currency)
	remaining := limits.Daily.Sub(usedToday)
	if remaining.IsNegative() {
		remaining = decimal.Zero
	}
	return LimitUsage{
		WalletID:       w.ID,
		Tier:           w.Tier,
		Currency:       w.Currency,
		Limits:         limits,
		UsedToday:      usedToday,
		RemainingToday: remaining,
		ResetsAt:       StartOfDay(now).Add(24 * time.Hour),
	}
}
