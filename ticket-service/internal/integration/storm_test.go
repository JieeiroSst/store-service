package integration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

// One ticket left and a crowd after it: exactly one buyer gets it, and nobody is told anything but "sold out".
func TestLastTicketGoesToExactlyOneBuyer(t *testing.T) {
	d := newDatabase(t)
	pods := newPods(t, d, 3, podOptions{maxPending: 3, ratePerSec: 5})
	e, tt := seedEvent(t, pods[0], 1, nil)

	users, workers := envInt("STORM_USERS", 20_000), envInt("STORM_WORKERS", 2_000)
	start := time.Now()
	res := storm(t, pods, users, workers, func(u int64) inbound.ReserveCommand { return cmd(e.ID, tt.ID, 1, fmt.Sprintf("req-%d", u)) })
	took := time.Since(start)

	winners, soldOut := 0, 0
	for i, r := range res {
		switch {
		case r.err == nil:
			winners++
			if r.order.UserID != int64(i+1) || r.order.Quantity() != 1 || r.order.Status != domain.OrderPending {
				t.Fatalf("winner's order is wrong: %+v", r.order)
			}
		case errors.Is(r.err, domain.ErrSoldOut):
			soldOut++
		default:
			t.Fatalf("user %d: unexpected error %v", i+1, r.err)
		}
	}
	t.Logf("%d users, %d pods: %d winner, %d told sold out, in %v (%.0f attempts/s)", users, len(pods), winners, soldOut, took, float64(users)/took.Seconds())
	if winners != 1 || soldOut != users-1 {
		t.Fatalf("want exactly 1 winner and %d sold out, got %d and %d", users-1, winners, soldOut)
	}
	c := typeCounters(t, pods[0], tt.ID)
	if c.available != 0 || c.held != 1 || c.sold != 0 {
		t.Fatalf("counters = %+v, want available 0, held 1", c)
	}
	checkLedger(t, pods[0], tt.ID)
}

// A hundred tickets and a crowd: exactly a hundred orders, none oversold, and none left over once everyone has been served.
func TestNeverSellsMoreThanExists(t *testing.T) {
	d := newDatabase(t)
	pods := newPods(t, d, 3, podOptions{maxPending: 3, ratePerSec: 5})
	const tickets = 100
	e, tt := seedEvent(t, pods[0], tickets, nil)

	users, workers := envInt("STORM_USERS", 20_000), envInt("STORM_WORKERS", 2_000)
	res := storm(t, pods, users, workers, func(u int64) inbound.ReserveCommand { return cmd(e.ID, tt.ID, 1, fmt.Sprintf("req-%d", u)) })
	winners := 0
	for i, r := range res {
		switch {
		case r.err == nil:
			winners++
		case errors.Is(r.err, domain.ErrSoldOut):
		default:
			t.Fatalf("user %d: unexpected error %v", i+1, r.err)
		}
	}
	if winners != tickets {
		t.Fatalf("%d orders for %d tickets", winners, tickets)
	}
	c := typeCounters(t, pods[0], tt.ID)
	if c.available != 0 || c.held != tickets {
		t.Fatalf("counters = %+v", c)
	}
	checkLedger(t, pods[0], tt.ID)
}

// Orders of several tickets mixed with single ones: the total can never be exceeded, and the rest stay consistent.
func TestMixedQuantitiesNeverOversell(t *testing.T) {
	d := newDatabase(t)
	pods := newPods(t, d, 3, podOptions{maxPending: 3, ratePerSec: 5})
	const tickets = 250
	e, tt := seedEvent(t, pods[0], tickets, nil)

	res := storm(t, pods, 5_000, 1_000, func(u int64) inbound.ReserveCommand {
		return cmd(e.ID, tt.ID, int(u%4)+1, fmt.Sprintf("req-%d", u))
	})
	held := 0
	for i, r := range res {
		switch {
		case r.err == nil:
			held += r.order.Quantity()
		case errors.Is(r.err, domain.ErrSoldOut):
		default:
			t.Fatalf("user %d: unexpected error %v", i+1, r.err)
		}
	}
	if held > tickets {
		t.Fatalf("OVERSOLD: %d tickets held of %d", held, tickets)
	}
	c := typeCounters(t, pods[0], tt.ID)
	if c.held != held || c.available != tickets-held {
		t.Fatalf("counters = %+v, orders hold %d", c, held)
	}
	checkLedger(t, pods[0], tt.ID)
}

