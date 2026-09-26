package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/postgres"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/service"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type show struct {
	event    domain.Event
	sessions []domain.Session
	ga       []domain.TicketType // the GA type of each session, in the same order
}

// newShow makes a published event whose sessions start at the given offsets from now (each lasting three hours). The first
// session has a GA type of the given size; the others copy it.
func newShow(t testing.TB, p *pod, tickets int, refundHours int, starts ...time.Duration) show {
	t.Helper()
	ctx := context.Background()
	now := time.Now()
	e, err := p.events.Create(ctx, organizer, inbound.EventInput{Title: "Anh Trai Say Hi", Category: "music", City: "Hue", Venue: "Stadium", WalletID: "w",
		StartsAt: now.Add(starts[0]), EndsAt: now.Add(starts[0] + 3*time.Hour), Transferable: true, ResaleCapPercent: 100, RefundCutoffHours: refundHours})
	if err != nil {
		t.Fatal(err)
	}
	ga, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "GA", Price: 500_000, Total: tickets, MaxPerOrder: 10, Active: true})
	if err != nil {
		t.Fatal(err)
	}
	sh := show{sessions: []domain.Session{}, ga: []domain.TicketType{ga}}
	e, _ = p.events.View(ctx, &organizer, e.ID)
	sh.sessions = append(sh.sessions, e.Sessions[0])
	for _, off := range starts[1:] {
		sess, err := p.events.CreateSession(ctx, organizer, e.ID, inbound.SessionInput{StartsAt: now.Add(off), EndsAt: now.Add(off + 3*time.Hour), CopyFrom: sh.sessions[0].ID})
		if err != nil {
			t.Fatal(err)
		}
		sh.sessions = append(sh.sessions, sess)
	}
	publish(t, p, e.ID)
	sh.event, _ = p.events.View(ctx, &organizer, e.ID)
	sh.ga = nil
	for _, s := range sh.sessions {
		for _, tt := range sh.event.TicketTypes {
			if tt.SessionID == s.ID && tt.Name == "GA" {
				sh.ga = append(sh.ga, tt)
			}
		}
	}
	if len(sh.ga) != len(sh.sessions) {
		t.Fatalf("every session should have its GA type: %d types for %d sessions", len(sh.ga), len(sh.sessions))
	}
	return sh
}

func orderFor(t testing.TB, p *pod, u int64, sh show, i, qty int, paid bool) domain.Order {
	t.Helper()
	ctx := context.Background()
	o, err := p.orders.Reserve(ctx, user(u), cmd(sh.event.ID, sh.ga[i].ID, qty, "s-"+itoa(u)+"-"+itoa(int64(i))+"-"+itoa(time.Now().UnixNano())))
	if err != nil {
		t.Fatal(err)
	}
	if paid {
		if o, err = p.orders.Pay(ctx, user(u), o.ID, inbound.PayCommand{Method: domain.MethodWallet}); err != nil {
			t.Fatal(err)
		}
	}
	return o
}

