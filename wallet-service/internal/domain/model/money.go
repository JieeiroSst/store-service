package model

import "github.com/shopspring/decimal"

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyVND Currency = "VND"
)

var currencyPrecision = map[Currency]int32{
	CurrencyUSD: 2,
	CurrencyEUR: 2,
	CurrencyVND: 0,
}

var currencyScale = map[Currency]int64{
	CurrencyUSD: 1,
	CurrencyEUR: 1,
	CurrencyVND: 25000,
}

func (c Currency) Valid() bool {
	_, ok := currencyPrecision[c]
	return ok
}

func (c Currency) Precision() int32 { return currencyPrecision[c] }

func (c Currency) fromUSD(amount decimal.Decimal) decimal.Decimal {
	scale, ok := currencyScale[c]
	if !ok {
		scale = 1
	}
	return amount.Mul(decimal.NewFromInt(scale))
}

func (c Currency) Round(amount decimal.Decimal) decimal.Decimal {
	return amount.Round(c.Precision())
}

func ValidateAmount(amount decimal.Decimal, currency Currency) error {
	if !currency.Valid() {
		return ErrUnsupportedCurrency
	}
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	if !amount.Equal(currency.Round(amount)) {
		return ErrAmountPrecision
	}
	return nil
}
