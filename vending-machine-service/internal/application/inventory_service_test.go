package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
)

func newInventoryFixture() (*store, *inventoryService) {
	st := newStore()
	st.machines["m1"] = domain.Machine{ID: "m1", Status: domain.MachineActive}
	st.products["p1"] = domain.Product{ID: "p1", IsActive: true}
	svc := NewInventoryService(fakeTx{}, machineRepo{st}, productRepo{st}, inventoryRepo{st}, eventRepo{st}).(*inventoryService)
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
	return st, svc
}

func TestAssignSlotCreatesThenUpdates(t *testing.T) {
	st, svc := newInventoryFixture()
	in := port.SlotInput{ProductID: "p1", Quantity: 5, MaxCapacity: 10, LowThreshold: 2}
	first, err := svc.AssignSlot(context.Background(), "m1", "A1", in)
	if err != nil {
		t.Fatal(err)
	}
	in.Quantity = 8
	second, err := svc.AssignSlot(context.Background(), "m1", "A1", in)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || len(st.inventory) != 1 || st.inventory[first.ID].Quantity != 8 {
		t.Fatalf("slot was not updated in place: %+v", st.inventory)
	}
}

func TestAssignSlotValidates(t *testing.T) {
	_, svc := newInventoryFixture()
	_, err := svc.AssignSlot(context.Background(), "m1", "A1", port.SlotInput{ProductID: "p1", Quantity: 11, MaxCapacity: 10})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
	_, err = svc.AssignSlot(context.Background(), "m1", "A1", port.SlotInput{ProductID: "missing", Quantity: 1, MaxCapacity: 10})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestRestockCapsAtCapacity(t *testing.T) {
	st, svc := newInventoryFixture()
	st.inventory["i1"] = domain.Inventory{ID: "i1", MachineID: "m1", ProductID: "p1", SlotIdentifier: "A1", Quantity: 7, MaxCapacity: 10}
	inv, err := svc.Restock(context.Background(), "i1", 5)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Quantity != 10 || inv.LastRestocked == nil {
		t.Fatalf("inventory %+v", inv)
	}
	if st.events[0].Data["added"] != 3 {
		t.Fatalf("event %+v", st.events[0])
	}
	if _, err := svc.Restock(context.Background(), "i1", 0); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

func TestRecordMaintenanceStampsMachine(t *testing.T) {
	st := newStore()
	st.machines["m1"] = domain.Machine{ID: "m1", Status: domain.MachineActive}
	svc := NewMachineService(fakeTx{}, machineRepo{st}, maintenanceRepo{st}, eventRepo{st}, fakeReports{})
	l, err := svc.RecordMaintenance(context.Background(), &domain.MaintenanceLog{MachineID: "m1", MaintenanceType: "cleaning"})
	if err != nil {
		t.Fatal(err)
	}
	if m := st.machines["m1"]; m.LastMaintenance == nil || !m.LastMaintenance.Equal(l.PerformedAt) {
		t.Fatalf("machine %+v", m)
	}
}

func TestSalesReportTotals(t *testing.T) {
	st := newStore()
	st.machines["m1"] = domain.Machine{ID: "m1"}
	svc := NewMachineService(fakeTx{}, machineRepo{st}, maintenanceRepo{st}, eventRepo{st}, fakeReports{sales: []domain.ProductSales{
		{ProductID: "a", Units: 3, RevenueCents: 300, DiscountCents: 20},
		{ProductID: "b", Units: 1, RevenueCents: 150},
	}})
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r, err := svc.SalesReport(context.Background(), "m1", from, from.AddDate(0, 1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if r.Orders != 4 || r.RevenueCents != 450 || r.DiscountCents != 20 {
		t.Fatalf("report %+v", r)
	}
	if _, err := svc.SalesReport(context.Background(), "m1", from, from); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("empty range: %v", err)
	}
	if _, err := svc.SalesReport(context.Background(), "m1", from, from.AddDate(2, 0, 0)); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("huge range: %v", err)
	}
}

func TestUpdateProductValidates(t *testing.T) {
	st := newStore()
	st.products["p1"] = domain.Product{ID: "p1", Name: "Cola", CategoryID: "c", PriceCents: 100, IsActive: true}
	svc := NewCatalogService(nil, productRepo{st})
	price, off := 120, false
	p, err := svc.UpdateProduct(context.Background(), "p1", domain.ProductPatch{PriceCents: &price, IsActive: &off})
	if err != nil || p.PriceCents != 120 || p.IsActive || st.products["p1"].PriceCents != 120 {
		t.Fatalf("err=%v product=%+v", err, p)
	}
	neg := -1
	if _, err := svc.UpdateProduct(context.Background(), "p1", domain.ProductPatch{PriceCents: &neg}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}