// One event, several showtimes: each sells, holds and counts on its own.
func TestSessionsSellIndependently(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	sh := newShow(t, p, 2, 0, 72*time.Hour, 96*time.Hour, 100*time.Hour)

	e := sh.event
	if len(e.Sessions) != 3 || len(e.TicketTypes) != 3 {
		t.Fatalf("%d sessions, %d ticket types", len(e.Sessions), len(e.TicketTypes))
	}
	if !e.StartsAt.Equal(sh.sessions[0].StartsAt) || !e.EndsAt.Equal(sh.sessions[2].EndsAt) {
		t.Fatalf("the event's dates must be the first start (%v) and the last end (%v): %v - %v", sh.sessions[0].StartsAt, sh.sessions[2].EndsAt, e.StartsAt, e.EndsAt)
	}
	// the first showtime sells out; the others do not notice
	o1 := orderFor(t, p, 5, sh, 0, 2, true)
	if _, err := p.orders.Reserve(ctx, user(6), cmd(e.ID, sh.ga[0].ID, 1, "late")); !errors.Is(err, domain.ErrSoldOut) {
		t.Fatalf("a sold-out showtime: %v", err)
	}
	o2 := orderFor(t, p, 6, sh, 1, 2, false)
	if o1.SessionID != sh.sessions[0].ID || o2.SessionID != sh.sessions[1].ID {
		t.Fatalf("orders belong to their showtime: %d %d", o1.SessionID, o2.SessionID)
	}
	for _, tk := range o1.Tickets {
		if tk.SessionID != sh.sessions[0].ID {
			t.Fatalf("ticket of session %d, want %d", tk.SessionID, sh.sessions[0].ID)
		}
	}
	for i, want := range []counters{{total: 2, available: 0, sold: 2}, {total: 2, available: 0, held: 2}, {total: 2, available: 2}} {
		if c := typeCounters(t, p, sh.ga[i].ID); c.available != want.available || c.sold != want.sold || c.held != want.held {
			t.Fatalf("session %d: %+v, want %+v", i, c, want)
		}
		checkLedger(t, p, sh.ga[i].ID)
	}
	// the same ticket type name in every showtime; new types default to the next showtime still to come
	vip, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "VIP", Price: 1, Total: 3, Active: true})
	if err != nil || vip.SessionID != sh.sessions[0].ID {
		t.Fatalf("a new type goes to the first showtime: %+v %v", vip, err)
	}
	// one order is for one showtime
	mixed := inbound.ReserveCommand{EventID: e.ID, BuyerName: "B", BuyerEmail: "b@example.com", Items: []inbound.ReserveItem{
		{TicketTypeID: vip.ID, Quantity: 1}, {TicketTypeID: sh.ga[2].ID, Quantity: 1}}}
	if _, err := p.orders.Reserve(ctx, user(7), mixed); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("an order across two showtimes: %v", err)
	}
	if _, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "VIP", Price: 1, Total: 3, Active: true, SessionID: sh.sessions[1].ID}); err != nil {
		t.Fatalf("the same name in another showtime: %v", err)
	}
	if _, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "VIP", Price: 1, Total: 3, Active: true, SessionID: sh.sessions[0].ID}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("the same name twice in one showtime: %v", err)
	}
	if _, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "X", Price: 1, Total: 3, Active: true, SessionID: 9999}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("a type for a showtime of no event: %v", err)
	}
	if _, err := p.events.UpdateTicketType(ctx, organizer, e.ID, vip.ID, inbound.TicketTypeInput{Name: "VIP", Price: 1, Total: 3, Active: true, SessionID: sh.sessions[2].ID}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("moving a type to another showtime: %v", err)
	}
	// the buyer sees the showtime on the ticket
	mine, _ := p.tickets.List(ctx, user(5), inbound.TicketListQuery{})
	if len(mine.Items) != 2 || !mine.Items[0].EventStartsAt.Equal(sh.sessions[0].StartsAt) || !mine.Items[0].EventEndsAt.Equal(sh.sessions[0].EndsAt) {
		t.Fatalf("tickets show their showtime: %+v", mine.Items)
	}
}

// Refund cutoffs, transfers and resale follow the showtime of the ticket, not the first showtime of the event.
func TestSessionDatesDecideRefundsTransfersAndResale(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	// refunds until 48 hours before a showtime; the first showtime is far away, the second is tomorrow
	sh := newShow(t, p, 5, 48, 120*time.Hour, 30*time.Hour)

	far, soon := orderFor(t, p, 5, sh, 0, 1, true), orderFor(t, p, 6, sh, 1, 1, true)
	if _, err := p.orders.Cancel(ctx, user(6), soon.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("refunding a showtime that is 30 hours away (cutoff 48): %v", err)
	}
	if r, err := p.orders.Cancel(ctx, user(5), far.ID); err != nil || r.Status != domain.OrderRefunded {
		t.Fatalf("refunding a showtime that is 120 hours away: %+v %v", r, err)
	}
	// a transfer of a ticket lapses when *its* showtime starts
	x, err := p.tickets.Offer(ctx, user(6), soon.Tickets[0].ID, inbound.TransferCommand{ToUserID: 9})
	if err != nil {
		t.Fatal(err)
	}
	if !x.ExpiresAt.Before(sh.sessions[1].StartsAt.Add(time.Second)) || x.ExpiresAt.After(sh.sessions[1].StartsAt) {
		t.Fatalf("the offer must lapse at or before its showtime (%v): %v", sh.sessions[1].StartsAt, x.ExpiresAt)
	}
	// once that showtime is under way it cannot be transferred or resold, though the event as a whole is still ahead
	if _, err := moveEvent2(ctx, p, sh.sessions[1].ID, "now() - interval '1 hour'", "now() + interval '2 hours'"); err != nil {
		t.Fatal(err)
	}
	other := orderFor(t, p, 7, sh, 0, 1, true) // the first showtime is still ahead: still transferable
	if _, err := p.tickets.Offer(ctx, user(7), other.Tickets[0].ID, inbound.TransferCommand{ToUserID: 9}); err != nil {
		t.Fatalf("a ticket of a showtime still ahead: %v", err)
	}
	if _, err := p.tickets.CancelTransfer(ctx, user(6), x.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := p.tickets.Offer(ctx, user(6), soon.Tickets[0].ID, inbound.TransferCommand{ToUserID: 9}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("transferring a ticket of a showtime that started: %v", err)
	}
	if _, err := p.resale.List(ctx, user(6), soon.Tickets[0].ID, 100_000); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("reselling a ticket of a showtime that started: %v", err)
	}
	if _, err := p.resale.List(ctx, user(8), other.Tickets[0].ID, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("someone else's ticket: %v", err)
	}
}

