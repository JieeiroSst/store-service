package mysql

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPaidOrderStoreIsIdempotentPerBasket(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewPaidOrderStore(db)
	if err != nil {
		t.Fatal(err)
	}
	in := entity.PaidOrderInput{
		BasketID: 12, UserID: 7, Currency: "VND", Total: 300000, PaymentReference: "tx-1", ShippingAddressID: 4,
		Lines: []entity.PaidOrderLine{{ProductID: 3, StockrecordID: 9, Title: "Áo", Quantity: 2, UnitPrice: 150000}},
	}
	ctx := context.Background()

	first, err := s.CreatePaidOrder(ctx, in)
	if err != nil || !first.Created || first.Number != "KS00000012" {
		t.Fatalf("first create: %+v %v", first, err)
	}
	again, err := s.CreatePaidOrder(ctx, in)
	if err != nil || again.Created || again.ID != first.ID {
		t.Fatalf("retry must return the same order: %+v %v", again, err)
	}

	var orders, lines int64
	db.Model(&orderRow{}).Count(&orders)
	db.Model(&orderLineRow{}).Count(&lines)
	if orders != 1 || lines != 1 {
		t.Fatalf("orders=%d lines=%d", orders, lines)
	}
	var row orderRow
	db.First(&row, first.ID)
	if row.Status != "Paid" || row.UserID != 7 || row.TotalInclTax != 300000 || row.PaymentReference != "tx-1" || row.BasketID != 12 {
		t.Fatalf("unexpected row: %+v", row)
	}
	var line orderLineRow
	db.First(&line)
	if line.LinePrice != 300000 || line.OrderID != first.ID {
		t.Fatalf("unexpected line: %+v", line)
	}

	if err := s.SetStatus(ctx, "KS00000012", "Processing"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetStatus(ctx, "KS99999999", "Processing"); err == nil {
		t.Fatal("unknown order must fail")
	}
}