// Retrying the same request (a double click, a client timeout) returns the same order and holds tickets once.
func TestSameRequestHoldsOnce(t *testing.T) {
	d := newDatabase(t)
	pods := newPods(t, d, 3, podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, pods[0], 50, nil)

	const tries = 60
	ids := make([]int64, tries)
	var wg sync.WaitGroup
	for i := 0; i < tries; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			x, err := attempt(context.Background(), pods[i%len(pods)], user(7), cmd(e.ID, tt.ID, 2, "double-click"))
			if err != nil {
				t.Errorf("try %d: %v", i, err)
				return
			}
			ids[i] = x.ID
		}()
	}
	wg.Wait()
	for _, id := range ids {
		if id != ids[0] {
			t.Fatalf("retries produced different orders: %v", ids)
		}
	}
	if c := typeCounters(t, pods[0], tt.ID); c.held != 2 {
		t.Fatalf("held %d tickets for one request of 2", c.held)
	}
	checkLedger(t, pods[0], tt.ID)
}

// One account firing parallel requests cannot get around its caps.
func TestPerUserCaps(t *testing.T) {
	d := newDatabase(t)
	pods := newPods(t, d, 3, podOptions{maxPending: 3, ratePerSec: 1000})
	e, tt := seedEvent(t, pods[0], 500, func(in *inbound.TicketTypeInput) { in.MaxPerUser = 4 })

	var ok atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := attempt(context.Background(), pods[i%len(pods)], user(9), cmd(e.ID, tt.ID, 2, fmt.Sprintf("r-%d", i)))
			if err == nil {
				ok.Add(1)
			} else if !errors.Is(err, domain.ErrConflict) {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
	// max_per_user is 4 tickets = two orders of 2, whichever caps bites first
	if ok.Load() != 2 {
		t.Fatalf("%d orders succeeded, want 2 (max 4 tickets per account)", ok.Load())
	}
	checkLedger(t, pods[0], tt.ID)

	// and the unpaid-orders cap: 3 pending at most
	_, tt2 := seedEvent(t, pods[0], 500, nil)
	var okPending atomic.Int32
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := attempt(context.Background(), pods[i%len(pods)], user(10), cmd(tt2.EventID, tt2.ID, 1, fmt.Sprintf("p-%d", i))); err == nil {
				okPending.Add(1)
			}
		}()
	}
	wg.Wait()
	if okPending.Load() != 3 {
		t.Fatalf("%d pending orders for one user, want 3", okPending.Load())
	}
}

// A promo code with N uses is redeemed exactly N times, however many race for it.
func TestPromoCodeUsesAreExact(t *testing.T) {
	d := newDatabase(t)
	pods := newPods(t, d, 3, podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, pods[0], 1000, nil)
	if _, err := pods[0].events.CreatePromotion(context.Background(), organizer, e.ID, inbound.PromotionInput{
		Code: "early10", Kind: domain.PromoPercent, Value: 10, MaxUses: 5, Active: true}); err != nil {
		t.Fatal(err)
	}

	var withPromo atomic.Int32
	var wg sync.WaitGroup
	for u := int64(1); u <= 200; u++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := cmd(e.ID, tt.ID, 1, fmt.Sprintf("r-%d", u))
			c.PromoCode = "EARLY10"
			x, err := attempt(context.Background(), pods[int(u)%len(pods)], user(u), c)
			if err != nil {
				if !errors.Is(err, domain.ErrInvalid) {
					t.Errorf("user %d: %v", u, err)
				}
				return
			}
			if x.Discount != 50_000 || x.Total != 450_000 {
				t.Errorf("order %d: discount %d total %d", x.ID, x.Discount, x.Total)
			}
			withPromo.Add(1)
		}()
	}
	wg.Wait()
	if withPromo.Load() != 5 {
		t.Fatalf("%d orders got the promo, want exactly 5", withPromo.Load())
	}
	var used int
	if err := pods[0].pool.QueryRow(context.Background(), `select used from promotion where code = 'EARLY10'`).Scan(&used); err != nil || used != 5 {
		t.Fatalf("promo used = %d (%v)", used, err)
	}
}

