// Package integration runs the service as several independent replicas ("pods") against one real Postgres, the way it
// runs in Kubernetes: each pod has its own connection pool, its own sold-out cache, its own bulkhead and rate limiter,
// and shares nothing with the others except the database.
//
// These tests need a Postgres server; without one they are skipped:
//
//	TEST_DATABASE_URL='postgres://postgres@127.0.0.1:55432/postgres?sslmode=disable' go test -race ./internal/integration/
//
// Each test creates (and drops) its own database. The size of the storms is set with STORM_USERS (default 20000
// purchase attempts, each from a different user) and STORM_WORKERS (attempts in flight at once, default 2000).
package integration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/pdf"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/postgres"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/service"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

func envInt(name string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(name)); err == nil && v > 0 {
		return v
	}
	return def
}

// ---- a throwaway database

type database struct {
	name string
	dsn  string
	base string
}

func newDatabase(t *testing.T) *database {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("set TEST_DATABASE_URL to run the integration tests against Postgres")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, base)
	if err != nil {
		t.Fatalf("connect to %s: %v", base, err)
	}
	defer admin.Close(ctx)

	name := fmt.Sprintf("ts_it_%d_%d", os.Getpid(), time.Now().UnixNano()%1_000_000_000)
	if _, err := admin.Exec(ctx, "create database "+name); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	d := &database{name: name, dsn: u.String(), base: base}

	pool, err := postgres.OpenPool(ctx, d.dsn, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := postgres.ApplySchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		c, err := pgx.Connect(ctx, base)
		if err != nil {
			return
		}
		defer c.Close(ctx)
		_, _ = c.Exec(ctx, "drop database if exists "+name+" with (force)")
	})
	return d
}

// ---- a fake wallet service

type transferRec struct {
	id       string
	from, to string
	amount   int64
	reversed bool
}

type fakeWallets struct {
	mu        sync.Mutex
	next      int
	transfers map[string]bool // id -> still charged
	ledger    []*transferRec
	// fail makes transfers fail: by paying wallet id or by receiving wallet id.
	failFrom, failTo map[string]bool
	// onTransfer runs during every transfer (to make something happen in the middle of a payment).
	onTransfer func()
}

func newFakeWallets() *fakeWallets {
	return &fakeWallets{transfers: map[string]bool{}, failFrom: map[string]bool{}, failTo: map[string]bool{}}
}

func (f *fakeWallets) GetByUser(_ context.Context, userID int64) (domain.Wallet, error) {
	return domain.Wallet{ID: fmt.Sprintf("w-%d", userID), UserID: userID, Currency: "VND", Balance: 1 << 40}, nil
}

func (f *fakeWallets) Transfer(_ context.Context, p outbound.TransferParams) (string, error) {
	f.mu.Lock()
	hook := f.onTransfer
	fail := f.failFrom[p.FromWalletID] || f.failTo[p.ToWalletID]
	f.mu.Unlock()
	if hook != nil {
		hook()
	}
	if fail {
		return "", domain.ErrInsufficientFunds
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.next++
	id := fmt.Sprintf("tr-%d", f.next)
	f.transfers[id] = true
	f.ledger = append(f.ledger, &transferRec{id: id, from: p.FromWalletID, to: p.ToWalletID, amount: p.Amount})
	return id, nil
}

func (f *fakeWallets) ReverseTransfer(_ context.Context, id, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.transfers[id] = false
	for _, r := range f.ledger {
		if r.id == id {
			r.reversed = true
		}
	}
	return nil
}

func (f *fakeWallets) charged() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.transfers {
		if c {
			n++
		}
	}
	return n
}

// balance is what a wallet received minus what it paid, counting only transfers that were not reversed.
func (f *fakeWallets) balance(wallet string) int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	var b int64
	for _, r := range f.ledger {
		if r.reversed {
			continue
		}
		if r.to == wallet {
			b += r.amount
		}
		if r.from == wallet {
			b -= r.amount
		}
	}
	return b
}

// ---- a pod

