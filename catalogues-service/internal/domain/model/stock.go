package model

import (
	"errors"
	"time"
)

var ErrOutOfStock = errors.New("not enough stock")

type StockRecord struct {
	ID            int64
	ProductID     int64
	PartnerSKU    string
	PriceCurrency string
	Price         float64
	NumInStock    *int
	NumAllocated  int
	DateCreated   time.Time
	DateUpdated   time.Time
}

func (s StockRecord) Available() *int {
	if s.NumInStock == nil {
		return nil
	}
	n := *s.NumInStock - s.NumAllocated
	if n < 0 {
		n = 0
	}
	return &n
}

func (s StockRecord) CanSupply(qty int) bool {
	a := s.Available()
	return a == nil || *a >= qty
}
