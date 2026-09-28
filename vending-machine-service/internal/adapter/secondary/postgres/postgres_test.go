package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/config"
	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/payment"
	"github.com/JIeeiroSst/vending-machine-service/internal/application"
	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	schema := "vending_test_" + time.Now().Format("150405000000")

	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		admin.Close()
	})

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	db, err := NewDB(pool)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.migrate(ctx); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	return db
}

type services struct {
	machines  port.MachineService
	catalog   port.CatalogService
	inventory port.InventoryService
	vending   port.VendingService
}

func newServices(db *DB) services {
	cfg := &config.Config{Vending: config.VendingConfig{
		SessionTTL: 5 * time.Minute, ReservationTTL: 2 * time.Minute, Currency: "USD",
	}}
	machines, products, inventory, events := NewMachineRepository(db), NewProductRepository(db), NewInventoryRepository(db), NewEventRepository(db)
	return services{
		machines:  application.NewMachineService(db, machines, NewMaintenanceRepository(db), events, NewReportRepository(db)),
		catalog:   application.NewCatalogService(NewCategoryRepository(db), products),
		inventory: application.NewInventoryService(db, machines, products, inventory, events),
		vending: application.NewVendingService(application.VendingDeps{
			Tx:           db,
			Machines:     machines,
			Inventory:    inventory,
			Sessions:     NewSessionRepository(db),
			Reservations: NewReservationRepository(db),
			Payments:     NewPaymentRepository(db),
			Orders:       NewOrderRepository(db),
			Events:       events,
			Gateway:      payment.Simulated{},
			Coupons:      coupons{},
		}, cfg),
	}
}

type coupons struct{}

func (coupons) Quote(context.Context, string, int) (int, error)  { return 20, nil }
func (coupons) Redeem(context.Context, string, int64, int) error { return nil }

