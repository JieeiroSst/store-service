package shared

import "fmt"

type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func NewMoney(amount int64, currency string) Money {
	return Money{Amount: amount, Currency: currency}
}

func ZeroMoney(currency string) Money {
	return Money{Amount: 0, Currency: currency}
}

func (m Money) IsZero() bool     { return m.Amount == 0 }
func (m Money) IsNegative() bool { return m.Amount < 0 }
func (m Money) IsPositive() bool { return m.Amount > 0 }

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

func (m Money) GreaterThanOrEqual(other Money) bool {
	return m.Currency == other.Currency && m.Amount >= other.Amount
}

func (m Money) LessThan(other Money) bool {
	return m.Currency == other.Currency && m.Amount < other.Amount
}

func (m Money) Equal(other Money) bool {
	return m.Currency == other.Currency && m.Amount == other.Amount
}

func (m Money) String() string {
	return fmt.Sprintf("%d %s", m.Amount, m.Currency)
}
