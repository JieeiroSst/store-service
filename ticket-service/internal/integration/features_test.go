package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/service"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

// seatedEvent is a draft with one seated ticket type and rows x perRow seats, not yet submitted.
func seatedEvent(t testing.TB, p *pod, rows, perRow int) (domain.Event, domain.TicketType) {
	t.Helper()
	ctx := context.Background()
	e, err := p.events.Create(ctx, organizer, inbound.EventInput{Title: "Opera night", Category: "arts", City: "Hanoi", Venue: "Opera House",
		StartsAt: time.Now().Add(72 * time.Hour), EndsAt: time.Now().Add(75 * time.Hour), WalletID: "w", Transferable: true})
	if err != nil {
		t.Fatal(err)
	}
	tt, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "Stalls", Price: 900_000, Seated: true, Active: true, MaxPerOrder: 4})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.events.GenerateSeats(ctx, organizer, e.ID, tt.ID, inbound.SeatLayout{Section: "Stalls", Rows: rows, SeatsPerRow: perRow, OffsetX: 100, OffsetY: 200}); err != nil {
		t.Fatal(err)
	}
	return e, tt
}

// Seats have a place on the map and the venue has an outline.
func TestSeatMapLayout(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	catalog := service.NewCatalogService(p.erepo, 0)
	e, tt := seatedEvent(t, p, 2, 3)

	// generated seats sit on a grid from the offset
	seats, err := p.erepo.Seats(ctx, tt.ID)
	if err != nil || len(seats) != 6 {
		t.Fatalf("seats: %d %v", len(seats), err)
	}
	if s := seats[0]; s.X == nil || s.Y == nil || *s.X != 100 || *s.Y != 200 {
		t.Fatalf("A-1 is at %v,%v, want 100,200", s.X, s.Y)
	}
	if s := seats[5]; *s.X != 100+2*30 || *s.Y != 200+1*34 { // B-3
		t.Fatalf("B-3 is at %v,%v", *s.X, *s.Y)
	}
	publish(t, p, e.ID)

	layout := domain.SeatMapLayout{Width: 800, Height: 600,
		Stage:    &domain.MapRect{Label: "STAGE", X: 250, Y: 20, W: 300, H: 60},
		Sections: []domain.SectionShape{{Name: "Stalls", Label: "Stalls", Color: "#c33", Points: [][2]float64{{80, 180}, {400, 180}, {400, 300}, {80, 300}}}}}
	placed, err := p.events.SetSeatMap(ctx, organizer, e.ID, tt.ID, layout, []domain.SeatPosition{
		{Row: "A", Number: 1, X: 150, Y: 220}, {Row: "B", Number: 3, X: 210, Y: 260}, {Row: "Z", Number: 9, X: 1, Y: 1}})
	if err != nil || placed != 2 { // Z-9 is not a seat: ignored
		t.Fatalf("set seat map: placed %d, %v", placed, err)
	}
	m, err := catalog.Seats(ctx, e.ID, tt.ID)
	if err != nil || m.Layout.Width != 800 || m.Layout.Stage == nil || m.Layout.Stage.Label != "STAGE" || len(m.Layout.Sections) != 1 || len(m.Layout.Sections[0].Points) != 4 {
		t.Fatalf("layout: %+v %v", m.Layout, err)
	}
	if a1 := m.Seats[0]; *a1.X != 150 || *a1.Y != 220 {
		t.Fatalf("A-1 moved to %v,%v", *a1.X, *a1.Y)
	}
	if a2 := m.Seats[1]; *a2.X != 130 { // not mentioned: keeps its grid place
		t.Fatalf("A-2 must not move: %v", *a2.X)
	}

	// what is refused
	bad := map[string]domain.SeatMapLayout{
		"no canvas":         {},
		"canvas too big":    {Width: 1e9, Height: 10},
		"stage outside":     {Width: 100, Height: 100, Stage: &domain.MapRect{X: 90, Y: 0, W: 50, H: 10}},
		"section too small": {Width: 100, Height: 100, Sections: []domain.SectionShape{{Name: "x", Points: [][2]float64{{0, 0}, {1, 1}}}}},
		"point outside":     {Width: 100, Height: 100, Sections: []domain.SectionShape{{Name: "x", Points: [][2]float64{{0, 0}, {10, 0}, {500, 10}}}}},
		"nameless section":  {Width: 100, Height: 100, Sections: []domain.SectionShape{{Points: [][2]float64{{0, 0}, {10, 0}, {10, 10}}}}},
	}
	for name, l := range bad {
		if _, err := p.events.SetSeatMap(ctx, organizer, e.ID, tt.ID, l, nil); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := p.events.SetSeatMap(ctx, organizer, e.ID, tt.ID, layout, []domain.SeatPosition{{Row: "A", Number: 1, X: 5000, Y: 1}}); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("a seat outside the canvas: %v", err)
	}
	if _, err := p.events.SetSeatMap(ctx, user(5), e.ID, tt.ID, layout, nil); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("a stranger drawing the map: %v", err)
	}
	ga, _ := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "GA", Price: 1, Total: 5, Active: true})
	if _, err := p.events.SetSeatMap(ctx, organizer, e.ID, ga.ID, layout, nil); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("a map for general admission: %v", err)
	}
}