func TestPurchaseFlow(t *testing.T) {
	db := newTestDB(t)
	s := newServices(db)
	ctx := context.Background()

	m, err := s.machines.Register(ctx, &domain.Machine{Location: "Lobby", Model: "VX-1"})
	if err != nil {
		t.Fatal(err)
	}
	cat, err := s.catalog.CreateCategory(ctx, &domain.Category{Name: "Drinks"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.catalog.CreateProduct(ctx, &domain.Product{
		Name: "Water", PriceCents: 120, CategoryID: cat.ID, IsActive: true,
		Attributes: map[string]string{"volume": "500ml"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := s.catalog.GetProduct(ctx, p.ID); got.Attributes["volume"] != "500ml" {
		t.Fatalf("attributes %v", got.Attributes)
	}
	inv, err := s.inventory.AssignSlot(ctx, m.ID, "A1", port.SlotInput{ProductID: p.ID, Quantity: 2, MaxCapacity: 5, LowThreshold: 1})
	if err != nil {
		t.Fatal(err)
	}

	sess, err := s.vending.StartSession(ctx, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.vending.Reserve(ctx, sess.ID, "A1")
	if err != nil {
		t.Fatal(err)
	}
	out, err := s.vending.Checkout(ctx, r.ID, port.CheckoutInput{Method: domain.PaymentCard, CouponCode: "SAVE"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Payment.AmountCents != 100 || out.Payment.DiscountCents != 20 || out.Payment.TransactionID == "" || out.Order.OrderNo == 0 {
		t.Fatalf("payment %+v order %+v", out.Payment, out.Order)
	}
	if got, _ := s.vending.GetOrder(ctx, out.Order.ID); got.OrderNo != out.Order.OrderNo {
		t.Fatalf("order_no %d, want %d", got.OrderNo, out.Order.OrderNo)
	}
	if _, err := s.vending.Checkout(ctx, r.ID, port.CheckoutInput{Method: domain.PaymentCard}); !errors.Is(err, domain.ErrReservationClosed) {
		t.Fatalf("second checkout: %v", err)
	}
	o, err := s.vending.ReportDispense(ctx, out.Order.ID, true)
	if err != nil || o.Status != domain.OrderCompleted {
		t.Fatalf("dispense: %v %+v", err, o)
	}

	low, err := s.inventory.ListLow(ctx, m.ID, domain.PageRequest{})
	if err != nil || len(low.Items) != 1 || low.Items[0].ID != inv.ID || low.Items[0].Quantity != 1 || !low.IsLastPage {
		t.Fatalf("low stock: %v %+v", err, low)
	}
	events, err := s.machines.ListEvents(ctx, m.ID, 50)
	if err != nil || len(events) == 0 {
		t.Fatalf("events: %v %d", err, len(events))
	}

	restocked, err := s.inventory.Restock(ctx, inv.ID, 10)
	if err != nil || restocked.Quantity != 5 {
		t.Fatalf("restock: %v %+v", err, restocked)
	}

	report, err := s.machines.SalesReport(ctx, m.ID, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil || report.Orders != 1 || report.RevenueCents != 100 || report.DiscountCents != 20 || report.Products[0].ProductName != "Water" {
		t.Fatalf("report: %v %+v", err, report)
	}
	if _, err := s.vending.RefundOrder(ctx, out.Order.ID, "customer complaint"); err != nil {
		t.Fatal(err)
	}
	report, _ = s.machines.SalesReport(ctx, m.ID, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if report.Orders != 0 {
		t.Fatalf("refunded order still counted: %+v", report)
	}

	price := 150
	if p, err := s.catalog.UpdateProduct(ctx, p.ID, domain.ProductPatch{PriceCents: &price}); err != nil || p.PriceCents != 150 {
		t.Fatalf("update product: %v %+v", err, p)
	}
	orders, err := s.vending.ListSessionOrders(ctx, sess.ID)
	if err != nil || len(orders) != 1 {
		t.Fatalf("session orders: %v %d", err, len(orders))
	}
}

func TestAlertOutbox(t *testing.T) {
	db := newTestDB(t)
	s := newServices(db)
	ctx := context.Background()
	m, _ := s.machines.Register(ctx, &domain.Machine{Location: "Gym", Model: "VX-3"})
	if _, err := s.machines.ChangeStatus(ctx, m.ID, domain.MachineMaintenance); err != nil {
		t.Fatal(err)
	}

	n := &notifier{}
	alerts := application.NewAlertService(db, NewEventRepository(db), n)
	sent, err := alerts.DispatchPending(ctx)
	if err != nil || sent != 1 || n.got[0] != domain.EventMachineStatusChanged {
		t.Fatalf("sent=%d err=%v got=%v", sent, err, n.got)
	}
	if sent, _ := alerts.DispatchPending(ctx); sent != 0 {
		t.Fatalf("alert resent: %d", sent)
	}
}

type notifier struct{ got []string }

func (n *notifier) Notify(_ context.Context, e domain.Event) error {
	n.got = append(n.got, e.EventType)
	return nil
}

func TestConcurrentReservationsDoNotOversell(t *testing.T) {
	db := newTestDB(t)
	s := newServices(db)
	ctx := context.Background()

	m, _ := s.machines.Register(ctx, &domain.Machine{Location: "Hall", Model: "VX-2"})
	cat, _ := s.catalog.CreateCategory(ctx, &domain.Category{Name: "Snacks"})
	p, _ := s.catalog.CreateProduct(ctx, &domain.Product{Name: "Chips", PriceCents: 99, CategoryID: cat.ID, IsActive: true})
	if _, err := s.inventory.AssignSlot(ctx, m.ID, "B2", port.SlotInput{ProductID: p.ID, Quantity: 3, MaxCapacity: 10}); err != nil {
		t.Fatal(err)
	}

	const buyers = 10
	var (
		wg              sync.WaitGroup
		mu              sync.Mutex
		reserved, empty int
	)
	for range buyers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sess, err := s.vending.StartSession(ctx, m.ID)
			if err != nil {
				t.Error(err)
				return
			}
			_, err = s.vending.Reserve(ctx, sess.ID, "B2")
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				reserved++
			case errors.Is(err, domain.ErrOutOfStock):
				empty++
			default:
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if reserved != 3 || empty != buyers-3 {
		t.Fatalf("reserved=%d out_of_stock=%d", reserved, empty)
	}
}

func walk[T any](t *testing.T, size int, list func(domain.PageRequest) (domain.Page[T], error), id func(T) string) []string {
	t.Helper()
	var ids []string
	req := domain.PageRequest{Limit: size}
	for range 50 {
		page, err := list(req)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range page.Items {
			ids = append(ids, id(item))
		}
		if page.IsLastPage {
			if page.NextCursor != "" {
				t.Fatalf("last page must not carry a cursor")
			}
			return ids
		}
		if len(page.Items) != size || page.NextCursor == "" {
			t.Fatalf("middle page: %d items, cursor %q", len(page.Items), page.NextCursor)
		}
		req.Cursor = page.NextCursor
	}
	t.Fatal("pagination never reached the last page")
	return nil
}

func assertAllOnce(t *testing.T, got []string, want int) {
	t.Helper()
	seen := map[string]bool{}
	for _, id := range got {
		if seen[id] {
			t.Fatalf("%s returned twice", id)
		}
		seen[id] = true
	}
	if len(seen) != want {
		t.Fatalf("got %d items, want %d", len(seen), want)
	}
}

func TestCursorPagination(t *testing.T) {
	db := newTestDB(t)
	s := newServices(db)
	ctx := context.Background()

	var machineID string
	for i := range 5 {
		m, err := s.machines.Register(ctx, &domain.Machine{Location: fmt.Sprintf("Floor %d", i), Model: "VX"})
		if err != nil {
			t.Fatal(err)
		}
		machineID = m.ID
	}
	ids := walk(t, 2, func(p domain.PageRequest) (domain.Page[domain.Machine], error) {
		return s.machines.List(ctx, "", p)
	}, func(m domain.Machine) string { return m.ID })
	assertAllOnce(t, ids, 5)

	var categoryID string
	for i := range 4 {
		c, err := s.catalog.CreateCategory(ctx, &domain.Category{Name: fmt.Sprintf("cat-%d", i), DisplayOrder: i % 2})
		if err != nil {
			t.Fatal(err)
		}
		categoryID = c.ID
	}
	ids = walk(t, 3, func(p domain.PageRequest) (domain.Page[domain.Category], error) {
		return s.catalog.ListCategories(ctx, p)
	}, func(c domain.Category) string { return c.ID })
	assertAllOnce(t, ids, 4)

	for i := range 7 {
		p, err := s.catalog.CreateProduct(ctx, &domain.Product{Name: "Same name", PriceCents: 100, CategoryID: categoryID, IsActive: true})
		if err != nil {
			t.Fatal(err)
		}
		slot := fmt.Sprintf("A%d", i)
		if _, err := s.inventory.AssignSlot(ctx, machineID, slot, port.SlotInput{ProductID: p.ID, Quantity: i % 3, MaxCapacity: 5, LowThreshold: 1}); err != nil {
			t.Fatal(err)
		}
	}
	ids = walk(t, 3, func(p domain.PageRequest) (domain.Page[domain.Product], error) {
		return s.catalog.ListProducts(ctx, categoryID, p)
	}, func(p domain.Product) string { return p.ID })
	assertAllOnce(t, ids, 7)

	ids = walk(t, 2, func(p domain.PageRequest) (domain.Page[domain.Inventory], error) {
		return s.inventory.ListByMachine(ctx, machineID, p)
	}, func(i domain.Inventory) string { return i.ID })
	assertAllOnce(t, ids, 7)

	ids = walk(t, 2, func(p domain.PageRequest) (domain.Page[domain.Inventory], error) {
		return s.inventory.ListLow(ctx, "", p)
	}, func(i domain.Inventory) string { return i.ID })
	assertAllOnce(t, ids, 5)

	exact, err := s.machines.List(ctx, "", domain.PageRequest{Limit: 5})
	if err != nil || len(exact.Items) != 5 || !exact.IsLastPage {
		t.Fatalf("a page holding exactly the rest must be the last: %v %+v", err, exact)
	}
	if _, err := s.machines.List(ctx, "", domain.PageRequest{Cursor: "not-a-cursor"}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("bad cursor: %v", err)
	}
}
