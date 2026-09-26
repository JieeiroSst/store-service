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

// paidOrder buys and pays qty tickets of a type as user u.
func paidOrder(t testing.TB, p *pod, u int64, e domain.Event, tt domain.TicketType, qty int) domain.Order {
	t.Helper()
	ctx := context.Background()
	x, err := p.orders.Reserve(ctx, user(u), cmd(e.ID, tt.ID, qty, fmt.Sprintf("paid-%d-%d", u, time.Now().UnixNano())))
	if err != nil {
		t.Fatal(err)
	}
	if x, err = p.orders.Pay(ctx, user(u), x.ID, inbound.PayCommand{Method: domain.MethodWallet}); err != nil {
		t.Fatal(err)
	}
	return x
}

func history(t testing.TB, p *pod, actor inbound.Principal, id int64) []string {
	t.Helper()
	rows, err := p.tickets.History(context.Background(), actor, id)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, r := range rows {
		out = append(out, r.Event)
	}
	return out
}

func sameList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func wantHistory(t testing.TB, p *pod, actor inbound.Principal, id int64, want ...string) {
	t.Helper()
	if got := history(t, p, actor, id); !sameList(got, want) {
		t.Fatalf("ticket %d history = %v, want %v", id, got, want)
	}
}

func status(t testing.TB, p *pod, id int64) domain.TicketStatus {
	t.Helper()
	tk, err := p.trepo.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return tk.Status
}