type pod struct {
	pool     *pgxpool.Pool
	events   *service.EventService
	orders   *service.OrderService
	gate     *service.GateService
	tickets  *service.TicketService
	staff    *service.StaffService
	waitlist *service.WaitlistService
	resale   *service.ResaleService
	rrepo    outbound.ResaleRepository
	venues   *service.VenueService
	docs     *service.DocumentService
	store    *fakeStore
	wallets  *fakeWallets
	repo     outbound.OrderRepository
	erepo    outbound.EventRepository
	trepo    outbound.TicketRepository
}

type podOptions struct {
	maxPending  int
	bulkhead    int
	ratePerSec  float64
	holdMinutes int
	// raw drops every in-memory protection (sold-out cache, bulkhead, rate limit): only the database is left.
	raw bool
	// invoice sets who issues invoices and the VAT rate.
	invoice service.InvoiceOptions
	// email queues e-mails with notifications.
	email bool
	// lifecycle follows orders in a (fake) workflow engine.
	lifecycle outbound.OrderLifecycle
}

func newPod(t testing.TB, d *database, wallets *fakeWallets, o podOptions) *pod {
	t.Helper()
	if o.bulkhead == 0 {
		o.bulkhead = 16
	}
	if o.holdMinutes == 0 {
		o.holdMinutes = 10
	}
	pool, err := postgres.OpenPool(context.Background(), d.dsn, 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	erepo := postgres.NewEventRepository(pool)
	orepo := postgres.NewOrderRepository(pool)
	rail := service.NewPaymentRail(wallets, nil, nil)
	opts := service.OrderOptions{
		HoldTTL:    time.Duration(o.holdMinutes) * time.Minute,
		SoldOut:    service.NewSoldOutCache(2 * time.Second),
		Bulkhead:   service.NewBulkhead(o.bulkhead, 250*time.Millisecond),
		Limiter:    service.NewUserLimiter(o.ratePerSec, 10),
		MaxPending: o.maxPending,
		Notifier:   service.NewNotifier(postgres.NewNotificationRepository(pool), nil).WithEmail(o.email),
		Invoice:    o.invoice,
		Lifecycle:  o.lifecycle,
	}
	if o.raw {
		opts.SoldOut, opts.Bulkhead, opts.Limiter = nil, nil, nil
	}
	trepo := postgres.NewTicketRepository(pool)
	srepo := postgres.NewStaffRepository(pool)
	waitlist := service.NewWaitlistService(postgres.NewWaitlistRepository(pool), erepo, opts.Notifier, nil)
	opts.Waitlist = waitlist
	orders := service.NewOrderService(orepo, erepo, rail, opts)
	rrepo := postgres.NewResaleRepository(pool)
	resale := service.NewResaleService(rrepo, erepo, rail, service.ResaleOptions{
		FeePercent: 10, PlatformWalletID: "platform", MaxTransfers: 2, Notifier: opts.Notifier})
	vrepo := postgres.NewVenueRepository(pool)
	store := newFakeStore()
	docs := service.NewDocumentService(orders, orepo, erepo, trepo, postgres.NewDocumentRepository(pool), store, pdf.NewRenderer(), nil, nil)
	return &pod{pool: pool, docs: docs, store: store, venues: service.NewVenueService(vrepo), events: service.NewEventService(erepo, trepo, 0).WithWaitlist(waitlist).WithVenues(vrepo).WithNotices(orepo, opts.Notifier), orders: orders,
		gate:    service.NewGateService(erepo, trepo, srepo),
		tickets: service.NewTicketService(trepo, erepo, srepo, service.TicketOptions{TransferTTL: 48 * time.Hour, MaxTransfers: 2, Notifier: opts.Notifier}),
		staff:   service.NewStaffService(erepo, srepo), waitlist: waitlist, resale: resale, rrepo: rrepo, wallets: wallets, repo: orepo, erepo: erepo, trepo: trepo}
}

func newPods(t testing.TB, d *database, n int, o podOptions) []*pod {
	pods := make([]*pod, n)
	w := newFakeWallets()
	for i := range pods {
		pods[i] = newPod(t, d, w, o)
	}
	return pods
}

// ---- data

var (
	admin     = inbound.Principal{UserID: 1_000_000_001, Admin: true}
	organizer = inbound.Principal{UserID: 1_000_000_002}
)

func user(id int64) inbound.Principal {
	return inbound.Principal{UserID: id, Email: fmt.Sprintf("user%d@example.com", id)}
}

// seedEvent creates a published event with one general-admission ticket type of the given size.
func seedEvent(t testing.TB, p *pod, tickets int, tweak func(*inbound.TicketTypeInput)) (domain.Event, domain.TicketType) {
	t.Helper()
	ctx := context.Background()
	e, err := p.events.Create(ctx, organizer, inbound.EventInput{
		Title: "Flash Sale Concert", Category: "music", City: "Ho Chi Minh City", Venue: "Stadium",
		StartsAt: time.Now().Add(72 * time.Hour), EndsAt: time.Now().Add(76 * time.Hour), WalletID: "organizer-wallet",
		Transferable: true, ResaleCapPercent: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	in := inbound.TicketTypeInput{Name: "GA", Price: 500_000, Total: tickets, MaxPerOrder: 10, Active: true}
	if tweak != nil {
		tweak(&in)
	}
	tt, err := p.events.CreateTicketType(ctx, organizer, e.ID, in)
	if err != nil {
		t.Fatal(err)
	}
	publish(t, p, e.ID)
	return e, tt
}

func publish(t testing.TB, p *pod, eventID int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := p.events.Submit(ctx, organizer, eventID); err != nil {
		t.Fatal(err)
	}
	if _, err := p.events.Approve(ctx, admin, eventID); err != nil {
		t.Fatal(err)
	}
}

func cmd(eventID, typeID int64, qty int, reqID string) inbound.ReserveCommand {
	return inbound.ReserveCommand{
		EventID: eventID, Items: []inbound.ReserveItem{{TicketTypeID: typeID, Quantity: qty}},
		BuyerName: "Buyer", BuyerEmail: "buyer@example.com", RequestID: reqID,
	}
}

// attempt is what one buyer's client does: the cheap sold-out check first (as the HTTP handler does before it
// authenticates), then the real reservation; a 429 ("busy") is retried, as a real client would.
func attempt(ctx context.Context, p *pod, u inbound.Principal, c inbound.ReserveCommand) (domain.Order, error) {
	for try := 0; ; try++ {
		if p.orders.SoldOut(ctx, c) {
			return domain.Order{}, &domain.SoldOutError{}
		}
		x, err := p.orders.Reserve(ctx, u, c)
		if errors.Is(err, domain.ErrBusy) && try < 200 {
			time.Sleep(time.Duration(1+try%5) * time.Millisecond)
			continue
		}
		return x, err
	}
}

// storm sends users purchase attempts, spread over the pods, workers at a time, and returns each user's result.
type outcome struct {
	order domain.Order
	err   error
}

func storm(t testing.TB, pods []*pod, users, workers int, mk func(user int64) inbound.ReserveCommand) []outcome {
	t.Helper()
	out := make([]outcome, users)
	var next atomic.Int64
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := int(next.Add(1)) - 1
				if i >= users {
					return
				}
				uid := int64(i + 1)
				x, err := attempt(context.Background(), pods[i%len(pods)], user(uid), mk(uid))
				out[i] = outcome{x, err}
			}
		}()
	}
	wg.Wait()
	return out
}