// moveEvent2 moves one session (and the event's dates to match the sessions).
func moveEvent2(ctx context.Context, p *pod, sessionID int64, starts, ends string) (any, error) {
	if _, err := p.pool.Exec(ctx, `update event_session set starts_at = `+starts+`, ends_at = `+ends+` where id = $1`, sessionID); err != nil {
		return nil, err
	}
	_, err := p.pool.Exec(ctx, `update event e set starts_at = (select min(starts_at) from event_session where event_id = e.id),
		ends_at = (select max(ends_at) from event_session where event_id = e.id) where id = (select event_id from event_session where id = $1)`, sessionID)
	return nil, err
}

// The gate is open for a ticket while its own showtime is; expiry and reminders follow it too.
func TestSessionDatesDecideTheGateExpiryAndReminders(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000, email: false})
	sh := newShow(t, p, 5, 0, 72*time.Hour, 96*time.Hour, 20*time.Hour)

	a := orderFor(t, p, 5, sh, 0, 1, true) // in 72 hours
	b := orderFor(t, p, 6, sh, 1, 1, true) // in 96 hours
	c := orderFor(t, p, 7, sh, 2, 1, true) // in 20 hours: inside the reminder window

	// the reminder goes to the buyer of the showtime that is near, and only to them
	if n, err := p.orders.SendReminders(ctx); err != nil || n != 1 {
		t.Fatalf("reminders: %d %v", n, err)
	}
	var who int64
	if err := p.pool.QueryRow(ctx, `select user_id from notification where kind = 'event_reminder'`).Scan(&who); err != nil || who != 7 {
		t.Fatalf("reminded user %d (%v), want 7", who, err)
	}
	_ = a
	_ = b

	// the first showtime is over (it started 5 hours ago and ended 2 hours ago); the second is yet to come
	if _, err := moveEvent2(ctx, p, sh.sessions[0].ID, "now() - interval '5 hours'", "now() - interval '2 hours'"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.gate.CheckIn(ctx, organizer, sh.event.ID, a.Tickets[0].Code); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("scanning a ticket of a showtime that is over: %v", err)
	}
	if _, err := p.gate.CheckIn(ctx, organizer, sh.event.ID, c.Tickets[0].Code); err != nil {
		t.Fatalf("scanning a ticket of a showtime still ahead: %v", err)
	}
	res, err := p.gate.BatchCheckIn(ctx, organizer, sh.event.ID, []inbound.ScanItem{{Code: a.Tickets[0].Code}, {Code: b.Tickets[0].Code}})
	if err != nil || res[0].Result != "event_over" || res[1].Result != "admitted" {
		t.Fatalf("a batch across showtimes: %+v %v", res, err)
	}
	// only the tickets of the showtime that ended (six hours' grace aside) expire
	if _, err := moveEvent2(ctx, p, sh.sessions[0].ID, "now() - interval '20 hours'", "now() - interval '17 hours'"); err != nil {
		t.Fatal(err)
	}
	if n, err := p.tickets.ExpireTickets(ctx); err != nil || n != 1 {
		t.Fatalf("expire: %d %v", n, err)
	}
	if status(t, p, a.Tickets[0].ID) != domain.TicketExpired || status(t, p, c.Tickets[0].ID) != domain.TicketUsed || status(t, p, b.Tickets[0].ID) != domain.TicketUsed {
		t.Fatalf("a=%s b=%s c=%s", status(t, p, a.Tickets[0].ID).Name(), status(t, p, b.Tickets[0].ID).Name(), status(t, p, c.Tickets[0].ID).Name())
	}
}

