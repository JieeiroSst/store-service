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

// The whole life of an event: created, reviewed, sold, paid, scanned at the gate, refunded, expired and cancelled.
func TestEventLifecycle(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	wallets := newFakeWallets()
	p := newPod(t, d, wallets, podOptions{ratePerSec: 1000})

	// ---- organizer sets the event up
	e, err := p.events.Create(ctx, organizer, inbound.EventInput{
		Title: "Anh Trai Say Hi", Category: "music", City: "Ho Chi Minh City", Venue: "Phu Tho Stadium",
		StartsAt: time.Now().Add(72 * time.Hour), EndsAt: time.Now().Add(76 * time.Hour), WalletID: "organizer-wallet", RefundCutoffHours: 24,
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.Status != domain.EventDraft {
		t.Fatalf("new event is %s", e.Status.Name())
	}
	ga, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "GA", Price: 100_000, Total: 10, Active: true})
	if err != nil {
		t.Fatal(err)
	}
	vip, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "VIP", Price: 1_000_000, Total: 3, Active: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "GA", Price: 1, Total: 1, Active: true}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate ticket type name: %v", err)
	}
	promo, err := p.events.CreatePromotion(ctx, organizer, e.ID, inbound.PromotionInput{Code: "save50", Kind: domain.PromoFixed, Value: 50_000, MaxUses: 2, Active: true})
	if err != nil {
		t.Fatal(err)
	}

	// nobody else can touch it, and buyers cannot see it yet
	if _, err := p.events.Update(ctx, user(5), e.ID, inbound.EventInput{}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger editing a draft: %v", err)
	}
	if _, err := p.events.View(ctx, nil, e.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a buyer viewing a draft: %v", err)
	}
	if _, err := p.orders.Reserve(ctx, user(5), cmd(e.ID, ga.ID, 1, "too-early")); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("buying from an unpublished event: %v", err)
	}
	if _, err := p.events.Approve(ctx, organizer, e.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("an organizer approving their own event: %v", err)
	}
	if _, err := p.events.Approve(ctx, admin, e.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("approving a draft that was never submitted: %v", err)
	}

	// ---- review: rejected once, fixed, approved
	if _, err := p.events.Submit(ctx, organizer, e.ID); err != nil {
		t.Fatal(err)
	}
	q, err := p.events.ReviewQueue(ctx, admin, 0, 10)
	if err != nil || len(q.Items) != 1 {
		t.Fatalf("review queue: %v %v", q, err)
	}
	if _, err := p.events.Reject(ctx, admin, e.ID, "banner is missing"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.events.Submit(ctx, organizer, e.ID); err != nil {
		t.Fatal(err)
	}
	if e, err = p.events.Approve(ctx, admin, e.ID); err != nil || e.Status != domain.EventPublished {
		t.Fatalf("approve: %v %v", e.Status, err)
	}

	// ---- buyers browse
	found, err := p.erepo.Search(ctx, domain.EventFilter{Query: "anh trai", Limit: 10})
	if err != nil || len(found) != 1 || found[0].MinPrice == nil || *found[0].MinPrice != 100_000 {
		t.Fatalf("search: %+v %v", found, err)
	}
	if none, err := p.erepo.Search(ctx, domain.EventFilter{Query: "%", Limit: 10}); err != nil || len(none) != 0 {
		t.Fatalf("a literal %% must not match everything: %+v %v", none, err)
	}
	view, err := p.events.View(ctx, nil, e.ID)
	if err != nil || len(view.TicketTypes) != 2 || view.WalletID != "" {
		t.Fatalf("public view: %+v %v", view, err)
	}

	// ---- buyer 5 buys 2 GA + 1 VIP with a promo code, and pays from the wallet
	c := inbound.ReserveCommand{EventID: e.ID, PromoCode: "save50", BuyerName: "Nguyen Van A", BuyerEmail: "a@example.com", RequestID: "o1",
		Items: []inbound.ReserveItem{{TicketTypeID: ga.ID, Quantity: 2}, {TicketTypeID: vip.ID, Quantity: 1}}}
	o1, err := p.orders.Reserve(ctx, user(5), c)
	if err != nil {
		t.Fatal(err)
	}
	if o1.Subtotal != 1_200_000 || o1.Discount != 50_000 || o1.Total != 1_150_000 || o1.Status != domain.OrderPending {
		t.Fatalf("order 1: %+v", o1)
	}
	if got := typeCounters(t, p, ga.ID); got.held != 2 || got.available != 8 {
		t.Fatalf("GA after reserve: %+v", got)
	}
	if _, err := p.orders.Pay(ctx, user(6), o1.ID, inbound.PayCommand{Method: domain.MethodWallet}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("paying someone else's order: %v", err)
	}
	if _, err := p.orders.Pay(ctx, user(5), o1.ID, inbound.PayCommand{Method: "bitcoin"}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("bad payment method: %v", err)
	}
	o1, err = p.orders.Pay(ctx, user(5), o1.ID, inbound.PayCommand{Method: domain.MethodWallet})
	if err != nil || o1.Status != domain.OrderPaid || len(o1.Tickets) != 3 {
		t.Fatalf("pay: %+v %v", o1, err)
	}
	seen := map[string]bool{}
	for _, tk := range o1.Tickets {
		if len(tk.Code) < 20 || seen[tk.Code] || tk.Status != domain.TicketValid {
			t.Fatalf("ticket %+v", tk)
		}
		seen[tk.Code] = true
	}
	if again, err := p.orders.Pay(ctx, user(5), o1.ID, inbound.PayCommand{Method: domain.MethodWallet}); err != nil || again.ID != o1.ID || wallets.charged() != 1 {
		t.Fatalf("paying twice charged %d times (%v)", wallets.charged(), err)
	}
	mine, err := p.tickets.List(ctx, user(5), inbound.TicketListQuery{Limit: 10})
	if err != nil || len(mine.Items) != 3 {
		t.Fatalf("my tickets: %v %v", mine, err)
	}

	// ---- the report
	rep, err := p.events.Report(ctx, organizer, e.ID)
	if err != nil || rep.Orders != 1 || rep.TicketsSold != 3 || rep.Revenue != 1_150_000 || rep.CheckedIn != 0 {
		t.Fatalf("report: %+v %v", rep, err)
	}
	if _, err := p.events.Report(ctx, user(5), e.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a buyer reading the sales report: %v", err)
	}

	// ---- the gate
	code := o1.Tickets[0].Code
	if _, err := p.gate.CheckIn(ctx, user(5), e.ID, code); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a buyer scanning their own ticket: %v", err)
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, "NOSUCHCODE"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown code: %v", err)
	}
	tk, err := p.gate.CheckIn(ctx, organizer, e.ID, code)
	if err != nil || tk.CheckedInAt == nil || tk.CheckedInBy != organizer.UserID {
		t.Fatalf("check-in: %+v %v", tk, err)
	}
	if again, err := p.gate.CheckIn(ctx, organizer, e.ID, code); !errors.Is(err, domain.ErrAlreadyCheckedIn) || again.CheckedInAt == nil {
		t.Fatalf("second scan: %+v %v", again, err)
	}
	if _, err := p.orders.Cancel(ctx, user(5), o1.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("refunding an order with a used ticket: %v", err)
	}

	// ---- buyer 6 buys one, then refunds it (allowed: the event is more than 24h away)
	o2, err := p.orders.Reserve(ctx, user(6), inbound.ReserveCommand{EventID: e.ID, PromoCode: "SAVE50", BuyerName: "B", BuyerEmail: "b@example.com",
		Items: []inbound.ReserveItem{{TicketTypeID: ga.ID, Quantity: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	if o2, err = p.orders.Pay(ctx, user(6), o2.ID, inbound.PayCommand{Method: domain.MethodWallet}); err != nil {
		t.Fatal(err)
	}
	if o2.Total != 50_000 {
		t.Fatalf("order 2 total %d", o2.Total)
	}
	before := wallets.charged()
	o2, err = p.orders.Cancel(ctx, user(6), o2.ID)
	if err != nil || o2.Status != domain.OrderRefunded || o2.RefundAmount != 50_000 || wallets.charged() != before-1 {
		t.Fatalf("refund: %+v %v (charges %d -> %d)", o2, err, before, wallets.charged())
	}
	if got := typeCounters(t, p, ga.ID); got.sold != 2 || got.held != 0 || got.available != 8 {
		t.Fatalf("GA after refund: %+v", got)
	}
	if _, err := p.orders.Get(ctx, user(6), o2.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := p.orders.Get(ctx, user(5), o2.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("reading someone else's order: %v", err)
	}
	org, err := p.orders.List(ctx, organizer, inbound.OrderListQuery{EventID: e.ID})
	if err != nil || len(org.Items) != 2 {
		t.Fatalf("organizer's orders: %v %v", org, err)
	}
	if _, err := p.gate.CheckIn(ctx, organizer, e.ID, o2.Tickets[0].Code); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("scanning a refunded ticket: %v", err)
	}

	// ---- a hold that expires gives its promo use back
	o3, err := p.orders.Reserve(ctx, user(7), inbound.ReserveCommand{EventID: e.ID, PromoCode: "SAVE50", BuyerName: "C", BuyerEmail: "c@example.com",
		Items: []inbound.ReserveItem{{TicketTypeID: ga.ID, Quantity: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.pool.Exec(ctx, `update ticket_order set expires_at = now() - interval '1 second' where id = $1`, o3.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := p.orders.Pay(ctx, user(7), o3.ID, inbound.PayCommand{Method: domain.MethodWallet}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("paying after the hold expired: %v", err)
	}
	if n, err := p.orders.ReleaseExpired(ctx); err != nil || n != 1 {
		t.Fatalf("release expired: %d %v", n, err)
	}
	if o3, err = p.orders.Get(ctx, user(7), o3.ID); err != nil || o3.Status != domain.OrderExpired {
		t.Fatalf("expired order: %+v %v", o3, err)
	}
	var used int
	if err := p.pool.QueryRow(ctx, `select used from promotion where id = $1`, promo.ID).Scan(&used); err != nil || used != 1 {
		// only order 1 is alive: order 2 was refunded and order 3 expired, and both gave their use back
		t.Fatalf("promo used = %d (%v), want 1", used, err)
	}

	// ---- wishlist
	wl := service.NewWishlistService(postgres.NewWishlistRepository(p.pool), p.erepo)
	if err := wl.Add(ctx, user(8), e.ID); err != nil {
		t.Fatal(err)
	}
	if list, err := wl.List(ctx, user(8), 10, 0); err != nil || len(list) != 1 {
		t.Fatalf("wishlist: %v %v", list, err)
	}
	if err := wl.Remove(ctx, user(8), e.ID); err != nil {
		t.Fatal(err)
	}

	// ---- the organizer cancels the event: pending and paid orders are refunded in the background
	o4, err := p.orders.Reserve(ctx, user(9), cmd(e.ID, ga.ID, 3, "o4"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.events.Cancel(ctx, organizer, e.ID, ""); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("cancelling a published event without a reason: %v", err)
	}
	if _, err := p.events.Cancel(ctx, organizer, e.ID, "the venue is closed"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.orders.Reserve(ctx, user(10), cmd(e.ID, ga.ID, 1, "after-cancel")); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("buying after the event was cancelled: %v", err)
	}
	if n, err := p.orders.SettleCancelledEvents(ctx); err != nil || n != 2 {
		t.Fatalf("settle: %d %v", n, err)
	}
	if o1, err = p.orders.Get(ctx, user(5), o1.ID); err != nil || o1.Status != domain.OrderRefunded {
		t.Fatalf("paid order of a cancelled event: %+v %v", o1, err) // even though a ticket was already used
	}
	if o4, err = p.orders.Get(ctx, user(9), o4.ID); err != nil || o4.Status != domain.OrderCancelled {
		t.Fatalf("pending order of a cancelled event: %+v %v", o4, err)
	}
	if wallets.charged() != 0 {
		t.Fatalf("%d wallet charges still standing after every order was refunded", wallets.charged())
	}
	for _, id := range []int64{ga.ID, vip.ID} {
		if got := typeCounters(t, p, id); got.available != got.total || got.sold != 0 || got.held != 0 {
			t.Fatalf("stock after refunding everything: %+v", got)
		}
		checkLedger(t, p, id)
	}
	// buyers were told
	var told int
	if err := p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 5 and kind in ('order_paid', 'event_cancelled')`).Scan(&told); err != nil || told != 2 {
		t.Fatalf("buyer 5 has %d notifications, want 2 (%v)", told, err)
	}
	if n, err := p.orders.SettleCancelledEvents(ctx); err != nil || n != 0 {
		t.Fatalf("settling twice: %d %v", n, err)
	}
}

// Buyers are reminded once, a day before; and every search filter and sort order is valid SQL that returns what it should.
func TestRemindersAndSearchFilters(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})

	mk := func(title, category, city string, startsIn time.Duration, price int64) (domain.Event, domain.TicketType) {
		e, err := p.events.Create(ctx, organizer, inbound.EventInput{Title: title, Category: category, City: city, Venue: "Somewhere",
			StartsAt: time.Now().Add(startsIn), EndsAt: time.Now().Add(startsIn + 2*time.Hour), WalletID: "w"})
		if err != nil {
			t.Fatal(err)
		}
		tt, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "GA", Price: price, Total: 50, Active: true})
		if err != nil {
			t.Fatal(err)
		}
		publish(t, p, e.ID)
		return e, tt
	}
	soon, soonTT := mk("Jazz tonight", "music", "Da Nang", 12*time.Hour, 200_000)
	later, laterTT := mk("Marathon", "sports", "Hanoi", 30*24*time.Hour, 0)

	// a buyer of each; only the event starting within a day triggers a reminder
	for _, x := range []struct {
		u  int64
		e  domain.Event
		tt domain.TicketType
	}{{21, soon, soonTT}, {22, later, laterTT}} {
		o, err := p.orders.Reserve(ctx, user(x.u), cmd(x.e.ID, x.tt.ID, 1, "r"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.orders.Pay(ctx, user(x.u), o.ID, inbound.PayCommand{Method: domain.MethodWallet}); err != nil {
			t.Fatal(err)
		}
	}
	if n, err := p.orders.SendReminders(ctx); err != nil || n != 1 {
		t.Fatalf("reminders: %d %v", n, err)
	}
	if n, err := p.orders.SendReminders(ctx); err != nil || n != 0 {
		t.Fatalf("a second sweep must not remind again: %d %v", n, err)
	}
	var kind string
	if err := p.pool.QueryRow(ctx, `select kind from notification where user_id = 21 and kind = 'event_reminder'`).Scan(&kind); err != nil {
		t.Fatalf("buyer 21 was not reminded: %v", err)
	}
	var other int
	_ = p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 22 and kind = 'event_reminder'`).Scan(&other)
	if other != 0 {
		t.Fatal("the buyer of the far-away event was reminded")
	}

	// search
	catalog := service.NewCatalogService(p.erepo, 0)
	search := func(f domain.EventFilter) []string {
		t.Helper()
		rows, err := catalog.Search(ctx, f)
		if err != nil {
			t.Fatalf("%+v: %v", f, err)
		}
		var titles []string
		for _, e := range rows {
			titles = append(titles, e.Title)
		}
		return titles
	}
	in := func(d time.Duration) *time.Time { x := time.Now().Add(d); return &x }
	price := func(n int64) *int64 { return &n }
	for name, c := range map[string]struct {
		f    domain.EventFilter
		want []string
	}{
		"all, soonest first": {domain.EventFilter{}, []string{"Jazz tonight", "Marathon"}},
		"newest first":       {domain.EventFilter{Sort: "newest"}, []string{"Marathon", "Jazz tonight"}},
		"popular":            {domain.EventFilter{Sort: "popular"}, []string{"Jazz tonight", "Marathon"}}, // one sale each, ties by date
		"by city":            {domain.EventFilter{City: "hanoi"}, []string{"Marathon"}},
		"by category":        {domain.EventFilter{Category: "music"}, []string{"Jazz tonight"}},
		"by text":            {domain.EventFilter{Query: "MARA"}, []string{"Marathon"}},
		"this week":          {domain.EventFilter{From: in(0), To: in(7 * 24 * time.Hour)}, []string{"Jazz tonight"}},
		"paid only":          {domain.EventFilter{MinPrice: price(1)}, []string{"Jazz tonight"}},
		"free only":          {domain.EventFilter{MaxPrice: price(0)}, []string{"Marathon"}},
		"page 2":             {domain.EventFilter{Limit: 1, Offset: 1}, []string{"Marathon"}},
		"nothing":            {domain.EventFilter{Query: "opera"}, nil},
	} {
		got := search(c.f)
		if len(got) != len(c.want) {
			t.Errorf("%s: got %v, want %v", name, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: got %v, want %v", name, got, c.want)
				break
			}
		}
	}
	if _, err := p.events.SetFeatured(ctx, admin, later.ID, true); err != nil {
		t.Fatal(err)
	}
	if got := search(domain.EventFilter{Featured: true}); len(got) != 1 || got[0] != "Marathon" {
		t.Errorf("featured: %v", got)
	}
}