// ---- invariants, straight from the database

type counters struct{ total, available, sold, held int }

func typeCounters(t testing.TB, p *pod, typeID int64) counters {
	t.Helper()
	var c counters
	err := p.pool.QueryRow(context.Background(),
		`select total, available, sold, total - available - sold from ticket_type where id = $1`, typeID).Scan(&c.total, &c.available, &c.sold, &c.held)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// checkLedger verifies that the counters of a ticket type agree with the orders, item by item: nothing sold or held
// that no order accounts for, nothing an order holds that the counters do not show.
func checkLedger(t testing.TB, p *pod, typeID int64) {
	t.Helper()
	c := typeCounters(t, p, typeID)
	if c.available < 0 || c.sold < 0 || c.held < 0 {
		t.Fatalf("counters went negative: %+v", c)
	}
	var pendingQty, paidQty int
	err := p.pool.QueryRow(context.Background(), `
		select coalesce(sum(i.quantity) filter (where o.status = 1), 0), coalesce(sum(i.quantity) filter (where o.status = 2), 0)
		from order_item i join ticket_order o on o.id = i.order_id where i.ticket_type_id = $1`, typeID).Scan(&pendingQty, &paidQty)
	if err != nil {
		t.Fatal(err)
	}
	if c.held != pendingQty || c.sold != paidQty {
		t.Fatalf("ledger mismatch: type says held=%d sold=%d, orders say pending=%d paid=%d", c.held, c.sold, pendingQty, paidQty)
	}
	// every ticket of a paid order counts as sold, whatever became of it since (used, expired, even revoked: a revoked
	// ticket is not sold again); the tickets of refunded orders went back on sale
	var tickets int
	if err := p.pool.QueryRow(context.Background(), `select count(*) from ticket t join ticket_order o on o.id = t.order_id
		where t.ticket_type_id = $1 and o.status = 2`, typeID).Scan(&tickets); err != nil {
		t.Fatal(err)
	}
	if tickets != c.sold {
		t.Fatalf("%d tickets issued for paid orders but %d sold", tickets, c.sold)
	}
}

func isSoldOutOrBusy(err error) bool {
	return errors.Is(err, domain.ErrSoldOut) || errors.Is(err, domain.ErrBusy)
}

// ---- a fake document store (upload-service)

type storedFile struct {
	user int64
	name string
	data []byte
}

type fakeStore struct {
	mu    sync.Mutex
	next  int
	files map[string]storedFile
	// putFail makes the next Put calls fail.
	putFail int
}

func newFakeStore() *fakeStore { return &fakeStore{files: map[string]storedFile{}} }

func (f *fakeStore) Put(_ context.Context, user int64, name string, data []byte) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.putFail > 0 {
		f.putFail--
		return "", domain.ErrUpstreamUnavailable
	}
	f.next++
	id := fmt.Sprintf("file-%d", f.next)
	f.files[id] = storedFile{user: user, name: name, data: data}
	return id, nil
}