// A ticket's whole life: issued, given an attendee, re-coded, offered, accepted, declined, cancelled, expired,
// scanned, un-scanned, revoked and finally expired with its event: every step allowed only from the right state, and
// every step in its history.
func TestTicketLifecycle(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 50, nil)

	o := paidOrder(t, p, 5, e, tt, 2)
	a, b := o.Tickets[0], o.Tickets[1]
	if a.HolderID != 5 || a.Status != domain.TicketValid || a.HolderEmail != "buyer@example.com" {
		t.Fatalf("a new ticket: %+v", a)
	}
	wantHistory(t, p, user(5), a.ID, "issued")

	// only the holder edits the attendee; a stranger does not even see the ticket
	if _, err := p.tickets.SetHolder(ctx, user(5), a.ID, inbound.HolderInput{Name: "Tran Thi B", Email: "b@example.com"}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.tickets.SetHolder(ctx, user(5), a.ID, inbound.HolderInput{Email: "not-an-email"}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("bad e-mail: %v", err)
	}
	if _, err := p.tickets.SetHolder(ctx, user(6), a.ID, inbound.HolderInput{Name: "x"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger editing a ticket: %v", err)
	}
	if _, err := p.tickets.SetHolder(ctx, organizer, a.ID, inbound.HolderInput{Name: "x"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("the organizer editing a ticket's attendee: %v", err)
	}
	if got, err := p.tickets.Get(ctx, organizer, a.ID); err != nil || got.HolderName != "Tran Thi B" {
		t.Fatalf("the organizer reading a ticket: %+v %v", got, err)
	}

	// a leaked QR code is replaced: the old one stops working at once
	old := a.Code
	re, err := p.tickets.Reissue(ctx, user(5), a.ID)
	if err != nil || re.Code == old || re.Status != domain.TicketValid {
		t.Fatalf("reissue: %+v %v", re, err)
	}
	a = re
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, old); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("scanning the replaced code: %v", err)
	}

	// ---- transfer to a user id, accepted
	if _, err := p.tickets.Offer(ctx, user(5), a.ID, inbound.TransferCommand{ToUserID: 5}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("transfer to yourself: %v", err)
	}
	if _, err := p.tickets.Offer(ctx, user(5), a.ID, inbound.TransferCommand{}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("transfer to nobody: %v", err)
	}
	if _, err := p.tickets.Offer(ctx, user(6), a.ID, inbound.TransferCommand{ToUserID: 7}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("offering someone else's ticket: %v", err)
	}
	x, err := p.tickets.Offer(ctx, user(5), a.ID, inbound.TransferCommand{ToUserID: 6, Message: "enjoy!"})
	if err != nil || x.Status != domain.TransferPending {
		t.Fatalf("offer: %+v %v", x, err)
	}
	if status(t, p, a.ID) != domain.TicketTransferring {
		t.Fatal("an offered ticket must be transferring")
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, a.Code); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("scanning a ticket that is on offer: %v", err)
	}
	if _, err := p.tickets.Offer(ctx, user(5), a.ID, inbound.TransferCommand{ToUserID: 8}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a second offer of the same ticket: %v", err)
	}
	if _, err := p.tickets.Reissue(ctx, user(5), a.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("re-coding a ticket that is on offer: %v", err)
	}
	in6, err := p.tickets.Transfers(ctx, user(6), true, 0, 10)
	if err != nil || len(in6.Items) != 1 || in6.Items[0].Ticket == nil {
		t.Fatalf("incoming offers of the recipient: %+v %v", in6, err)
	}
	if in7, _ := p.tickets.Transfers(ctx, user(7), true, 0, 10); len(in7.Items) != 0 {
		t.Fatal("someone else sees the offer")
	}
	if _, err := p.tickets.AcceptTransfer(ctx, user(7), x.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("accepting an offer addressed to someone else: %v", err)
	}
	if _, err := p.tickets.AcceptTransfer(ctx, user(5), x.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("accepting your own offer: %v", err)
	}
	got, err := p.tickets.AcceptTransfer(ctx, user(6), x.ID)
	if err != nil || got.HolderID != 6 || got.Status != domain.TicketValid || got.Code == a.Code || got.TransferCount != 1 {
		t.Fatalf("accept: %+v %v", got, err)
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, a.Code); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the sender's old code still works: %v", err)
	}
	if _, err := p.tickets.Get(ctx, user(5), a.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the sender still sees the ticket: %v", err)
	}
	if mine, _ := p.tickets.List(ctx, user(5), inbound.TicketListQuery{}); len(mine.Items) != 1 || mine.Items[0].ID != b.ID {
		t.Fatalf("the sender's tickets: %+v", mine.Items)
	}
	if mine, _ := p.tickets.List(ctx, user(6), inbound.TicketListQuery{Status: domain.TicketValid}); len(mine.Items) != 1 || mine.Items[0].EventTitle == "" {
		t.Fatalf("the recipient's tickets: %+v", mine.Items)
	}
	if _, err := p.tickets.AcceptTransfer(ctx, user(6), x.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("accepting twice: %v", err)
	}
	wantHistory(t, p, user(6), a.ID, "issued", "holder_changed", "reissued", "transfer_offered", "transfer_accepted")
	a = got

	// ---- decline, cancel, and the limit of two hand-overs
	x, _ = p.tickets.Offer(ctx, user(6), a.ID, inbound.TransferCommand{ToUserID: 9})
	if _, err := p.tickets.CancelTransfer(ctx, user(9), x.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the recipient cancelling: %v", err)
	}
	if _, err := p.tickets.DeclineTransfer(ctx, user(6), x.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the sender declining: %v", err)
	}
	if x, err = p.tickets.DeclineTransfer(ctx, user(9), x.ID); err != nil || x.Status != domain.TransferDeclined || status(t, p, a.ID) != domain.TicketValid {
		t.Fatalf("decline: %+v %v", x, err)
	}
	x, _ = p.tickets.Offer(ctx, user(6), a.ID, inbound.TransferCommand{ToEmail: "Friend@Example.com"})
	if x, err = p.tickets.CancelTransfer(ctx, user(6), x.ID); err != nil || x.Status != domain.TransferCancelled || status(t, p, a.ID) != domain.TicketValid {
		t.Fatalf("cancel: %+v %v", x, err)
	}
	// by e-mail: whoever signs in with that address (any case) may accept
	x, _ = p.tickets.Offer(ctx, user(6), a.ID, inbound.TransferCommand{ToEmail: "friend@example.com"})
	if _, err := p.tickets.AcceptTransfer(ctx, inbound.Principal{UserID: 10, Email: "other@example.com"}, x.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("accepting with the wrong e-mail: %v", err)
	}
	if got, err = p.tickets.AcceptTransfer(ctx, inbound.Principal{UserID: 11, Email: "FRIEND@example.com"}, x.ID); err != nil || got.HolderID != 11 || got.TransferCount != 2 {
		t.Fatalf("accept by e-mail: %+v %v", got, err)
	}
	if _, err := p.tickets.Offer(ctx, user(11), a.ID, inbound.TransferCommand{ToUserID: 12}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a third hand-over (the limit is 2): %v", err)
	}

	// ---- an offer nobody answers lapses
	x, _ = p.tickets.Offer(ctx, user(5), b.ID, inbound.TransferCommand{ToUserID: 6})
	if _, err := p.pool.Exec(ctx, `update ticket_transfer set expires_at = now() - interval '1 minute' where id = $1`, x.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := p.tickets.AcceptTransfer(ctx, user(6), x.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("accepting an expired offer: %v", err)
	}
	if n, err := p.tickets.ExpireTransfers(ctx); err != nil || n != 1 {
		t.Fatalf("expire transfers: %d %v", n, err)
	}
	if status(t, p, b.ID) != domain.TicketValid {
		t.Fatal("an unanswered offer must give the ticket back")
	}
	wantHistory(t, p, user(5), b.ID, "issued", "transfer_offered", "transfer_expired")

	// ---- the gate: scan, revert, revoke
	if _, err := p.tickets.Void(ctx, user(5), b.ID, "x"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a holder revoking their own ticket: %v", err)
	}
	if _, err := p.tickets.Void(ctx, organizer, b.ID, " "); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("revoking without a reason: %v", err)
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, b.Code); err != nil {
		t.Fatal(err)
	}
	if status(t, p, b.ID) != domain.TicketUsed {
		t.Fatal("a scanned ticket is used")
	}
	if _, err := p.tickets.Void(ctx, organizer, b.ID, "fraud"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("revoking a ticket already used: %v", err)
	}
	if _, err := p.tickets.Offer(ctx, user(5), b.ID, inbound.TransferCommand{ToUserID: 6}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("transferring a used ticket: %v", err)
	}
	if _, err := p.tickets.RevertCheckIn(ctx, user(5), b.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a holder undoing their own scan: %v", err)
	}
	if r, err := p.tickets.RevertCheckIn(ctx, organizer, b.ID); err != nil || r.Status != domain.TicketValid || r.CheckedInAt != nil {
		t.Fatalf("revert: %+v %v", r, err)
	}
	if _, err := p.tickets.RevertCheckIn(ctx, organizer, b.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("reverting a scan that was not made: %v", err)
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, b.Code); err != nil { // scan again, stays used
		t.Fatal(err)
	}
	c := paidOrder(t, p, 12, e, tt, 1).Tickets[0]
	if v, err := p.tickets.Void(ctx, organizer, c.ID, "chargeback"); err != nil || v.Status != domain.TicketVoid {
		t.Fatalf("void: %+v %v", v, err)
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, c.Code); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("scanning a revoked ticket: %v", err)
	}
	wantHistory(t, p, organizer, c.ID, "issued", "voided")
	var told int
	if err := p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 12 and kind = 'ticket_voided'`).Scan(&told); err != nil || told != 1 {
		t.Fatalf("the holder of a revoked ticket was told %d times (%v)", told, err)
	}

	// ---- refunding an order voids its tickets and calls off their transfers
	d1 := paidOrder(t, p, 13, e, tt, 2)
	off, err := p.tickets.Offer(ctx, user(13), d1.Tickets[0].ID, inbound.TransferCommand{ToUserID: 14})
	if err != nil {
		t.Fatal(err)
	}
	if r, err := p.orders.Cancel(ctx, organizer, d1.ID); err != nil || r.Status != domain.OrderRefunded {
		t.Fatalf("organizer refunds the order: %+v %v", r, err)
	}
	for _, tk := range d1.Tickets {
		if status(t, p, tk.ID) != domain.TicketVoid {
			t.Fatalf("ticket %d of a refunded order is %s", tk.ID, status(t, p, tk.ID).Name())
		}
	}
	if got, _ := p.trepo.GetTransfer(ctx, off.ID); got.Status != domain.TransferCancelled {
		t.Fatalf("the offer of a refunded ticket is %s", got.Status.Name())
	}
	if _, err := p.tickets.AcceptTransfer(ctx, user(14), off.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("accepting the offer of a refunded ticket: %v", err)
	}

	// ---- the event ends: what was not used expires, what was used stays used
	if _, err := moveEvent(ctx, p, e.ID, `now() - interval '10 hours'`, `now() - interval '7 hours'`); err != nil {
		t.Fatal(err)
	}
	if n, err := p.tickets.ExpireTickets(ctx); err != nil || n != 1 {
		t.Fatalf("expire tickets: %d %v", n, err)
	}
	if status(t, p, a.ID) != domain.TicketExpired || status(t, p, b.ID) != domain.TicketUsed || status(t, p, c.ID) != domain.TicketVoid {
		t.Fatalf("after the event: a=%s b=%s c=%s", status(t, p, a.ID).Name(), status(t, p, b.ID).Name(), status(t, p, c.ID).Name())
	}
	if _, err := p.tickets.Offer(ctx, user(11), a.ID, inbound.TransferCommand{ToUserID: 12}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("transferring an expired ticket: %v", err)
	}
	h := history(t, p, user(11), a.ID)
	if h[len(h)-1] != "expired" {
		t.Fatalf("history of an expired ticket: %v", h)
	}
	checkLedger(t, p, tt.ID)
}

// An organizer can forbid transfers for an event.
func TestTransfersCanBeForbidden(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 5, nil)
	in := inbound.EventInput{Title: e.Title, Category: e.Category, City: e.City, Venue: e.Venue, StartsAt: e.StartsAt, EndsAt: e.EndsAt, WalletID: "organizer-wallet", Transferable: false}
	if _, err := p.events.Update(ctx, organizer, e.ID, in); err != nil {
		t.Fatal(err)
	}
	o := paidOrder(t, p, 5, e, tt, 1)
	if _, err := p.tickets.Offer(ctx, user(5), o.Tickets[0].ID, inbound.TransferCommand{ToUserID: 6}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("transferring at an event that forbids it: %v", err)
	}
}

// Many hands reaching for the same offer: exactly one wins, and nothing is left half-done.
func TestTransferRaces(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	pods := newPods(t, d, 3, podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, pods[0], 200, nil)

	for round := 0; round < 20; round++ {
		o := paidOrder(t, pods[0], 5, e, tt, 1)
		tk := o.Tickets[0]
		x, err := pods[0].tickets.Offer(ctx, user(5), tk.ID, inbound.TransferCommand{ToUserID: 6})
		if err != nil {
			t.Fatal(err)
		}
		var accepted, other atomic.Int32
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p := pods[i%len(pods)]
				switch i % 3 {
				case 0, 1: // the recipient, from several devices
					if _, err := p.tickets.AcceptTransfer(ctx, user(6), x.ID); err == nil {
						accepted.Add(1)
					} else if !errors.Is(err, domain.ErrConflict) {
						t.Errorf("accept: %v", err)
					}
				case 2: // the sender changes their mind
					if _, err := p.tickets.CancelTransfer(ctx, user(5), x.ID); err == nil {
						other.Add(1)
					} else if !errors.Is(err, domain.ErrConflict) {
						t.Errorf("cancel: %v", err)
					}
				}
			}()
		}
		wg.Wait()
		if accepted.Load()+other.Load() != 1 {
			t.Fatalf("round %d: %d accepts and %d cancels succeeded; exactly one may", round, accepted.Load(), other.Load())
		}
		final, _ := pods[0].trepo.Get(ctx, tk.ID)
		switch {
		case accepted.Load() == 1 && (final.HolderID != 6 || final.Status != domain.TicketValid || final.TransferCount != 1):
			t.Fatalf("round %d: accepted but the ticket is %+v", round, final)
		case other.Load() == 1 && (final.HolderID != 5 || final.Status != domain.TicketValid || final.TransferCount != 0):
			t.Fatalf("round %d: cancelled but the ticket is %+v", round, final)
		}
	}
	var stuck int
	if err := pods[0].pool.QueryRow(ctx, `select count(*) from ticket t where status = 5 and not exists
		(select 1 from ticket_transfer x where x.ticket_id = t.id and x.status = 1)`).Scan(&stuck); err != nil || stuck != 0 {
		t.Fatalf("%d tickets are transferring with no open offer (%v)", stuck, err)
	}
	checkLedger(t, pods[0], tt.ID)
}

// The same code scanned by many gates at once admits exactly one person.
func TestConcurrentScansAdmitOnce(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	pods := newPods(t, d, 3, podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, pods[0], 10, nil)
	tk := paidOrder(t, pods[0], 5, e, tt, 1).Tickets[0]

	var admitted, again atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := pods[i%len(pods)].gate.CheckIn(ctx, organizer, e.ID, tk.Code)
			switch {
			case err == nil:
				admitted.Add(1)
			case errors.Is(err, domain.ErrAlreadyCheckedIn):
				again.Add(1)
			default:
				t.Errorf("scan: %v", err)
			}
		}()
	}
	wg.Wait()
	if admitted.Load() != 1 || again.Load() != 199 {
		t.Fatalf("%d admitted, %d turned away as already used", admitted.Load(), again.Load())
	}
	wantHistory(t, pods[0], organizer, tk.ID, "issued", "checked_in")
}

// Gate staff scan but do not manage; offline scans are uploaded in a batch.
func TestGateStaffAndBatchScans(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 20, nil)
	o := paidOrder(t, p, 5, e, tt, 5)
	staff, outsider := user(40), user(41)

	if _, err := p.gate.CheckIn(ctx, staff, e.ID, o.Tickets[0].Code); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a stranger scanning: %v", err)
	}
	if err := p.staff.Add(ctx, outsider, e.ID, 40); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a stranger naming staff: %v", err)
	}
	if err := p.staff.Add(ctx, organizer, e.ID, organizer.UserID); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("the organizer as staff: %v", err)
	}
	if err := p.staff.Add(ctx, organizer, e.ID, 40); err != nil {
		t.Fatal(err)
	}
	if err := p.staff.Add(ctx, organizer, e.ID, 40); err != nil { // adding twice is harmless
		t.Fatal(err)
	}
	if list, err := p.staff.List(ctx, organizer, e.ID); err != nil || len(list) != 1 {
		t.Fatalf("staff list: %v %v", list, err)
	}
	if evs, err := p.staff.Events(ctx, staff); err != nil || len(evs) != 1 || evs[0].ID != e.ID {
		t.Fatalf("the staff member's events: %v %v", evs, err)
	}
	if _, err := p.gate.CheckIn(ctx, staff, e.ID, o.Tickets[0].Code); err != nil {
		t.Fatalf("staff scanning: %v", err)
	}
	if got, err := p.tickets.Lookup(ctx, staff, e.ID, " "+lowerString(o.Tickets[1].Code)+" "); err != nil || got.ID != o.Tickets[1].ID {
		t.Fatalf("staff looking a ticket up (any case, with spaces): %+v %v", got, err)
	}
	if _, err := p.tickets.Void(ctx, staff, o.Tickets[1].ID, "x"); !errors.Is(err, domain.ErrForbidden) && !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("staff revoking a ticket: %v", err)
	}
	if _, err := p.events.Report(ctx, staff, e.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("staff reading the sales report: %v", err)
	}
	if _, err := p.tickets.Lookup(ctx, staff, 0, o.Tickets[1].Code); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("staff searching every event: %v", err)
	}
	if got, err := p.tickets.Lookup(ctx, admin, 0, o.Tickets[1].Code); err != nil || got.ID != o.Tickets[1].ID {
		t.Fatalf("an admin searching every event: %+v %v", got, err)
	}

	// an offline gate uploads its scans: some good, one repeated, one unknown, one revoked, one transferring
	if _, err := p.tickets.Void(ctx, organizer, o.Tickets[3].ID, "lost phone"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.tickets.Offer(ctx, user(5), o.Tickets[4].ID, inbound.TransferCommand{ToUserID: 6}); err != nil {
		t.Fatal(err)
	}
	if _, err := moveEvent(ctx, p, e.ID, `now() - interval '3 hours'`, `now() + interval '2 hours'`); err != nil {
		t.Fatal(err) // the event is under way
	}
	scanned := time.Now().Add(-90 * time.Minute).Truncate(time.Second)
	res, err := p.gate.BatchCheckIn(ctx, staff, e.ID, []inbound.ScanItem{
		{Code: o.Tickets[1].Code, ScannedAt: &scanned},
		{Code: o.Tickets[0].Code}, // already scanned above
		{Code: "NOPE"},
		{Code: o.Tickets[3].Code},
		{Code: o.Tickets[4].Code},
		{Code: o.Tickets[2].Code, ScannedAt: ptrTime(time.Now().Add(time.Hour))}, // "from the future": counts as now
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"admitted", "already_used", "not_found", "void", "transferring", "admitted"}
	for i, w := range want {
		if res[i].Result != w {
			t.Errorf("scan %d = %q, want %q", i, res[i].Result, w)
		}
	}
	tk, _ := p.trepo.Get(ctx, o.Tickets[1].ID)
	if tk.CheckedInAt == nil || !tk.CheckedInAt.Equal(scanned) {
		t.Fatalf("the offline scan time was not kept: %v (want %v)", tk.CheckedInAt, scanned)
	}
	if tk, _ := p.trepo.Get(ctx, o.Tickets[2].ID); tk.CheckedInAt == nil || time.Since(*tk.CheckedInAt) > time.Minute {
		t.Fatalf("a scan time in the future must not be believed: %v", tk.CheckedInAt)
	}
	if _, err := p.gate.BatchCheckIn(ctx, staff, e.ID, nil); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("an empty batch: %v", err)
	}

	// staff removed: no more scanning
	if err := p.staff.Remove(ctx, organizer, e.ID, 40); err != nil {
		t.Fatal(err)
	}
	if _, err := p.gate.CheckIn(ctx, staff, e.ID, o.Tickets[0].Code); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("removed staff scanning: %v", err)
	}

	// after the event, or if it was cancelled, the gate is closed
	if _, err := moveEvent(ctx, p, e.ID, `now() - interval '5 hours'`, `now() - interval '1 hour'`); err != nil {
		t.Fatal(err)
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, o.Tickets[0].Code); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("scanning after the event: %v", err)
	}
	res, err = p.gate.BatchCheckIn(ctx, organizer, e.ID, []inbound.ScanItem{{Code: o.Tickets[0].Code, ScannedAt: ptrTime(time.Now().Add(-90 * time.Minute))}})
	if err != nil || res[0].Result != "already_used" {
		t.Fatalf("an offline scan from during the event, uploaded soon after it ended: %+v %v", res, err)
	}
}

func lowerString(s string) string {
	b := []byte(s)
	for i, c := range b {
		if 'A' <= c && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func ptrTime(t time.Time) *time.Time { return &t }

// Invitations are free tickets from the same stock, outside the buyer's limits.
func TestInvitations(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	future := time.Now().Add(24 * time.Hour)
	e, tt := seedEvent(t, p, 4, func(in *inbound.TicketTypeInput) {
		in.MaxPerUser = 1
		in.SaleStartsAt = &future // not on sale yet
	})
	inv := inbound.InviteCommand{TicketTypeID: tt.ID, Quantity: 3, UserID: 50, Name: "Guest of Honour", Email: "vip@example.com", Note: "See you there"}

	if _, err := p.orders.Invite(ctx, user(5), e.ID, inv); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a buyer inviting: %v", err)
	}
	if _, err := p.orders.Reserve(ctx, user(5), cmd(e.ID, tt.ID, 1, "early")); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a buyer before the sale opens: %v", err)
	}
	bad := inv
	bad.Email = "nope"
	if _, err := p.orders.Invite(ctx, organizer, e.ID, bad); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("a bad invitation: %v", err)
	}
	x, err := p.orders.Invite(ctx, organizer, e.ID, inv) // 3 tickets although max_per_user is 1 and sales are closed
	if err != nil || x.Status != domain.OrderPaid || x.Total != 0 || len(x.Tickets) != 3 || x.UserID != 50 {
		t.Fatalf("invite: %+v %v", x, err)
	}
	if got := typeCounters(t, p, tt.ID); got.sold != 3 || got.available != 1 {
		t.Fatalf("stock after inviting: %+v", got)
	}
	mine, _ := p.tickets.List(ctx, user(50), inbound.TicketListQuery{})
	if len(mine.Items) != 3 || mine.Items[0].HolderID != 50 {
		t.Fatalf("the invited person's tickets: %+v", mine.Items)
	}
	var told int
	if err := p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 50 and kind = 'ticket_invited'`).Scan(&told); err != nil || told != 1 {
		t.Fatalf("invited person notified %d times (%v)", told, err)
	}
	if _, err := p.orders.Invite(ctx, organizer, e.ID, inbound.InviteCommand{TicketTypeID: tt.ID, Quantity: 2, UserID: 51, Name: "B", Email: "b@example.com"}); !errors.Is(err, domain.ErrSoldOut) {
		t.Fatalf("inviting beyond the stock: %v", err)
	}
	checkLedger(t, p, tt.ID)

	// sales figures: invitations are counted apart and earn nothing; the paid day is reported
	if _, err := p.pool.Exec(ctx, `update ticket_type set sale_starts_at = null where id = $1`, tt.ID); err != nil {
		t.Fatal(err)
	}
	paidOrder(t, p, 5, e, tt, 1)
	rep, err := p.events.Report(ctx, organizer, e.ID)
	if err != nil || rep.TicketsSold != 4 || rep.Invited != 3 || rep.Revenue != 500_000 || len(rep.Daily) != 1 {
		t.Fatalf("report: %+v %v", rep, err)
	}
	if d0 := rep.Daily[0]; d0.Revenue != 500_000 || d0.Tickets != 1 || d0.Orders != 1 {
		t.Fatalf("daily: %+v", d0)
	}
}

