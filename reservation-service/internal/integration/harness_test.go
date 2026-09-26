// Package integration runs the service as several independent replicas ("pods") against one real Postgres, the way
// it runs in Kubernetes: each pod has its own connection pool, its own sold-out cache, its own bulkhead and its own
// background loops, and shares nothing with the others except the database.
//
// These tests need a Postgres server; without one they are skipped:
//
//	TEST_DATABASE_URL='postgres://postgres:pw@localhost:5432/postgres?sslmode=disable' go test ./internal/integration/
//
// Each test creates (and drops) its own database. The size of the storms is set with MULTIPOD_USERS (default 20000
// requests, each from a different user) and MULTIPOD_WORKERS (requests in flight at once, default 2000).
package integration

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSSt/reservation-service/internal/adapter/outbound/postgres"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/service"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

func envInt(name string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(name)); err == nil && v > 0 {
		return v
	}
	return def
}

// ---- a throwaway database

type database struct {
	base string // TEST_DATABASE_URL, to reach the server
	name string
	dsn  string // this test's own database
}

func newDatabase(t *testing.T) *database {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("set TEST_DATABASE_URL to run the multi-pod tests against Postgres")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, base)
	if err != nil {
		t.Fatalf("connect to %s: %v", base, err)
	}
	defer admin.Close(ctx)

	name := fmt.Sprintf("rs_it_%d_%d", os.Getpid(), time.Now().UnixNano()%1_000_000_000)
	if _, err := admin.Exec(ctx, "create database "+name); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	d := &database{base: base, name: name, dsn: u.String()}

	pool, err := postgres.OpenPool(ctx, d.dsn, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.ApplySchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	pool.Close()

	t.Cleanup(func() {
		c, err := pgx.Connect(context.Background(), base)
		if err != nil {
			return
		}
		defer c.Close(context.Background())
		_, _ = c.Exec(context.Background(), "drop database if exists "+name+" with (force)")
	})
	return d
}

