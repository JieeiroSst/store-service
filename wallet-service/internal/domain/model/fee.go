package model

import "github.com/shopspring/decimal"

type FeeSchedule struct {
	CardRates        map[CardNetwork]decimal.Decimal
	DefaultCardRate  decimal.Decimal
	WithdrawalRate   decimal.Decimal
	WithdrawalMinUSD decimal.Decimal
	TransferRate     decimal.Decimal
}

func DefaultFeeSchedule() FeeSchedule {
	return FeeSchedule{
		CardRates: map[CardNetwork]decimal.Decimal{
			CardNetworkVisa:       decimal.RequireFromString("0.015"),
			CardNetworkMastercard: decimal.RequireFromString("0.016"),
			CardNetworkAmex:       decimal.RequireFromString("0.025"),
		},
		DefaultCardRate:  decimal.RequireFromString("0.02"),
		WithdrawalRate:   decimal.RequireFromString("0.005"),
		WithdrawalMinUSD: decimal.RequireFromString("0.5"),
		TransferRate:     decimal.Zero,
	}
}

func (f FeeSchedule) CardInterchange(amount decimal.Decimal, network CardNetwork, currency Currency) decimal.Decimal {
	rate, ok := f.CardRates[network]
	if !ok {
		rate = f.DefaultCardRate
	}
	return currency.Round(amount.Mul(rate))
}

func (f FeeSchedule) Withdrawal(amount decimal.Decimal, currency Currency) decimal.Decimal {
	fee := amount.Mul(f.WithdrawalRate)
	if min := currency.fromUSD(f.WithdrawalMinUSD); fee.LessThan(min) {
		fee = min
	}
	return currency.Round(fee)
}

func (f FeeSchedule) Transfer(amount decimal.Decimal, currency Currency) decimal.Decimal {
	return currency.Round(amount.Mul(f.TransferRate))
}