// Seats: a crowd picks overlapping seats; no seat ends up in two orders, and each order holds exactly what it asked for.
func TestSeatsAreNeverDoubleBooked(t *testing.T) {
	d := newDatabase(t)
	pods := newPods(t, d, 3, podOptions{ratePerSec: 1000})
	ctx := context.Background()
	e, err := pods[0].events.Create(ctx, organizer, inbound.EventInput{Title: "Theatre night", Category: "arts", City: "Hanoi", Venue: "Opera House",
		StartsAt: time.Now().Add(72 * time.Hour), EndsAt: time.Now().Add(75 * time.Hour), Transferable: true})
	if err != nil {
		t.Fatal(err)
	}
	tt, err := pods[0].events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "Stalls", Price: 900_000, Seated: true, Active: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pods[0].events.Submit(ctx, organizer, e.ID); err == nil {
		t.Fatal("an event whose only ticket type has no seats yet was accepted for review")
	}
	if n, err := pods[0].events.GenerateSeats(ctx, organizer, e.ID, tt.ID, inbound.SeatLayout{Rows: 5, SeatsPerRow: 8}); err != nil || n != 40 {
		t.Fatalf("seats: %d %v", n, err)
	}
	publish(t, pods[0], e.ID)
	seats, err := pods[0].erepo.Seats(context.Background(), tt.ID)
	if err != nil || len(seats) != 40 {
		t.Fatalf("seat map: %d %v", len(seats), err)
	}

	// every buyer wants 2 neighbouring seats, and the neighbours overlap
	var wg sync.WaitGroup
	var won atomic.Int32
	for u := int64(1); u <= 400; u++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			i := int(u) % 39
			c := cmd(e.ID, tt.ID, 2, fmt.Sprintf("r-%d", u))
			c.Items[0].SeatIDs = []int64{seats[i].ID, seats[i+1].ID}
			_, err := attempt(context.Background(), pods[int(u)%len(pods)], user(u), c)
			switch {
			case err == nil:
				won.Add(1)
			case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrSoldOut):
			default:
				t.Errorf("user %d: %v", u, err)
			}
		}()
	}
	wg.Wait()

	var dup int
	if err := pods[0].pool.QueryRow(context.Background(), `
		select count(*) from (select s from order_item i join ticket_order o on o.id = i.order_id, unnest(i.seat_ids) s
			where o.status in (1, 2) group by s having count(*) > 1) x`).Scan(&dup); err != nil {
		t.Fatal(err)
	}
	if dup != 0 {
		t.Fatalf("%d seats are in more than one order", dup)
	}
	var heldSeats int
	if err := pods[0].pool.QueryRow(context.Background(), `select count(*) from seat where status = 2 and order_id is not null`).Scan(&heldSeats); err != nil {
		t.Fatal(err)
	}
	if heldSeats != int(won.Load())*2 {
		t.Fatalf("%d seats held for %d orders of 2", heldSeats, won.Load())
	}
	if won.Load() < 1 || won.Load() > 20 {
		t.Fatalf("%d orders of 2 seats among 40 seats", won.Load())
	}
	checkLedger(t, pods[0], tt.ID)
}

