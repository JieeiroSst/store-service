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

func TestPaidOrderStoreRecordsShipment(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:ship?mode=memory&cache=shared"), &gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewPaidOrderStore(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, id := range []int{20, 21} {
		in := entity.PaidOrderInput{BasketID: id, UserID: 7, Currency: "VND", Total: 1, PaymentReference: "tx", Lines: []entity.PaidOrderLine{{ProductID: 1, Quantity: 1, UnitPrice: 1}}}
		if _, err := s.CreatePaidOrder(ctx, in); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.SetShipment(ctx, "KS00000020", entity.ShipmentResult{Outcome: entity.ShipmentOutcomeCreated, Code: "SHP-1", Fee: 32500}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetShipment(ctx, "KS00000021", entity.ShipmentResult{Outcome: entity.ShipmentOutcomeSkipped}); err != nil {
		t.Fatal(err)
	}
	var created, skipped orderRow
	db.Where("number = ?", "KS00000020").First(&created)
	db.Where("number = ?", "KS00000021").First(&skipped)
	if created.ShippingCode != "SHP-1" || created.ShippingMethod != "GHN" || created.Status != StatusAwaitingPickup || created.CarrierFee != 32500 {
		t.Fatalf("created: %+v", created)
	}
	if skipped.ShippingCode != "" || skipped.Status != StatusNeedsShipping {
		t.Fatalf("skipped: %+v", skipped)
	}
	if err := s.SetShipment(ctx, "KS99999999", entity.ShipmentResult{}); err == nil {
		t.Fatal("unknown order accepted")
	}
}

func TestPaidOrderStoreKeepsTheShippingFee(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:fee?mode=memory&cache=shared"), &gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewPaidOrderStore(db)
	if err != nil {
		t.Fatal(err)
	}
	in := entity.PaidOrderInput{BasketID: 30, UserID: 7, Currency: "VND", Total: 332000, ShippingFee: 32000, PaymentReference: "tx", Lines: []entity.PaidOrderLine{{ProductID: 1, Quantity: 2, UnitPrice: 150000}}}
	if _, err := s.CreatePaidOrder(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	var row orderRow
	db.Where("number = ?", "KS00000030").First(&row)
	if row.TotalInclTax != 332000 || row.ShippingInclTax != 32000 {
		t.Fatalf("row %+v", row)
	}
	bad := in
	bad.ShippingFee = 400000
	if bad.Validate() == nil {
		t.Fatal("a shipping fee above the total must be rejected")
	}
}