// Cancelling one showtime refunds its orders and leaves the others alone.
func TestCancellingOneSession(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	wallets := newFakeWallets()
	p := newPod(t, d, wallets, podOptions{ratePerSec: 1000})
	sh := newShow(t, p, 10, 0, 72*time.Hour, 96*time.Hour)

	paid0, pend0 := orderFor(t, p, 5, sh, 0, 2, true), orderFor(t, p, 6, sh, 0, 1, false)
	paid1 := orderFor(t, p, 7, sh, 1, 1, true)

	if _, err := p.events.CancelSession(ctx, user(5), sh.event.ID, sh.sessions[0].ID, "x"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a buyer cancelling a showtime: %v", err)
	}
	if _, err := p.events.CancelSession(ctx, organizer, sh.event.ID, sh.sessions[0].ID, " "); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("cancelling without a reason: %v", err)
	}
	if _, err := p.events.CancelSession(ctx, organizer, sh.event.ID, 99999, "x"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("an unknown showtime: %v", err)
	}
	cs, err := p.events.CancelSession(ctx, organizer, sh.event.ID, sh.sessions[0].ID, "the singer is ill")
	if err != nil || cs.Status != domain.SessionCancelled || cs.CancelReason != "the singer is ill" {
		t.Fatalf("cancel: %+v %v", cs, err)
	}
	if _, err := p.events.CancelSession(ctx, organizer, sh.event.ID, sh.sessions[0].ID, "again"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("cancelling twice: %v", err)
	}
	// nothing more is sold for it, though the event goes on
	if _, err := p.orders.Reserve(ctx, user(8), cmd(sh.event.ID, sh.ga[0].ID, 1, "after-cancel")); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("buying a cancelled showtime: %v", err)
	}
	if _, err := p.orders.Reserve(ctx, user(8), cmd(sh.event.ID, sh.ga[1].ID, 1, "other")); err != nil {
		t.Fatalf("buying the other showtime: %v", err)
	}
	ev, _ := p.events.View(ctx, nil, sh.event.ID)
	if ev.Status != domain.EventPublished || !ev.StartsAt.Equal(sh.sessions[1].StartsAt) {
		t.Fatalf("the event goes on, now starting with its remaining showtime: %v %v", ev.Status, ev.StartsAt)
	}
	// its orders are refunded, the other showtime's are not
	if n, err := p.orders.SettleCancelledEvents(ctx); err != nil || n != 2 {
		t.Fatalf("settle: %d %v", n, err)
	}
	if orderStatus(t, p, paid0.ID) != domain.OrderRefunded || orderStatus(t, p, pend0.ID) != domain.OrderCancelled || orderStatus(t, p, paid1.ID) != domain.OrderPaid {
		t.Fatalf("paid0=%s pend0=%s paid1=%s", orderStatus(t, p, paid0.ID).Name(), orderStatus(t, p, pend0.ID).Name(), orderStatus(t, p, paid1.ID).Name())
	}
	if wallets.charged() != 1 {
		t.Fatalf("%d wallet charges standing, want only the other showtime's", wallets.charged())
	}
	for _, tk := range paid0.Tickets {
		if status(t, p, tk.ID) != domain.TicketVoid {
			t.Fatalf("a ticket of a cancelled showtime is %s", status(t, p, tk.ID).Name())
		}
	}
	if c := typeCounters(t, p, sh.ga[0].ID); c.sold != 0 || c.held != 0 || c.available != 10 {
		t.Fatalf("stock of the cancelled showtime: %+v", c)
	}
	var told int
	if err := p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 5 and kind = 'event_cancelled'`).Scan(&told); err != nil || told != 1 {
		t.Fatalf("the buyer was told %d times (%v)", told, err)
	}
	if n, _ := p.orders.SettleCancelledEvents(ctx); n != 0 {
		t.Fatalf("settling twice: %d", n)
	}
	// its ticket cannot be moved back to life, and a type can be added only to a showtime that is on
	if _, err := p.events.CreateTicketType(ctx, organizer, sh.event.ID, inbound.TicketTypeInput{Name: "Late", Price: 1, Total: 1, Active: true, SessionID: sh.sessions[0].ID}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a ticket type for a cancelled showtime: %v", err)
	}
	if _, err := p.events.UpdateSession(ctx, organizer, sh.event.ID, sh.sessions[0].ID, inbound.SessionInput{StartsAt: time.Now().Add(200 * time.Hour), EndsAt: time.Now().Add(203 * time.Hour)}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("editing a cancelled showtime: %v", err)
	}
	for _, i := range []int{0} {
		checkLedger(t, p, sh.ga[i].ID)
	}
}

// Moving, editing and removing showtimes.
func TestEditingSessions(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	sh := newShow(t, p, 10, 0, 72*time.Hour, 96*time.Hour)
	buyer := orderFor(t, p, 5, sh, 1, 1, true)
	now := time.Now()

	// moving a showtime tells its buyers and keeps the event's dates in step
	newStart := now.Add(150 * time.Hour)
	moved, err := p.events.UpdateSession(ctx, organizer, sh.event.ID, sh.sessions[1].ID, inbound.SessionInput{StartsAt: newStart, EndsAt: newStart.Add(2 * time.Hour), Label: "Saturday"})
	if err != nil || moved.Label != "Saturday" || !moved.StartsAt.Equal(newStart) {
		t.Fatalf("move: %+v %v", moved, err)
	}
	ev, _ := p.events.View(ctx, &organizer, sh.event.ID)
	if !ev.EndsAt.Equal(newStart.Add(2 * time.Hour)) {
		t.Fatalf("the event now ends when its last showtime does: %v", ev.EndsAt)
	}
	var told int
	if err := p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 5 and kind = 'showtime_moved'`).Scan(&told); err != nil || told != 1 {
		t.Fatalf("buyer told of the move %d times (%v)", told, err)
	}
	_ = buyer
	for name, in := range map[string]inbound.SessionInput{
		"ends before it starts": {StartsAt: newStart, EndsAt: newStart.Add(-time.Hour)},
		"into the past":         {StartsAt: now.Add(-5 * time.Hour), EndsAt: now.Add(-2 * time.Hour)},
		"absurdly long":         {StartsAt: newStart, EndsAt: newStart.Add(30 * 24 * time.Hour)},
	} {
		if _, err := p.events.UpdateSession(ctx, organizer, sh.event.ID, sh.sessions[1].ID, in); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := p.events.UpdateSession(ctx, user(5), sh.event.ID, sh.sessions[1].ID, inbound.SessionInput{StartsAt: newStart, EndsAt: newStart.Add(time.Hour)}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a buyer moving a showtime: %v", err)
	}
	// the event's own dates cannot be edited while it has several showtimes...
	in := inbound.EventInput{Title: ev.Title, Category: ev.Category, City: ev.City, Venue: ev.Venue, WalletID: "w", Transferable: true, ResaleCapPercent: 100,
		StartsAt: ev.StartsAt.Add(time.Hour), EndsAt: ev.EndsAt}
	if _, err := p.events.Update(ctx, organizer, sh.event.ID, in); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("editing the dates of a multi-showtime event: %v", err)
	}
	// ...but the rest of it can
	in.StartsAt, in.EndsAt, in.Title = ev.StartsAt, ev.EndsAt, "Renamed"
	if got, err := p.events.Update(ctx, organizer, sh.event.ID, in); err != nil || got.Title != "Renamed" || len(got.Sessions) != 2 {
		t.Fatalf("editing an event with two showtimes: %+v %v", got.Title, err)
	}

	// removing: not while it has ticket types, never the last one
	if err := p.events.DeleteSession(ctx, organizer, sh.event.ID, sh.sessions[1].ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleting a showtime that has ticket types: %v", err)
	}
	extra, err := p.events.CreateSession(ctx, organizer, sh.event.ID, inbound.SessionInput{StartsAt: now.Add(300 * time.Hour), EndsAt: now.Add(303 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if ev, _ = p.events.View(ctx, &organizer, sh.event.ID); !ev.EndsAt.Equal(extra.EndsAt) {
		t.Fatalf("a later showtime extends the event: %v", ev.EndsAt)
	}
	if err := p.events.DeleteSession(ctx, organizer, sh.event.ID, extra.ID); err != nil {
		t.Fatalf("deleting an empty showtime: %v", err)
	}
	if ev, _ = p.events.View(ctx, &organizer, sh.event.ID); !ev.EndsAt.Equal(newStart.Add(2 * time.Hour)) {
		t.Fatalf("the event shrinks back: %v", ev.EndsAt)
	}
	if err := p.events.DeleteSession(ctx, organizer, sh.event.ID, 12345); !errors.Is(err, domain.ErrConflict) && !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleting an unknown showtime: %v", err)
	}
	// an event with one session has its session follow its dates
	solo := newShow(t, p, 3, 0, 72*time.Hour)
	if err := p.events.DeleteSession(ctx, organizer, solo.event.ID, solo.sessions[0].ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleting the last showtime: %v", err)
	}
	buyer2 := orderFor(t, p, 6, solo, 0, 1, true)
	_ = buyer2
	sv, _ := p.events.View(ctx, &organizer, solo.event.ID)
	shift := sv.StartsAt.Add(10 * time.Hour)
	if _, err := p.events.Update(ctx, organizer, solo.event.ID, inbound.EventInput{Title: sv.Title, Category: sv.Category, City: sv.City, Venue: sv.Venue, WalletID: "w",
		Transferable: true, ResaleCapPercent: 100, StartsAt: shift, EndsAt: shift.Add(3 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	sv, _ = p.events.View(ctx, &organizer, solo.event.ID)
	if len(sv.Sessions) != 1 || !sv.Sessions[0].StartsAt.Equal(shift) {
		t.Fatalf("the only showtime must move with the event: %+v", sv.Sessions)
	}
	if err := p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 6 and kind = 'showtime_moved'`).Scan(&told); err != nil || told != 1 {
		t.Fatalf("buyer of the moved event told %d times (%v)", told, err)
	}
}

// Search lists an event once and by the showtime that is next.
func TestSearchWithSessions(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	catalog := service.NewCatalogService(p.erepo, 0)
	multi := newShow(t, p, 5, 0, 48*time.Hour, 200*time.Hour)
	later := newShow(t, p, 5, 0, 100*time.Hour)
	if _, err := p.pool.Exec(ctx, `update event set title = 'Later show' where id = $1`, later.event.ID); err != nil {
		t.Fatal(err)
	}

	list, err := catalog.Search(ctx, domain.EventFilter{Limit: 10})
	if err != nil || len(list) != 2 || list[0].ID != multi.event.ID || list[1].ID != later.event.ID {
		t.Fatalf("search: %v %v", ids(list), err)
	}
	if list[0].NextSession == nil || !list[0].NextSession.Equal(multi.sessions[0].StartsAt) {
		t.Fatalf("next showtime: %v", list[0].NextSession)
	}
	now := time.Now()
	// a date range matches if *any* showtime to come falls in it
	in := func(from, to time.Duration) []int64 {
		f, tt := now.Add(from), now.Add(to)
		l, err := catalog.Search(ctx, domain.EventFilter{From: &f, To: &tt, Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		return ids(l)
	}
	if got := in(190*time.Hour, 210*time.Hour); len(got) != 1 || got[0] != multi.event.ID {
		t.Fatalf("the second showtime of the first event: %v", got)
	}
	if got := in(60*time.Hour, 150*time.Hour); len(got) != 1 && len(got) != 2 {
		t.Fatalf("a range holding 100h: %v", got)
	}
	if got := in(60*time.Hour, 80*time.Hour); len(got) != 0 {
		t.Fatalf("a range with no showtime: %v", got)
	}
	// once the first showtime is over, the next one is what a listing shows; when none is left, the event is gone
	if _, err := moveEvent2(ctx, p, multi.sessions[0].ID, "now() - interval '6 hours'", "now() - interval '3 hours'"); err != nil {
		t.Fatal(err)
	}
	if list, _ = catalog.Search(ctx, domain.EventFilter{Query: "Anh Trai", Limit: 10}); len(list) == 0 || !list[0].NextSession.Equal(multi.sessions[1].StartsAt) {
		t.Fatalf("after the first showtime: %+v", list)
	}
	for _, s := range multi.sessions[1:] {
		if _, err := p.events.CancelSession(ctx, organizer, multi.event.ID, s.ID, "off"); err != nil {
			t.Fatal(err)
		}
	}
	for _, e := range func() []domain.Event { l, _ := catalog.Search(ctx, domain.EventFilter{Limit: 10}); return l }() {
		if e.ID == multi.event.ID {
			t.Fatal("an event with no showtime left to come is still listed")
		}
	}
}

// A copy of a multi-showtime event repeats every showtime, shifted, each with its own ticket types.
func TestDuplicatingAnEventWithSessions(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	sh := newShow(t, p, 6, 0, 72*time.Hour, 96*time.Hour)
	orderFor(t, p, 5, sh, 0, 2, true)
	if _, err := p.events.CancelSession(ctx, organizer, sh.event.ID, sh.sessions[1].ID, "no"); err != nil {
		t.Fatal(err)
	}
	third, err := p.events.CreateSession(ctx, organizer, sh.event.ID, inbound.SessionInput{StartsAt: time.Now().Add(120 * time.Hour), EndsAt: time.Now().Add(123 * time.Hour), Label: "Encore", CopyFrom: sh.sessions[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	_ = third

	start := time.Now().Add(1000 * time.Hour)
	dup, err := p.events.Duplicate(ctx, organizer, sh.event.ID, inbound.DuplicateInput{StartsAt: start, EndsAt: start.Add(3 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if len(dup.Sessions) != 2 { // the cancelled showtime is not repeated
		t.Fatalf("%d sessions in the copy", len(dup.Sessions))
	}
	if !dup.Sessions[0].StartsAt.Equal(start) || dup.Sessions[1].Label != "Encore" ||
		dup.Sessions[1].StartsAt.Sub(dup.Sessions[0].StartsAt) != third.StartsAt.Sub(sh.sessions[0].StartsAt) {
		t.Fatalf("shifted showtimes: %+v", dup.Sessions)
	}
	per := map[int64]int{}
	for _, tt := range dup.TicketTypes {
		per[tt.SessionID]++
		if tt.Sold != 0 || tt.Available != tt.Total || tt.Total != 6 {
			t.Fatalf("fresh stock: %+v", tt)
		}
	}
	if per[dup.Sessions[0].ID] != 1 || per[dup.Sessions[1].ID] != 1 {
		t.Fatalf("each showtime of the copy has its own GA: %v", per)
	}
}

// A database from before sessions is upgraded in place: every event gets its one session and everything points at it.
func TestSessionsUpgradeBackfills(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	sh := newShow(t, p, 5, 0, 72*time.Hour)
	o := orderFor(t, p, 5, sh, 0, 1, true)

	// take it back to the shape before sessions
	for _, q := range []string{
		`update ticket set session_id = null`, `update ticket_order set session_id = null`, `update ticket_type set session_id = null`,
		`delete from event_session`,
	} {
		if _, err := p.pool.Exec(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := postgres.ApplySchema(ctx, p.pool); err != nil {
			t.Fatal(err)
		}
	}
	e, err := p.events.View(ctx, &organizer, sh.event.ID)
	if err != nil || len(e.Sessions) != 1 || !e.Sessions[0].StartsAt.Equal(e.StartsAt) {
		t.Fatalf("the one session an event always had: %+v %v", e.Sessions, err)
	}
	sid := e.Sessions[0].ID
	var typeOK, orderOK, ticketOK bool
	if err := p.pool.QueryRow(ctx, `select (select session_id from ticket_type where id = $1) = $4, (select session_id from ticket_order where id = $2) = $4,
		(select session_id from ticket where id = $3) = $4`, sh.ga[0].ID, o.ID, o.Tickets[0].ID, sid).Scan(&typeOK, &orderOK, &ticketOK); err != nil || !typeOK || !orderOK || !ticketOK {
		t.Fatalf("backfill: type %v order %v ticket %v (%v)", typeOK, orderOK, ticketOK, err)
	}
	// and it works as before
	if _, err := p.gate.CheckIn(ctx, organizer, sh.event.ID, o.Tickets[0].Code); err != nil {
		t.Fatalf("scanning after the upgrade: %v", err)
	}
}