func (f *fakeStore) Open(_ context.Context, id string) (io.ReadCloser, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	x, ok := f.files[id]
	if !ok {
		return nil, 0, domain.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(x.data)), int64(len(x.data)), nil
}

func (f *fakeStore) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.files, id)
	return nil
}

func (f *fakeStore) of(user int64) []storedFile {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []storedFile
	for _, x := range f.files {
		if x.user == user {
			out = append(out, x)
		}
	}
	return out
}

// newBareDocs is a document service with no store behind it.
func newBareDocs(p *pod) *service.DocumentService {
	return service.NewDocumentService(p.orders, p.repo, p.erepo, p.trepo, postgres.NewDocumentRepository(p.pool), nil, pdf.NewRenderer(), nil, nil)
}

func fmtInt(n int64) string { return fmt.Sprintf("%d", n) }
func pad8(n int64) string   { return fmt.Sprintf("%08d", n) }
func time4(o domain.Order) string {
	y := o.CreatedAt.Year()
	if o.PaidAt != nil {
		y = o.PaidAt.Year()
	}
	return fmt.Sprintf("%d", y)
}

// moveEvent puts an event (and its only session) at other times, given as SQL expressions such as
// "now() - interval '3 hours'": how tests make the clock move.
func moveEvent(ctx context.Context, p *pod, eventID int64, starts, ends string) (any, error) {
	if _, err := p.pool.Exec(ctx, `update event set starts_at = `+starts+`, ends_at = `+ends+` where id = $1`, eventID); err != nil {
		return nil, err
	}
	_, err := p.pool.Exec(ctx, `update event_session set starts_at = `+starts+`, ends_at = `+ends+` where event_id = $1`, eventID)
	return nil, err
}