// Payments, cancellations and expiries racing each other and new buyers: the ledger always balances.
func TestChurnKeepsTheLedgerBalanced(t *testing.T) {
	d := newDatabase(t)
	wallets := newFakeWallets()
	pods := []*pod{newPod(t, d, wallets, podOptions{ratePerSec: 1000}), newPod(t, d, wallets, podOptions{ratePerSec: 1000}), newPod(t, d, wallets, podOptions{ratePerSec: 1000})}
	const tickets = 60
	e, tt := seedEvent(t, pods[0], tickets, func(in *inbound.TicketTypeInput) { in.MaxPerOrder = 3 })

	var wg sync.WaitGroup
	var paid atomic.Int32
	for u := int64(1); u <= 600; u++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			p := pods[int(u)%len(pods)]
			x, err := attempt(ctx, p, user(u), cmd(e.ID, tt.ID, int(u%3)+1, fmt.Sprintf("r-%d", u)))
			if err != nil {
				return
			}
			switch u % 4 {
			case 0: // pays
				if _, err := p.orders.Pay(ctx, user(u), x.ID, inbound.PayCommand{Method: domain.MethodWallet}); err == nil {
					paid.Add(1)
				} else {
					t.Errorf("pay: %v", err)
				}
			case 1: // gives up
				if _, err := p.orders.Cancel(ctx, user(u), x.ID); err != nil {
					t.Errorf("cancel: %v", err)
				}
			case 2: // lets the hold run out, and pays too late or races the sweeper
				_, _ = pods[0].pool.Exec(ctx, `update ticket_order set expires_at = now() - interval '1 minute' where id = $1`, x.ID)
				go func() { _, _ = p.orders.ReleaseExpired(ctx) }()
				_, _ = p.orders.Pay(ctx, user(u), x.ID, inbound.PayCommand{Method: domain.MethodWallet})
			}
		}()
	}
	wg.Wait()
	for range 3 {
		for _, p := range pods {
			if _, err := p.orders.ReleaseExpired(context.Background()); err != nil {
				t.Fatal(err)
			}
		}
	}
	checkLedger(t, pods[0], tt.ID)
	c := typeCounters(t, pods[0], tt.ID)
	t.Logf("after churn: %+v, %d paid via this test", c, paid.Load())
	if c.sold > tickets || c.sold+c.held+c.available != tickets {
		t.Fatalf("counters = %+v", c)
	}
	// every wallet charge that was not undone belongs to a paid order
	var paidOrders int
	if err := pods[0].pool.QueryRow(context.Background(), `select count(*) from ticket_order where status = 2 and payment_method = 'wallet'`).Scan(&paidOrders); err != nil {
		t.Fatal(err)
	}
	if wallets.charged() != paidOrders {
		t.Fatalf("%d wallet charges standing but %d paid orders: a buyer paid without getting tickets, or the reverse", wallets.charged(), paidOrders)
	}
}

// With every in-memory protection switched off, the database alone must still never sell a ticket twice: this is
// what the atomic UPDATE and the CHECK constraints are for. (The protections exist to spare the database, not to
// make it correct.)
func TestDatabaseAloneNeverOversells(t *testing.T) {
	d := newDatabase(t)
	pods := newPods(t, d, 3, podOptions{raw: true})
	const tickets = 5
	e, tt := seedEvent(t, pods[0], tickets, nil)

	res := storm(t, pods, 3_000, 1_000, func(u int64) inbound.ReserveCommand { return cmd(e.ID, tt.ID, 1, fmt.Sprintf("req-%d", u)) })
	winners := 0
	for i, r := range res {
		switch {
		case r.err == nil:
			winners++
		case errors.Is(r.err, domain.ErrSoldOut):
		default:
			t.Fatalf("user %d: unexpected error %v", i+1, r.err)
		}
	}
	if winners != tickets {
		t.Fatalf("%d orders for %d tickets", winners, tickets)
	}
	if c := typeCounters(t, pods[0], tt.ID); c.available != 0 || c.held != tickets {
		t.Fatalf("counters = %+v", c)
	}
	checkLedger(t, pods[0], tt.ID)
}

// ...and the CHECK constraint itself refuses a negative stock, whatever statement tries it.
func TestConstraintRefusesNegativeStock(t *testing.T) {
	d := newDatabase(t)
	p := newPods(t, d, 1, podOptions{})[0]
	_, tt := seedEvent(t, p, 1, nil)
	_, err := p.pool.Exec(context.Background(), `update ticket_type set available = available - 2 where id = $1`, tt.ID)
	if err == nil {
		t.Fatal("the database accepted a negative number of available tickets")
	}
	_, err = p.pool.Exec(context.Background(), `update ticket_type set sold = 2 where id = $1`, tt.ID)
	if err == nil {
		t.Fatal("the database accepted more sold tickets than exist")
	}
}