// "Tell me when tickets are back".
func TestWaitlist(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 1, nil)

	if err := p.waitlist.Join(ctx, user(31), e.ID, tt.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("waiting for tickets that are on sale: %v", err)
	}
	o, err := p.orders.Reserve(ctx, user(30), cmd(e.ID, tt.ID, 1, "w30"))
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range []int64{31, 32} {
		if err := p.waitlist.Join(ctx, user(u), e.ID, tt.ID); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.waitlist.Join(ctx, user(31), e.ID, tt.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("joining twice: %v", err)
	}
	if err := p.waitlist.Join(ctx, user(33), e.ID+999, tt.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("joining the list of another event: %v", err)
	}
	if mine, err := p.waitlist.Mine(ctx, user(31)); err != nil || len(mine) != 1 || mine[0].EventTitle == "" {
		t.Fatalf("my waiting lists: %+v %v", mine, err)
	}
	if err := p.waitlist.Leave(ctx, user(32), tt.ID); err != nil {
		t.Fatal(err)
	}

	count := func(u int64) int {
		var n int
		if err := p.pool.QueryRow(ctx, `select count(*) from notification where user_id = $1 and kind = 'tickets_available'`, u).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if _, err := p.orders.Cancel(ctx, user(30), o.ID); err != nil { // the ticket comes back
		t.Fatal(err)
	}
	if count(31) != 1 || count(32) != 0 || count(30) != 0 {
		t.Fatalf("notified: 31=%d 32=%d 30=%d, want 1 0 0", count(31), count(32), count(30))
	}
	if mine, _ := p.waitlist.Mine(ctx, user(31)); len(mine) != 0 {
		t.Fatal("a notified person must leave the list")
	}

	// sold out again; the organizer adds tickets: the new waiter hears about it
	if _, err := p.orders.Reserve(ctx, user(34), cmd(e.ID, tt.ID, 1, "w34")); err != nil {
		t.Fatal(err)
	}
	if err := p.waitlist.Join(ctx, user(35), e.ID, tt.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := p.events.UpdateTicketType(ctx, organizer, e.ID, tt.ID, inbound.TicketTypeInput{Name: "GA", Price: 500_000, Total: 5, MaxPerOrder: 10, Active: true}); err != nil {
		t.Fatal(err)
	}
	if count(35) != 1 {
		t.Fatalf("waiter told when the organizer added tickets: %d", count(35))
	}
}