// query runs a statement that returns one integer, for checking what the database really holds.
func (d *database) count(t *testing.T, sql string, args ...any) int {
	t.Helper()
	pool, err := postgres.OpenPool(context.Background(), d.dsn, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var n int
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	return n
}

func (d *database) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	pool, err := postgres.OpenPool(context.Background(), d.dsn, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

// deadlocks is how many deadlocks Postgres has detected in this database so far.
func (d *database) deadlocks(t *testing.T) int {
	time.Sleep(1200 * time.Millisecond) // statistics are flushed about once a second
	return d.count(t, `select deadlocks::int from pg_stat_database where datname = current_database()`)
}

// ---- pods

type podOpts struct {
	maxConns   int
	soldOutTTL time.Duration
	bulkhead   int
	wait       time.Duration
	maxPending int
	holdTTL    time.Duration
	push       outbound.PushGateway
}

func defaultPod() podOpts {
	return podOpts{maxConns: 8, soldOutTTL: 2 * time.Second, bulkhead: 16, wait: 250 * time.Millisecond, maxPending: 3, holdTTL: 15 * time.Minute}
}

type pod struct {
	name          string
	pool          *pgxpool.Pool
	hotels        *postgres.HotelRepository
	reservations  *postgres.ReservationRepository
	res           *service.ReservationService
	waitlist      *service.WaitlistService
	notifications *service.NotificationService
	inbox         *postgres.NotificationRepository
	cache         *service.SoldOutCache
}

func (d *database) newPod(t *testing.T, name string, o podOpts) *pod {
	t.Helper()
	pool, err := postgres.OpenPool(context.Background(), d.dsn, o.maxConns)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	p := &pod{name: name, pool: pool,
		hotels: postgres.NewHotelRepository(pool), reservations: postgres.NewReservationRepository(pool),
		inbox: postgres.NewNotificationRepository(pool), cache: service.NewSoldOutCache(o.soldOutTTL)}
	notifier := service.NewNotifier(p.inbox, nil)
	p.waitlist = service.NewWaitlistService(postgres.NewWaitlistRepository(pool), p.reservations, p.hotels, service.WaitlistOptions{
		OfferTTL: o.holdTTL, MaxPending: o.maxPending, Notifier: notifier, SoldOut: p.cache})
	p.res = service.NewReservationService(p.reservations, p.hotels, service.NewPaymentRail(nil, nil, nil, nil), service.ReservationOptions{
		HoldTTL: o.holdTTL, SoldOut: p.cache, Bulkhead: service.NewBulkhead(o.bulkhead, o.wait),
		MaxPending: o.maxPending, Notifier: notifier, Promoter: p.waitlist})
	p.notifications = service.NewNotificationService(p.inbox, o.push, nil)
	return p
}

func (d *database) pods(t *testing.T, n int, o podOpts) []*pod {
	t.Helper()
	out := make([]*pod, n)
	for i := range out {
		out[i] = d.newPod(t, fmt.Sprintf("pod-%d", i+1), o)
	}
	return out
}

// ---- fixtures

type fixture struct {
	hotelID, typeID int64
	start           time.Time // the first night
}

// seed creates a published 5-star hotel with one room type that has `rooms` rooms for sale on each of `nights` nights.
func (p *pod) seed(t *testing.T, rooms, nights int) fixture {
	t.Helper()
	ctx := context.Background()
	h, err := p.hotels.Create(ctx, domain.Hotel{Name: "Grand", Stars: 5, City: "Da Nang", Address: "1 Beach Rd", PhoneNumber: "+84 236 123 456",
		CheckInTime: "14:00", CheckOutTime: "12:00", Currency: "VND", FreeCancelHours: 48, Status: domain.HotelActive, OwnerID: 5})
	if err != nil {
		t.Fatal(err)
	}
	rt, err := p.hotels.CreateRoomType(ctx, domain.RoomType{HotelID: h.ID, Name: "Suite", Capacity: 4, Active: true})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 30)
	rate := "5000000"
	if err := p.hotels.SetInventory(ctx, outbound.SetInventoryParams{HotelID: h.ID, RoomTypeID: rt.ID,
		From: start, To: start.AddDate(0, 0, nights), Total: &rooms, Rate: &rate}); err != nil {
		t.Fatal(err)
	}
	return fixture{hotelID: h.ID, typeID: rt.ID, start: start}
}

// cmd is a request for `nights` nights starting `from` nights after the first one.
func (f fixture) cmd(from, nights int) inbound.ReserveCommand {
	return inbound.ReserveCommand{HotelID: f.hotelID, RoomTypeID: f.typeID, Rooms: 1, Adults: 2,
		Start: f.start.AddDate(0, 0, from), End: f.start.AddDate(0, 0, from+nights)}
}

// ---- storms

type outcome struct {
	ok, soldOut, busy, capped, other int64
	firstOther                       error
	winners                          []int64
	mu                               sync.Mutex
}

func (o *outcome) record(user int64, err error) {
	switch {
	case err == nil:
		atomic.AddInt64(&o.ok, 1)
		o.mu.Lock()
		o.winners = append(o.winners, user)
		o.mu.Unlock()
	case errors.Is(err, domain.ErrDatesUnavailable):
		atomic.AddInt64(&o.soldOut, 1)
	case errors.Is(err, domain.ErrBusy):
		atomic.AddInt64(&o.busy, 1)
	case errors.Is(err, domain.ErrConflict):
		atomic.AddInt64(&o.capped, 1)
	default:
		atomic.AddInt64(&o.other, 1)
		o.mu.Lock()
		if o.firstOther == nil {
			o.firstOther = err
		}
		o.mu.Unlock()
	}
}

func (o *outcome) String() string {
	return fmt.Sprintf("%d reserved, %d sold out, %d busy, %d over the hold cap, %d other", o.ok, o.soldOut, o.busy, o.capped, o.other)
}

const firstUser = 100_000

// storm sends `users` requests, each from a different user, spread over the pods, with `workers` in flight at once. All
// workers start at the same instant. Each request takes the path an HTTP request takes: the anonymous sold-out check
// first, then the reservation itself.
func storm(t *testing.T, pods []*pod, users, workers int, cmdFor func(i int) inbound.ReserveCommand) *outcome {
	return stormWith(t, func() []*pod { return pods }, users, workers, cmdFor, nil)
}

// stormWith is storm with a live list of pods (one may die during the run) and a hook called as the storm progresses.
func stormWith(t *testing.T, alive func() []*pod, users, workers int, cmdFor func(i int) inbound.ReserveCommand, progress func(done int64)) *outcome {
	t.Helper()
	out := &outcome{}
	var next, done atomic.Int64
	start := make(chan struct{})
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		rng := rand.New(rand.NewSource(int64(w) + time.Now().UnixNano()))
		go func() {
			defer wg.Done()
			<-start
			for {
				i := int(next.Add(1) - 1)
				if i >= users {
					return
				}
				ps := alive()
				p := ps[rng.Intn(len(ps))] // the load balancer sends it anywhere
				user := int64(firstUser + i)
				cmd := cmdFor(i)
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				if p.res.SoldOut(ctx, cmd) {
					out.record(user, domain.ErrDatesUnavailable)
				} else {
					_, err := p.res.Reserve(ctx, inbound.Principal{UserID: user}, cmd)
					out.record(user, err)
				}
				cancel()
				if n := done.Add(1); progress != nil {
					progress(n)
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	return out
}

func (o *outcome) mustHaveNoUnexpectedErrors(t *testing.T) {
	t.Helper()
	if o.other != 0 {
		t.Errorf("%d requests failed in a way that is neither 'sold out' nor 'busy': first: %v", o.other, o.firstOther)
	}
}