// Invoices exist for paid orders only, are readable by the buyer and the organizer, and show the VAT.
func TestInvoice(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000,
		invoice: service.InvoiceOptions{Seller: domain.Seller{Name: "Ticket Co", TaxID: "0312345678", Address: "1 Le Loi"}, VATPercent: 10}})
	e, tt := seedEvent(t, p, 20, nil)
	if _, err := p.events.CreatePromotion(ctx, organizer, e.ID, inbound.PromotionInput{Code: "TEN", Kind: domain.PromoPercent, Value: 10, Active: true}); err != nil {
		t.Fatal(err)
	}

	c := cmd(e.ID, tt.ID, 2, "inv")
	c.PromoCode = "TEN"
	x, err := p.orders.Reserve(ctx, user(5), c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.orders.Invoice(ctx, user(5), x.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("an invoice for an unpaid order: %v", err)
	}
	x, err = p.orders.Pay(ctx, user(5), x.ID, inbound.PayCommand{Method: domain.MethodWallet})
	if err != nil {
		t.Fatal(err)
	}
	inv, err := p.orders.Invoice(ctx, user(5), x.ID)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Number == "" || inv.Seller.Name != "Ticket Co" || inv.Subtotal != 1_000_000 || inv.Discount != 100_000 || inv.Total != 900_000 || inv.Status != "paid" {
		t.Fatalf("invoice: %+v", inv)
	}
	if inv.VATPercent != 10 || inv.VAT != 81_818 { // 900,000 x 10 / 110
		t.Fatalf("VAT: %d%% = %d", inv.VATPercent, inv.VAT)
	}
	if len(inv.Lines) != 1 || inv.Lines[0].Quantity != 2 || inv.Lines[0].Amount != 1_000_000 {
		t.Fatalf("lines: %+v", inv.Lines)
	}
	if _, err := p.orders.Invoice(ctx, organizer, x.ID); err != nil {
		t.Fatalf("the organizer reading the invoice: %v", err)
	}
	if _, err := p.orders.Invoice(ctx, user(6), x.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger reading the invoice: %v", err)
	}
	if _, err := p.orders.Cancel(ctx, organizer, x.ID); err != nil {
		t.Fatal(err)
	}
	if inv, err = p.orders.Invoice(ctx, user(5), x.ID); err != nil || inv.Status != "refunded" || inv.RefundAmount != 900_000 {
		t.Fatalf("invoice of a refunded order: %+v %v", inv, err)
	}
}

