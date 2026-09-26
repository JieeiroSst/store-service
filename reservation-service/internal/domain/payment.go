package domain

import (
	"fmt"
	"math/big"
	"strings"
	"time"
)

type PaymentMethod string

const (
	MethodWallet  PaymentMethod = "wallet"  // internal e-wallet (payment-wallet-service)
	MethodGateway PaymentMethod = "gateway" // external provider (payment_service)
)

type Wallet struct {
	ID       string
	UserID   int64
	Balance  int64 // minor units
	Currency string
	Status   string
}

type WalletTxn struct {
	ID          string
	Type        string
	Amount      int64
	Currency    string
	Status      string
	ReferenceID string
	Description string
	CreatedAt   time.Time
}

type GatewayStatus string

const (
	GatewayPending    GatewayStatus = "pending"
	GatewayAuthorized GatewayStatus = "authorized"
	GatewayCaptured   GatewayStatus = "captured"
	GatewayFailed     GatewayStatus = "failed"
	GatewayRefunded   GatewayStatus = "refunded"
	GatewayPartially  GatewayStatus = "partially_refunded"
)

type GatewayPayment struct {
	ID       int64
	Provider string
	Status   GatewayStatus
	Amount   int64
	Currency string
}

var zeroDecimal = map[string]bool{"VND": true, "JPY": true, "KRW": true, "CLP": true, "IDR": true}

func MinorUnits(amount, currency string) (int64, error) {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(amount))
	if !ok || r.Sign() < 0 {
		return 0, fmt.Errorf("%w: bad amount %q", ErrInvalid, amount)
	}
	if !zeroDecimal[strings.ToUpper(currency)] {
		r.Mul(r, big.NewRat(100, 1))
	}
	r.Add(r, big.NewRat(1, 2))
	n := new(big.Int).Quo(r.Num(), r.Denom())
	if !n.IsInt64() {
		return 0, fmt.Errorf("%w: amount too large", ErrInvalid)
	}
	return n.Int64(), nil
}

// FromMinorUnits is the inverse of MinorUnits: 150000 VND -> "150000.0000", 1999 USD -> "19.9900".
func FromMinorUnits(minor int64, currency string) string {
	r := big.NewRat(minor, 1)
	if !zeroDecimal[strings.ToUpper(currency)] {
		r.Quo(r, big.NewRat(100, 1))
	}
	return r.FloatString(4)
}
