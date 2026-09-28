package model

import (
	"testing"
	"time"
)

func TestNormalizeVNPhone(t *testing.T) {
	ok := map[string]string{
		"0912 345 678":   "0912345678",
		"+84912345678":   "0912345678",
		"84-912.345.678": "0912345678",
		"02838123456":    "02838123456",
	}
	for in, want := range ok {
		if got, err := NormalizeVNPhone(in); err != nil || got != want {
			t.Errorf("%q -> %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "12345", "+14155550100", "0112345678", "091234567"} {
		if _, err := NormalizeVNPhone(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestRankRoutes(t *testing.T) {
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	options := []RouteOption{
		{WarehouseCode: "HN", ServiceID: 1, Fee: 20000, ExpectedDeliveryAt: base.Add(72 * time.Hour)},
		{WarehouseCode: "HCM", ServiceID: 2, Fee: 45000, ExpectedDeliveryAt: base.Add(24 * time.Hour)},
		{WarehouseCode: "DN", ServiceID: 3, Fee: 30000, ExpectedDeliveryAt: base.Add(36 * time.Hour)},
	}
	if got := RankRoutes(options, StrategyCheapest)[0].WarehouseCode; got != "HN" {
		t.Errorf("cheapest = %s", got)
	}
	if got := RankRoutes(options, StrategyFastest)[0].WarehouseCode; got != "HCM" {
		t.Errorf("fastest = %s", got)
	}
	if got := RankRoutes(options, StrategyBalanced)[0].WarehouseCode; got != "DN" {
		t.Errorf("balanced = %s", got)
	}
	if options[0].Score != 0 || len(RankRoutes(nil, StrategyBalanced)) != 0 {
		t.Error("input mutated or empty input mishandled")
	}
}

func TestRankRoutesTieBreaks(t *testing.T) {
	eta := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	options := []RouteOption{
		{WarehouseCode: "A", Fee: 30000, ExpectedDeliveryAt: eta},
		{WarehouseCode: "B", Fee: 25000, ExpectedDeliveryAt: eta},
	}
	if got := RankRoutes(options, StrategyFastest)[0].WarehouseCode; got != "B" {
		t.Errorf("same ETA should prefer cheaper, got %s", got)
	}
}

func TestLifecycleTransitions(t *testing.T) {
	allowed := [][2]Status{
		{StatusPending, StatusCreated},
		{StatusCreated, StatusPicking},
		{StatusPicking, StatusPicked},
		{StatusPicked, StatusInTransit},
		{StatusInTransit, StatusDelivering},
		{StatusDelivering, StatusDeliveryFailed},
		{StatusDeliveryFailed, StatusDelivering},
		{StatusDelivering, StatusDelivered},
		{StatusDeliveryFailed, StatusReturning},
		{StatusReturning, StatusReturned},
		{StatusCreated, StatusCancelled},
	}
	for _, tr := range allowed {
		if !CanTransition(tr[0], tr[1]) {
			t.Errorf("%s -> %s should be allowed", tr[0], tr[1])
		}
	}
	denied := [][2]Status{
		{StatusDelivering, StatusPicked},
		{StatusDelivered, StatusDelivering},
		{StatusPicked, StatusCancelled},
		{StatusCancelled, StatusCreated},
		{StatusReturned, StatusDelivered},
	}
	for _, tr := range denied {
		if CanTransition(tr[0], tr[1]) {
			t.Errorf("%s -> %s should be denied", tr[0], tr[1])
		}
	}
	for _, s := range []Status{StatusDelivered, StatusReturned, StatusCancelled} {
		if !s.Terminal() || len(transitions[s]) != 0 {
			t.Errorf("%s must be terminal", s)
		}
	}
}

func TestMapGHNStatus(t *testing.T) {
	cases := map[string]Status{
		"ready_to_pick": StatusCreated, "money_collect_picking": StatusPicking, "sorting": StatusInTransit,
		"delivery_fail": StatusDeliveryFailed, "return_transporting": StatusReturning, "lost": StatusException,
		"cancel": StatusCancelled, "delivered": StatusDelivered,
	}
	for in, want := range cases {
		if got, ok := MapGHNStatus(in); !ok || got != want {
			t.Errorf("%s -> %s, want %s", in, got, want)
		}
	}
	if _, ok := MapGHNStatus("something_new"); ok {
		t.Error("unknown status mapped")
	}
}

func TestCustomerNotifications(t *testing.T) {
	s := Shipment{Code: "SHP1", ClientOrderCode: "DH-9", Status: StatusDelivering, Customer: Customer{UserID: 7, Email: "a@b.vn"}}
	got := CustomerNotifications(s)
	if len(got) != 2 || got[0].Channel != "email" || got[1].Channel != "push" || got[1].Recipient != "7" {
		t.Fatalf("got %+v", got)
	}
	if got[0].Message != "Đơn hàng DH-9 đang được giao đến bạn." {
		t.Errorf("message = %q", got[0].Message)
	}
	s.Status = StatusInTransit
	if len(CustomerNotifications(s)) != 0 {
		t.Error("in-transit hops should not notify")
	}
}