func TestVATIncluded(t *testing.T) {
	for _, c := range []struct {
		gross   int64
		percent int
		want    int64
	}{{110, 10, 10}, {900_000, 10, 81_818}, {1_000, 8, 74}, {1_000, 0, 0}, {0, 10, 0}} {
		if got := domain.VATIncluded(c.gross, c.percent); got != c.want {
			t.Errorf("VAT in %d at %d%% = %d, want %d", c.gross, c.percent, got, c.want)
		}
	}
}

// Another showtime of the same show: same tickets, fresh stock, a place in the series.
func TestEventSeries(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	catalog := service.NewCatalogService(p.erepo, 0)

	e, seated := seatedEvent(t, p, 2, 2)
	open := time.Now().Add(-time.Hour)
	if _, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "GA", Price: 300_000, Total: 40, Active: true, MaxPerUser: 4, SaleStartsAt: &open, SaleEndsAt: ptrTime(e.StartsAt.Add(-time.Hour))}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.events.CreatePromotion(ctx, organizer, e.ID, inbound.PromotionInput{Code: "EARLY", Kind: domain.PromoPercent, Value: 20, MaxUses: 5, ValidTo: ptrTime(e.StartsAt), Active: true}); err != nil {
		t.Fatal(err)
	}
	publish(t, p, e.ID)
	ga := func(ev domain.Event) domain.TicketType {
		for _, tt := range ev.TicketTypes {
			if tt.Name == "GA" {
				return tt
			}
		}
		t.Fatal("no GA")
		return domain.TicketType{}
	}
	src, _ := p.events.View(ctx, &organizer, e.ID)
	paidOrder(t, p, 5, src, ga(src), 3) // sells 3 GA of the first showtime
	if _, err := p.orders.Reserve(ctx, user(6), inbound.ReserveCommand{EventID: e.ID, PromoCode: "EARLY", BuyerName: "B", BuyerEmail: "b@example.com",
		Items: []inbound.ReserveItem{{TicketTypeID: ga(src).ID, Quantity: 1}}}); err != nil {
		t.Fatal(err)
	}

	next := e.StartsAt.Add(24 * time.Hour)
	if _, err := p.events.Duplicate(ctx, user(5), e.ID, inbound.DuplicateInput{StartsAt: next, EndsAt: next.Add(3 * time.Hour)}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a buyer duplicating: %v", err)
	}
	if _, err := p.events.Duplicate(ctx, organizer, e.ID, inbound.DuplicateInput{StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now()}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("a showtime in the past: %v", err)
	}
	if _, err := p.events.Duplicate(ctx, organizer, e.ID, inbound.DuplicateInput{StartsAt: next, EndsAt: next.Add(-time.Hour)}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("a showtime that ends before it starts: %v", err)
	}
	dup, err := p.events.Duplicate(ctx, organizer, e.ID, inbound.DuplicateInput{Title: "Opera night (matinee)", StartsAt: next, EndsAt: next.Add(3 * time.Hour), CopyPromotions: true})
	if err != nil {
		t.Fatal(err)
	}
	if dup.ID == e.ID || dup.Status != domain.EventDraft || dup.Title != "Opera night (matinee)" || dup.SeriesID != e.ID || dup.Venue != "Opera House" || !dup.Transferable {
		t.Fatalf("duplicate: %+v", dup)
	}
	if len(dup.TicketTypes) != 2 {
		t.Fatalf("ticket types: %+v", dup.TicketTypes)
	}
	// fresh stock: nothing sold or held, even though the first showtime sold 3 and holds 1
	for _, tt := range dup.TicketTypes {
		if tt.Sold != 0 || tt.Available != tt.Total || tt.Total == 0 {
			t.Fatalf("copied type %+v must start with all its tickets", tt)
		}
	}
	g := ga(dup)
	if g.Total != 40 || g.MaxPerUser != 4 || g.SaleStartsAt == nil || g.SaleEndsAt == nil ||
		!g.SaleStartsAt.Equal(open.Add(24*time.Hour)) || !g.SaleEndsAt.Equal(e.StartsAt.Add(-time.Hour).Add(24*time.Hour)) {
		t.Fatalf("the sale window must move with the date: %+v (source opened %v)", g, open)
	}
	// seated types bring their seats and their places
	var newSeated domain.TicketType
	for _, tt := range dup.TicketTypes {
		if tt.Seated {
			newSeated = tt
		}
	}
	seats, _ := p.erepo.Seats(ctx, newSeated.ID)
	if len(seats) != 4 || seats[0].X == nil || *seats[0].X != 100 || seats[0].Status != domain.SeatAvailable || newSeated.ID == seated.ID {
		t.Fatalf("copied seats: %+v", seats)
	}
	promos, _ := p.erepo.Promotions(ctx, dup.ID)
	if len(promos) != 1 || promos[0].Used != 0 || promos[0].MaxUses != 5 || promos[0].ValidTo == nil || !promos[0].ValidTo.Equal(e.StartsAt.Add(24*time.Hour)) {
		t.Fatalf("copied promotions: %+v", promos)
	}
	if none, _ := p.erepo.Promotions(ctx, dup.ID); len(none) != 1 {
		t.Fatal("promotions")
	}
	// the source is untouched and now part of the series
	if src, _ = p.events.View(ctx, &organizer, e.ID); src.SeriesID != e.ID || ga(src).Sold != 3 || ga(src).Held() != 1 {
		t.Fatalf("source after duplicating: series %d, %+v", src.SeriesID, ga(src))
	}

	// a drafted showtime is not listed; once published it is, in date order
	if list, err := catalog.Series(ctx, e.ID); err != nil || len(list) != 1 || list[0].ID != e.ID {
		t.Fatalf("series before publishing: %v %v", list, err)
	}
	publish(t, p, dup.ID)
	third, err := p.events.Duplicate(ctx, organizer, dup.ID, inbound.DuplicateInput{StartsAt: e.StartsAt.Add(48 * time.Hour), EndsAt: e.StartsAt.Add(51 * time.Hour)})
	if err != nil || third.SeriesID != e.ID {
		t.Fatalf("a copy of a copy stays in the same series: %+v %v", third, err)
	}
	publish(t, p, third.ID)
	list, err := catalog.Series(ctx, third.ID)
	if err != nil || len(list) != 3 || list[0].ID != e.ID || list[1].ID != dup.ID || list[2].ID != third.ID {
		t.Fatalf("series: %v %v", ids(list), err)
	}
	if list[0].MinPrice == nil || *list[0].MinPrice != 300_000 {
		t.Fatalf("a series entry shows its cheapest price: %v", list[0].MinPrice)
	}
	if solo, _ := catalog.Series(ctx, 999_999); len(solo) != 0 {
		t.Fatal("an unknown event has no series")
	}
}

func ids(es []domain.Event) []int64 {
	var out []int64
	for _, e := range es {
		out = append(out, e.ID)
	}
	return out
}

type missingUsers map[int64]bool

func (m missingUsers) Lookup(_ context.Context, id int64) (idn outbound.Identity, err error) {
	if m[id] {
		return idn, domain.ErrNotFound
	}
	return outbound.Identity{UserID: id}, nil
}

// An offer by user id goes only to a user that exists.
func TestTransferRecipientMustExist(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	p.tickets.WithDirectory(missingUsers{999: true})
	e, tt := seedEvent(t, p, 5, nil)
	tk := paidOrder(t, p, 5, e, tt, 1).Tickets[0]

	if _, err := p.tickets.Offer(ctx, user(5), tk.ID, inbound.TransferCommand{ToUserID: 999}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("a transfer to nobody: %v", err)
	}
	if got, _ := p.trepo.Get(ctx, tk.ID); got.Status != domain.TicketValid {
		t.Fatal("a refused offer must leave the ticket alone")
	}
	if _, err := p.tickets.Offer(ctx, user(5), tk.ID, inbound.TransferCommand{ToUserID: 6}); err != nil {
		t.Fatalf("a transfer to a real user: %v", err)
	}
}
