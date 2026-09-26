package integration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

// listed sets up an event, a ticket of user 5 (face value 500,000) and a listing of it for the given price.
func listed(t testing.TB, p *pod, price int64) (domain.Event, domain.Ticket, domain.ResaleListing) {
	t.Helper()
	e, tt := seedEvent(t, p, 20, nil)
	tk := paidOrder(t, p, 5, e, tt, 1).Tickets[0]
	l, err := p.resale.List(context.Background(), user(5), tk.ID, price)
	if err != nil {
		t.Fatal(err)
	}
	return e, tk, l
}

func listingStatus(t testing.TB, p *pod, id int64) domain.ResaleStatus {
	t.Helper()
	l, err := p.resale.Mine(context.Background(), user(5), 0, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range l.Items {
		if x.ID == id {
			return x.Status
		}
	}
	t.Fatalf("listing %d not found", id)
	return 0
}

func TestResale(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 20, nil)
	tk := paidOrder(t, p, 5, e, tt, 1).Tickets[0]

	// what may be listed, and for how much
	if _, err := p.resale.List(ctx, user(5), tk.ID, 500_001); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("above the face value: %v", err)
	}
	if _, err := p.resale.List(ctx, user(5), tk.ID, 0); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("a free listing: %v", err)
	}
	if _, err := p.resale.List(ctx, user(6), tk.ID, 100_000); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("listing somebody else's ticket: %v", err)
	}
	in := inbound.EventInput{Title: e.Title, Category: e.Category, City: e.City, Venue: e.Venue, StartsAt: e.StartsAt, EndsAt: e.EndsAt,
		WalletID: "organizer-wallet", Transferable: true, ResaleCapPercent: 0}
	if _, err := p.events.Update(ctx, organizer, e.ID, in); err != nil {
		t.Fatal(err)
	}
	if _, err := p.resale.List(ctx, user(5), tk.ID, 100_000); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("resale the organizer forbids: %v", err)
	}
	in.ResaleCapPercent = 80
	if _, err := p.events.Update(ctx, organizer, e.ID, in); err != nil {
		t.Fatal(err)
	}
	if _, err := p.resale.List(ctx, user(5), tk.ID, 400_001); !errors.Is(err, domain.ErrInvalid) { // 80% of 500,000
		t.Fatalf("above the organizer's ceiling: %v", err)
	}

	l, err := p.resale.List(ctx, user(5), tk.ID, 400_000)
	if err != nil || l.Status != domain.ResaleOpen || l.FacePrice != 500_000 || l.Price != 400_000 || l.Currency != "VND" {
		t.Fatalf("listing: %+v %v", l, err)
	}
	if got, _ := p.trepo.Get(ctx, tk.ID); got.Status != domain.TicketListed {
		t.Fatalf("a listed ticket is %s", got.Status.Name())
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, tk.Code); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("scanning a ticket on sale: %v", err)
	}
	if _, err := p.tickets.Offer(ctx, user(5), tk.ID, inbound.TransferCommand{ToUserID: 8}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("transferring a ticket on sale: %v", err)
	}
	if _, err := p.resale.List(ctx, user(5), tk.ID, 300_000); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("listing twice: %v", err)
	}
	if _, err := p.tickets.Reissue(ctx, user(5), tk.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("re-coding a ticket on sale: %v", err)
	}

	page, err := p.resale.Browse(ctx, inbound.ResaleQuery{EventID: e.ID, Limit: 10})
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != l.ID || page.Items[0].TypeName != "GA" {
		t.Fatalf("browse: %+v %v", page, err)
	}
	if _, err := p.resale.Buy(ctx, user(5), l.ID); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("buying your own ticket: %v", err)
	}
	// the buyer of the original order cannot get the money back for a ticket they have put on sale
	if _, err := p.orders.Cancel(ctx, user(5), tk.OrderID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("refunding an order whose ticket is on sale: %v", err)
	}

	bought, err := p.resale.Buy(ctx, user(6), l.ID)
	if err != nil {
		t.Fatal(err)
	}
	if bought.HolderID != 6 || bought.Status != domain.TicketValid || bought.Code == tk.Code || bought.TransferCount != 1 {
		t.Fatalf("bought ticket: %+v", bought)
	}
	// 400,000: 40,000 (10%) to the platform, 360,000 to the seller (who had paid 500,000 for the ticket), all from the buyer
	if p.wallets.balance("w-5") != 360_000-500_000 || p.wallets.balance("platform") != 40_000 {
		t.Fatalf("seller %d platform %d", p.wallets.balance("w-5"), p.wallets.balance("platform"))
	}
	if p.wallets.balance("w-6") != -400_000 {
		t.Fatalf("buyer paid %d", -p.wallets.balance("w-6"))
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, tk.Code); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the seller's old code still works: %v", err)
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, bought.Code); err != nil {
		t.Fatalf("the buyer's code: %v", err)
	}
	wantHistory(t, p, user(6), tk.ID, "issued", "resale_listed", "resold", "checked_in")
	if _, err := p.resale.Buy(ctx, user(7), l.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("buying a sold ticket: %v", err)
	}
	if _, err := p.resale.Cancel(ctx, user(5), l.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("cancelling a sold listing: %v", err)
	}
	if listingStatus(t, p, l.ID) != domain.ResaleSold {
		t.Fatal("the listing must be sold")
	}
	// the seller's order can no longer be refunded by the seller: the ticket is gone
	if _, err := p.orders.Cancel(ctx, user(5), tk.OrderID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("refunding an order whose ticket was resold: %v", err)
	}
	var told int
	if err := p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 5 and kind = 'resale_sold'`).Scan(&told); err != nil || told != 1 {
		t.Fatalf("seller notified %d times (%v)", told, err)
	}
	// the resold ticket counts as one hand-over: the buyer can pass it on once more, then no more
	if _, err := p.tickets.Get(ctx, user(6), tk.ID); err != nil {
		t.Fatal(err)
	}
}

func TestResaleCancelAndLimits(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	_, tk, l := listed(t, p, 300_000)

	if _, err := p.resale.Cancel(ctx, user(6), l.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("cancelling somebody else's listing: %v", err)
	}
	if c, err := p.resale.Cancel(ctx, user(5), l.ID); err != nil || c.Status != domain.ResaleCancelled {
		t.Fatalf("cancel: %+v %v", c, err)
	}
	if got, _ := p.trepo.Get(ctx, tk.ID); got.Status != domain.TicketValid {
		t.Fatalf("a ticket taken off sale is %s", got.Status.Name())
	}
	if _, err := p.resale.Buy(ctx, user(6), l.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("buying a cancelled listing: %v", err)
	}
	// the hand-over limit (2): sell it on twice, a third sale is refused
	var seller int64 = 5
	for buyer := int64(6); buyer <= 7; buyer++ {
		ll, err := p.resale.List(ctx, user(seller), tk.ID, 100_000)
		if err != nil {
			t.Fatalf("list by %d: %v", seller, err)
		}
		if _, err := p.resale.Buy(ctx, user(buyer), ll.ID); err != nil {
			t.Fatal(err)
		}
		seller = buyer
	}
	if _, err := p.resale.List(ctx, user(7), tk.ID, 100_000); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a third hand-over: %v", err)
	}
	if page, _ := p.resale.Mine(ctx, user(5), domain.ResaleSold, 0, 10); len(page.Items) != 1 {
		t.Fatalf("my sold listings: %+v", page.Items)
	}
}

// Many buyers, one listing: exactly one pays and gets the ticket.
func TestResaleRace(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	wallets := newFakeWallets()
	pods := []*pod{newPod(t, d, wallets, podOptions{ratePerSec: 1000}), newPod(t, d, wallets, podOptions{ratePerSec: 1000}), newPod(t, d, wallets, podOptions{ratePerSec: 1000})}
	_, tk, l := listed(t, pods[0], 450_000)

	var won, lost atomic.Int32
	winner := make(chan int64, 100)
	var wg sync.WaitGroup
	for u := int64(100); u < 160; u++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := pods[int(u)%len(pods)].resale.Buy(ctx, user(u), l.ID)
			switch {
			case err == nil:
				won.Add(1)
				winner <- u
			case errors.Is(err, domain.ErrConflict):
				lost.Add(1)
			default:
				t.Errorf("buyer %d: %v", u, err)
			}
		}()
	}
	wg.Wait()
	if won.Load() != 1 || lost.Load() != 59 {
		t.Fatalf("%d bought, %d turned away", won.Load(), lost.Load())
	}
	w := <-winner
	got, _ := pods[0].trepo.Get(ctx, tk.ID)
	if got.HolderID != w || got.Status != domain.TicketValid {
		t.Fatalf("the ticket is held by %d (%s), the winner is %d", got.HolderID, got.Status.Name(), w)
	}
	// 500,000 paid for the ticket by user 5 + exactly one resale: 405,000 + 45,000 out of the winner's wallet
	if b := wallets.balance(fmt.Sprintf("w-%d", w)); b != -450_000 {
		t.Fatalf("the winner paid %d", -b)
	}
	if wallets.charged() != 1+2 { // the first buyer's payment for the ticket, and the two transfers of the resale
		t.Fatalf("%d charges standing: a losing buyer was charged", wallets.charged())
	}
}

// Whatever goes wrong while paying, nobody is left charged and the listing goes back on the market.
func TestResaleFailuresLeaveNoTrace(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	wallets := newFakeWallets()
	p := newPod(t, d, wallets, podOptions{ratePerSec: 1000})
	_, tk, l := listed(t, p, 400_000)
	base := wallets.charged()

	settled := func(what string) {
		t.Helper()
		if wallets.charged() != base {
			t.Fatalf("%s: %d charges standing, want %d", what, wallets.charged(), base)
		}
		if listingStatus(t, p, l.ID) != domain.ResaleOpen {
			t.Fatalf("%s: the listing is %s, want open", what, listingStatus(t, p, l.ID).Name())
		}
		if got, _ := p.trepo.Get(ctx, tk.ID); got.Status != domain.TicketListed || got.HolderID != 5 {
			t.Fatalf("%s: the ticket is %+v", what, got)
		}
	}

	wallets.failFrom["w-7"] = true // a buyer without money
	if _, err := p.resale.Buy(ctx, user(7), l.ID); !errors.Is(err, domain.ErrInsufficientFunds) {
		t.Fatalf("a buyer without funds: %v", err)
	}
	settled("no funds")
	delete(wallets.failFrom, "w-7")

	wallets.failTo["platform"] = true // the fee cannot be collected: the seller's part is taken back
	if _, err := p.resale.Buy(ctx, user(8), l.ID); err == nil {
		t.Fatal("a sale whose fee failed went through")
	}
	settled("fee failed")
	delete(wallets.failTo, "platform")

	// the ticket is revoked while the buyer is paying
	var once sync.Once
	wallets.onTransfer = func() {
		once.Do(func() {
			if _, err := p.tickets.Void(ctx, organizer, tk.ID, "fraud"); err != nil {
				t.Errorf("void: %v", err)
			}
		})
	}
	_, err := p.resale.Buy(ctx, user(9), l.ID)
	wallets.onTransfer = nil
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("buying a ticket revoked meanwhile: %v", err)
	}
	if wallets.charged() != base {
		t.Fatalf("a buyer of a revoked ticket stays charged: %d", wallets.charged())
	}
	if st := listingStatus(t, p, l.ID); st != domain.ResaleCancelled {
		t.Fatalf("the listing of a revoked ticket is %s, want cancelled", st.Name())
	}
	if got, _ := p.trepo.Get(ctx, tk.ID); got.Status != domain.TicketVoid {
		t.Fatalf("the ticket is %s", got.Status.Name())
	}
}

// A process that dies half-way through a purchase leaves a claimed listing; the sweeper finishes or undoes it.
func TestResaleRecovery(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	wallets := newFakeWallets()
	p := newPod(t, d, wallets, podOptions{ratePerSec: 1000})
	age := func(id int64) {
		if _, err := p.pool.Exec(ctx, `update ticket_resale set updated_at = now() - interval '10 minutes' where id = $1`, id); err != nil {
			t.Fatal(err)
		}
	}

	// 1. died after claiming, before paying: back on the market
	_, tk1, l1 := listed(t, p, 100_000)
	if _, err := p.rrepo.Claim(ctx, l1.ID, 6); err != nil {
		t.Fatal(err)
	}
	if n, _ := p.resale.Recover(ctx); n != 0 {
		t.Fatalf("a claim that is still fresh must be left alone: %d", n)
	}
	age(l1.ID)
	if n, err := p.resale.Recover(ctx); err != nil || n != 1 {
		t.Fatalf("recover: %d %v", n, err)
	}
	if listingStatus(t, p, l1.ID) != domain.ResaleOpen {
		t.Fatal("an abandoned claim must reopen")
	}

	// 2. died after paying, before handing over: the ticket is handed over
	if _, err := p.rrepo.Claim(ctx, l1.ID, 6); err != nil {
		t.Fatal(err)
	}
	ref, _ := wallets.Transfer(ctx, transferParams("w-6", "w-5", 90_000))
	ref2, _ := wallets.Transfer(ctx, transferParams("w-6", "platform", 10_000))
	if err := p.rrepo.SetPayment(ctx, l1.ID, ref+"|"+ref2, 10_000); err != nil {
		t.Fatal(err)
	}
	age(l1.ID)
	if n, err := p.resale.Recover(ctx); err != nil || n != 1 {
		t.Fatalf("recover a paid claim: %d %v", n, err)
	}
	if got, _ := p.trepo.Get(ctx, tk1.ID); got.HolderID != 6 || got.Status != domain.TicketValid || got.Code == tk1.Code {
		t.Fatalf("the paid buyer must get the ticket: %+v", got)
	}
	if st, _ := p.resale.Mine(ctx, user(5), domain.ResaleSold, 0, 10); len(st.Items) != 1 {
		t.Fatal("the listing must be sold")
	}

	// 3. died after paying, and the ticket was revoked meanwhile: the buyer is refunded
	_, tk3, l3 := listed(t, p, 100_000)
	if _, err := p.rrepo.Claim(ctx, l3.ID, 7); err != nil {
		t.Fatal(err)
	}
	r1, _ := wallets.Transfer(ctx, transferParams("w-7", "w-5", 100_000))
	if err := p.rrepo.SetPayment(ctx, l3.ID, r1+"|", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := p.tickets.Void(ctx, organizer, tk3.ID, "fraud"); err != nil {
		t.Fatal(err)
	}
	before := wallets.charged()
	age(l3.ID)
	if n, err := p.resale.Recover(ctx); err != nil || n != 1 {
		t.Fatalf("recover a paid claim on a revoked ticket: %d %v", n, err)
	}
	if wallets.charged() != before-1 {
		t.Fatalf("the buyer was not refunded: %d charges (was %d)", wallets.charged(), before)
	}
	if st := listingStatus(t, p, l3.ID); st != domain.ResaleCancelled {
		t.Fatalf("the listing is %s, want cancelled", st.Name())
	}
}

// Listings end when the event does what they were for.
func TestResaleEndsWithTheEvent(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tk, l := listed(t, p, 100_000)

	if _, err := moveEvent(ctx, p, e.ID, `now() - interval '1 hour'`, `now() + interval '2 hours'`); err != nil {
		t.Fatal(err)
	}
	if _, err := p.resale.Buy(ctx, user(6), l.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("buying after the event started: %v", err)
	}
	if n, err := p.resale.Recover(ctx); err != nil || n != 1 {
		t.Fatalf("expire listings: %d %v", n, err)
	}
	if listingStatus(t, p, l.ID) != domain.ResaleCancelled {
		t.Fatal("the listing of a started event must come down")
	}
	if got, _ := p.trepo.Get(ctx, tk.ID); got.Status != domain.TicketValid {
		t.Fatalf("the ticket must be its holder's again (%s)", got.Status.Name())
	}
	wantHistory(t, p, user(5), tk.ID, "issued", "resale_listed", "resale_expired")

	// a listed ticket of an event that ends expires with it, and its listing goes
	e2, tk2, l2 := listed(t, p, 100_000)
	if _, err := moveEvent(ctx, p, e2.ID, `now() - interval '10 hours'`, `now() - interval '7 hours'`); err != nil {
		t.Fatal(err)
	}
	if _, err := p.tickets.ExpireTickets(ctx); err != nil {
		t.Fatal(err)
	}
	if got, _ := p.trepo.Get(ctx, tk2.ID); got.Status != domain.TicketExpired {
		t.Fatalf("a listed ticket after its event: %s", got.Status.Name())
	}
	if listingStatus(t, p, l2.ID) != domain.ResaleCancelled {
		t.Fatal("its listing must end")
	}
}

// The organizer's refund of an order takes the listings of its tickets down.
func TestResaleAndRefund(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	_, tk, l := listed(t, p, 100_000)
	if _, err := p.orders.Cancel(ctx, organizer, tk.OrderID); err != nil {
		t.Fatal(err)
	}
	if listingStatus(t, p, l.ID) != domain.ResaleCancelled {
		t.Fatal("the listing of a refunded ticket must come down")
	}
	if _, err := p.resale.Buy(ctx, user(6), l.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("buying the listing of a refunded ticket: %v", err)
	}
	if got, _ := p.trepo.Get(ctx, tk.ID); got.Status != domain.TicketVoid {
		t.Fatalf("the ticket is %s", got.Status.Name())
	}
}

func transferParams(from, to string, amount int64) outbound.TransferParams {
	return outbound.TransferParams{FromWalletID: from, ToWalletID: to, Amount: amount}
}
