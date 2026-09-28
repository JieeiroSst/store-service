package application

import (
	"context"
	"errors"
	"maps"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/config"
	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
)

type store struct {
	machines     map[string]domain.Machine
	maintenance  []domain.MaintenanceLog
	categories   map[string]domain.Category
	products     map[string]domain.Product
	inventory    map[string]domain.Inventory
	sessions     map[string]domain.Session
	reservations map[string]domain.Reservation
	payments     map[string]domain.Payment
	orders       map[string]domain.Order
	events       []domain.Event
	dispatched   map[string]bool
	orderNo      int64
}

func newStore() *store {
	return &store{
		machines:     map[string]domain.Machine{},
		categories:   map[string]domain.Category{},
		products:     map[string]domain.Product{},
		inventory:    map[string]domain.Inventory{},
		sessions:     map[string]domain.Session{},
		reservations: map[string]domain.Reservation{},
		payments:     map[string]domain.Payment{},
		orders:       map[string]domain.Order{},
		dispatched:   map[string]bool{},
	}
}

func get[T any](m map[string]T, id string) (*T, error) {
	v, ok := m[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &v, nil
}

func put[T any](m map[string]T, id string, v T) error {
	if _, ok := m[id]; !ok {
		return domain.ErrNotFound
	}
	m[id] = v
	return nil
}

var errNotUsed = errors.New("not used")

type fakeTx struct{}

func (fakeTx) WithinTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type machineRepo struct{ *store }

func (r machineRepo) Create(_ context.Context, m *domain.Machine) error {
	r.machines[m.ID] = *m
	return nil
}

func (r machineRepo) Get(_ context.Context, id string) (*domain.Machine, error) {
	return get(r.machines, id)
}

func (r machineRepo) List(context.Context, string, domain.PageRequest) (domain.Page[domain.Machine], error) {
	return domain.Page[domain.Machine]{}, errNotUsed
}

func (r machineRepo) Update(_ context.Context, m *domain.Machine) error {
	return put(r.machines, m.ID, *m)
}

type maintenanceRepo struct{ *store }

func (r maintenanceRepo) Create(_ context.Context, l *domain.MaintenanceLog) error {
	r.maintenance = append(r.maintenance, *l)
	return nil
}

func (r maintenanceRepo) ListByMachine(context.Context, string) ([]domain.MaintenanceLog, error) {
	return r.maintenance, nil
}

type productRepo struct{ *store }

func (r productRepo) Create(_ context.Context, p *domain.Product) error {
	r.products[p.ID] = *p
	return nil
}

func (r productRepo) Get(_ context.Context, id string) (*domain.Product, error) {
	return get(r.products, id)
}

func (r productRepo) Update(_ context.Context, p *domain.Product) error {
	return put(r.products, p.ID, *p)
}

func (r productRepo) List(context.Context, string, domain.PageRequest) (domain.Page[domain.Product], error) {
	return domain.Page[domain.Product]{}, errNotUsed
}

type inventoryRepo struct{ *store }

func (r inventoryRepo) withProduct(i domain.Inventory) *domain.Inventory {
	if p, ok := r.products[i.ProductID]; ok {
		i.Product = &p
	}
	return &i
}

func (r inventoryRepo) Create(_ context.Context, i *domain.Inventory) error {
	r.inventory[i.ID] = *i
	return nil
}

func (r inventoryRepo) Update(_ context.Context, i *domain.Inventory) error {
	return put(r.inventory, i.ID, *i)
}

func (r inventoryRepo) Get(ctx context.Context, id string) (*domain.Inventory, error) {
	return r.GetForUpdate(ctx, id)
}

func (r inventoryRepo) GetForUpdate(_ context.Context, id string) (*domain.Inventory, error) {
	i, ok := r.inventory[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return r.withProduct(i), nil
}

func (r inventoryRepo) GetBySlotForUpdate(_ context.Context, machineID, slot string) (*domain.Inventory, error) {
	for _, i := range r.inventory {
		if i.MachineID == machineID && i.SlotIdentifier == slot {
			return r.withProduct(i), nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r inventoryRepo) ListByMachine(context.Context, string, domain.PageRequest) (domain.Page[domain.Inventory], error) {
	return domain.Page[domain.Inventory]{}, errNotUsed
}

func (r inventoryRepo) ListLow(context.Context, string, domain.PageRequest) (domain.Page[domain.Inventory], error) {
	return domain.Page[domain.Inventory]{}, errNotUsed
}

type sessionRepo struct{ *store }

func (r sessionRepo) Create(_ context.Context, s *domain.Session) error {
	r.sessions[s.ID] = *s
	return nil
}

func (r sessionRepo) Get(_ context.Context, id string) (*domain.Session, error) {
	return get(r.sessions, id)
}

func (r sessionRepo) GetForUpdate(_ context.Context, id string) (*domain.Session, error) {
	return get(r.sessions, id)
}

func (r sessionRepo) Update(_ context.Context, s *domain.Session) error {
	return put(r.sessions, s.ID, *s)
}

func (r sessionRepo) ListExpired(_ context.Context, now time.Time, _ int) ([]string, error) {
	var ids []string
	for _, s := range r.sessions {
		if s.Status == domain.SessionActive && !now.Before(s.ExpiresAt) {
			ids = append(ids, s.ID)
		}
	}
	return ids, nil
}

type reservationRepo struct{ *store }

func (r reservationRepo) Create(_ context.Context, res *domain.Reservation) error {
	r.reservations[res.ID] = *res
	return nil
}

func (r reservationRepo) Get(_ context.Context, id string) (*domain.Reservation, error) {
	return get(r.reservations, id)
}

func (r reservationRepo) GetForUpdate(_ context.Context, id string) (*domain.Reservation, error) {
	return get(r.reservations, id)
}

func (r reservationRepo) Update(_ context.Context, res *domain.Reservation) error {
	return put(r.reservations, res.ID, *res)
}

func (r reservationRepo) ListExpiredPending(_ context.Context, now time.Time, _ int) ([]string, error) {
	var ids []string
	for _, res := range r.reservations {
		if res.Status == domain.ReservationPending && !now.Before(res.ExpiresAt) {
			ids = append(ids, res.ID)
		}
	}
	return ids, nil
}

func (r reservationRepo) ListPendingBySession(_ context.Context, sessionID string) ([]string, error) {
	var ids []string
	for _, res := range r.reservations {
		if res.SessionID == sessionID && res.Status == domain.ReservationPending {
			ids = append(ids, res.ID)
		}
	}
	return ids, nil
}

type paymentRepo struct{ *store }

func (r paymentRepo) Create(_ context.Context, p *domain.Payment) error {
	for _, existing := range r.payments {
		if existing.ReservationID == p.ReservationID &&
			(existing.Status == domain.PaymentPending || existing.Status == domain.PaymentCompleted) {
			return domain.ErrConflict
		}
	}
	cp := *p
	cp.Metadata = maps.Clone(p.Metadata)
	r.payments[p.ID] = cp
	return nil
}

func (r paymentRepo) Get(_ context.Context, id string) (*domain.Payment, error) {
	p, err := get(r.payments, id)
	if err == nil {
		p.Metadata = maps.Clone(p.Metadata)
	}
	return p, err
}

func (r paymentRepo) GetForUpdate(ctx context.Context, id string) (*domain.Payment, error) {
	return r.Get(ctx, id)
}

func (r paymentRepo) ListStalePending(_ context.Context, before time.Time, _ int) ([]string, error) {
	var ids []string
	for _, p := range r.payments {
		if p.Status == domain.PaymentPending && !p.CreatedAt.After(before) {
			ids = append(ids, p.ID)
		}
	}
	return ids, nil
}

func (r paymentRepo) Update(_ context.Context, p *domain.Payment) error {
	cp := *p
	cp.Metadata = maps.Clone(p.Metadata)
	return put(r.payments, p.ID, cp)
}

type orderRepo struct{ *store }

func (r orderRepo) Create(_ context.Context, o *domain.Order) error {
	r.orderNo++
	o.OrderNo = r.orderNo
	r.orders[o.ID] = *o
	return nil
}

func (r orderRepo) Get(_ context.Context, id string) (*domain.Order, error) {
	return get(r.orders, id)
}

func (r orderRepo) GetForUpdate(_ context.Context, id string) (*domain.Order, error) {
	return get(r.orders, id)
}

func (r orderRepo) Update(_ context.Context, o *domain.Order) error {
	return put(r.orders, o.ID, *o)
}

func (r orderRepo) GetByPayment(_ context.Context, paymentID string) (*domain.Order, error) {
	for _, o := range r.orders {
		if o.PaymentID == paymentID {
			return &o, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r orderRepo) ListBySession(_ context.Context, sessionID string) ([]domain.Order, error) {
	var out []domain.Order
	for _, o := range r.orders {
		if o.SessionID == sessionID {
			out = append(out, o)
		}
	}
	return out, nil
}

type eventRepo struct{ *store }

func (r eventRepo) Append(_ context.Context, e *domain.Event) error {
	r.events = append(r.events, *e)
	return nil
}

func (r eventRepo) ListByMachine(context.Context, string, int) ([]domain.Event, error) {
	return r.events, nil
}

func (r eventRepo) LockUndispatched(_ context.Context, limit int) ([]domain.Event, error) {
	var out []domain.Event
	for _, e := range r.events {
		if !r.dispatched[e.ID] && len(out) < limit {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r eventRepo) MarkDispatched(_ context.Context, ids []string, _ time.Time) error {
	for _, id := range ids {
		r.dispatched[id] = true
	}
	return nil
}

type fakeReports struct {
	sales []domain.ProductSales
}

func (f fakeReports) SalesByProduct(context.Context, string, time.Time, time.Time) ([]domain.ProductSales, error) {
	return f.sales, nil
}

type fakeGateway struct {
	lookupTx  string
	lookupErr error
	lookups   int
	chargeErr error
	refundErr error
	charges   int
	refunds   []string
	last      port.ChargeRequest
	onCharge  func()
}

func (g *fakeGateway) Charge(_ context.Context, req port.ChargeRequest) (string, error) {
	g.charges++
	g.last = req
	if g.onCharge != nil {
		g.onCharge()
	}
	if g.chargeErr != nil {
		return "", g.chargeErr
	}
	return "txn-1", nil
}

func (g *fakeGateway) Lookup(context.Context, port.LookupRequest) (string, bool, error) {
	g.lookups++
	return g.lookupTx, g.lookupTx != "", g.lookupErr
}

func (g *fakeGateway) Refund(_ context.Context, req port.RefundRequest) error {
	if g.refundErr != nil {
		return g.refundErr
	}
	g.refunds = append(g.refunds, req.TransactionID)
	return nil
}

type fakeCoupons struct {
	discount  int
	quoteErr  error
	redeemErr error
	redeemed  []int64
}

func (c *fakeCoupons) Quote(context.Context, string, int) (int, error) {
	return c.discount, c.quoteErr
}

func (c *fakeCoupons) Redeem(_ context.Context, _ string, orderNo int64, _ int) error {
	if c.redeemErr != nil {
		return c.redeemErr
	}
	c.redeemed = append(c.redeemed, orderNo)
	return nil
}

type fakeNotifier struct {
	sent   []string
	failOn string
}

func (n *fakeNotifier) Notify(_ context.Context, e domain.Event) error {
	if e.EventType == n.failOn {
		return errors.New("notification service down")
	}
	n.sent = append(n.sent, e.EventType)
	return nil
}

var testCfg = &config.Config{Vending: config.VendingConfig{
	SessionTTL:     5 * time.Minute,
	ReservationTTL: 2 * time.Minute,
	ReconcileAfter: time.Minute,
	Currency:       "USD",
}}
