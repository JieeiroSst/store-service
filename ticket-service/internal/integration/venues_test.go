package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/service"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

// draftWithTypes is a draft event with the named seated ticket types (no seats yet).
func draftWithTypes(t testing.TB, p *pod, names ...string) (domain.Event, []domain.TicketType) {
	t.Helper()
	ctx := context.Background()
	e, err := p.events.Create(ctx, organizer, inbound.EventInput{Title: "Stage show", Category: "arts", City: "Hue", Venue: "Somewhere",
		StartsAt: time.Now().Add(72 * time.Hour), EndsAt: time.Now().Add(75 * time.Hour), WalletID: "w", Transferable: true})
	if err != nil {
		t.Fatal(err)
	}
	var out []domain.TicketType
	for i, n := range names {
		tt, err := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: n, Price: int64(1+i) * 300_000, Seated: true, Active: true, MaxPerOrder: 6})
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, tt)
	}
	return e, out
}

func typeByID(t testing.TB, p *pod, id int64) domain.TicketType {
	t.Helper()
	tt, err := p.erepo.TicketType(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return tt
}

// Halls with different stages and different numbers of seats, each described once and used for different events.
func TestVenuesWithDifferentStagesAndSeatCounts(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	catalog := service.NewCatalogService(p.erepo, 0)

	// three halls from templates: an opera house, a lecture hall, an arena
	opera, err := p.venues.Create(ctx, organizer, inbound.VenueInput{Name: "City Opera", City: "Hanoi", Template: "theatre", Params: map[string]int{"rows": 10, "first_row_seats": 12, "growth": 2, "balcony_rows": 3}})
	if err != nil {
		t.Fatal(err)
	}
	lecture, err := p.venues.Create(ctx, organizer, inbound.VenueInput{Name: "Lecture hall", Template: "hall", Params: map[string]int{"rows": 8, "seats_per_row": 15, "aisle_after": 7}})
	if err != nil {
		t.Fatal(err)
	}
	arena, err := p.venues.Create(ctx, organizer, inbound.VenueInput{Name: "Sports arena", Template: "arena", Params: map[string]int{"rows": 6, "seats_per_row": 10}})
	if err != nil {
		t.Fatal(err)
	}
	if lecture.Seats != 120 || arena.Seats != 240 || opera.Sections != 2 || arena.Sections != 4 || opera.Seats <= 120 {
		t.Fatalf("seats: opera %d in %d sections, lecture %d, arena %d in %d sections", opera.Seats, opera.Sections, lecture.Seats, arena.Seats, arena.Sections)
	}
	// and a hall drawn by hand: two stages, an aisle, a pillar and wheelchair places
	custom, err := p.venues.Create(ctx, organizer, inbound.VenueInput{Name: "Black box", Map: &domain.VenueMap{Width: 700, Height: 500,
		Stages: []domain.StageShape{{Label: "MAIN", X: 100, Y: 20, W: 200, H: 40}, {Label: "SIDE", Shape: "polygon", Points: [][2]float64{{450, 20}, {600, 20}, {550, 80}}}},
		Sections: []domain.SectionSpec{
			{Key: "L", Name: "Left", Layout: "grid", Origin: domain.Point{X: 60, Y: 120}, RowSeats: []int{6, 8, 10}, AisleAfter: []int{3}, Skip: []string{"B:4"}, Accessible: []string{"A:1"}},
			{Key: "R", Name: "Right", Layout: "grid", Origin: domain.Point{X: 420, Y: 120}, Rows: 4, SeatsPerRow: 5, NumberFrom: "right"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if custom.Seats != 6+8+10-1+20 {
		t.Fatalf("custom hall: %d seats", custom.Seats)
	}

	// ---- the arena for one event: VIP faces the stage from the south and north, standing from the sides
	e, types := draftWithTypes(t, p, "VIP", "Side")
	vip, side := types[0], types[1]
	res, err := p.events.ApplyVenueMap(ctx, organizer, e.ID, inbound.ApplyVenueInput{VenueID: arena.ID, Assignments: []inbound.SectionAssignment{
		{Section: "S", TicketTypeID: vip.ID}, {Section: "N", TicketTypeID: vip.ID}, {Section: "E", TicketTypeID: side.ID}, {Section: "W", TicketTypeID: side.ID}}})
	if err != nil || res.Seats != 240 || len(res.Sections) != 4 {
		t.Fatalf("apply: %+v %v", res, err)
	}
	if v, s := typeByID(t, p, vip.ID), typeByID(t, p, side.ID); v.Total != 120 || v.Available != 120 || s.Total != 120 || s.Available != 120 {
		t.Fatalf("stock: vip %+v side %+v", v, s)
	}
	publish(t, p, e.ID)
	m, err := catalog.EventSeatMap(ctx, e.ID)
	if err != nil || len(m.Seats) != 240 || len(m.Types) != 2 {
		t.Fatalf("event seat map: %d seats, %d types, %v", len(m.Seats), len(m.Types), err)
	}
	if len(m.Layout.Sections) != 4 || len(m.Layout.Stages) != 1 || m.Layout.Stages[0].Shape != "ellipse" {
		t.Fatalf("drawing: %+v", m.Layout)
	}
	colour := map[string]int64{}
	for _, sec := range m.Layout.Sections {
		colour[sec.Key] = sec.TicketTypeID
	}
	if colour["S"] != vip.ID || colour["N"] != vip.ID || colour["E"] != side.ID || colour["W"] != side.ID {
		t.Fatalf("sections are sold by: %v", colour)
	}
	perSection := map[string]int{}
	byID := map[int64]domain.Seat{}
	for _, s := range m.Seats {
		perSection[s.Section]++
		byID[s.ID] = s
		if s.X == nil || s.Y == nil {
			t.Fatalf("seat %s %s-%d has no place", s.Section, s.Row, s.Number)
		}
	}
	if perSection["S"] != 60 || perSection["N"] != 60 || perSection["E"] != 60 || perSection["W"] != 60 {
		t.Fatalf("per section: %v", perSection)
	}
	// every section has a row A: same rows, different sections, one ticket type
	// buy two seats of different sections of one ticket type in one order
	var s1, s2 int64
	for _, s := range m.Seats {
		if s.TicketTypeID == vip.ID && s.Section == "S" && s1 == 0 {
			s1 = s.ID
		}
		if s.TicketTypeID == vip.ID && s.Section == "N" && s2 == 0 {
			s2 = s.ID
		}
	}
	c := cmd(e.ID, vip.ID, 2, "arena-1")
	c.Items[0].SeatIDs = []int64{s1, s2}
	o, err := p.orders.Reserve(ctx, user(5), c)
	if err != nil || o.Total != 600_000 {
		t.Fatalf("buying seats of two sections: %+v %v", o, err)
	}
	if _, err := p.orders.Pay(ctx, user(5), o.ID, inbound.PayCommand{Method: domain.MethodWallet}); err != nil {
		t.Fatal(err)
	}
	if tk, _ := p.tickets.List(ctx, user(5), inbound.TicketListQuery{}); len(tk.Items) != 2 || tk.Items[0].SeatLabel == "" {
		t.Fatalf("tickets: %+v", tk.Items)
	}

	// a seat map cannot be replaced once something is sold or held
	if _, err := p.events.ApplyVenueMap(ctx, organizer, e.ID, inbound.ApplyVenueInput{VenueID: lecture.ID, Assignments: []inbound.SectionAssignment{{Section: "MAIN", TicketTypeID: vip.ID}}}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("giving a type with seats another map without saying replace: %v", err)
	}
	if _, err := p.events.ApplyVenueMap(ctx, organizer, e.ID, inbound.ApplyVenueInput{VenueID: lecture.ID, Replace: true, Assignments: []inbound.SectionAssignment{{Section: "MAIN", TicketTypeID: vip.ID}}}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("replacing seats that are sold: %v", err)
	}
	if got := typeByID(t, p, vip.ID); got.Total != 120 || got.Sold != 2 {
		t.Fatalf("a refused replacement changed the stock: %+v", got)
	}

	// ---- the opera for another event: only the stalls are sold, the balcony stays closed
	e2, ty2 := draftWithTypes(t, p, "Stalls")
	if res, err = p.events.ApplyVenueMap(ctx, organizer, e2.ID, inbound.ApplyVenueInput{VenueID: opera.ID, Assignments: []inbound.SectionAssignment{{Section: "ORCH", TicketTypeID: ty2[0].ID}}}); err != nil {
		t.Fatal(err)
	}
	if res.Seats >= opera.Seats || len(res.Sections) != 1 || typeByID(t, p, ty2[0].ID).Total != res.Seats {
		t.Fatalf("stalls only: %+v of %d", res, opera.Seats)
	}
	publish(t, p, e2.ID)
	m2, _ := catalog.EventSeatMap(ctx, e2.ID)
	if len(m2.Layout.Sections) != 1 || m2.Layout.Sections[0].Key != "ORCH" || len(m2.Seats) != res.Seats {
		t.Fatalf("the closed balcony must not be on the map: %+v", m2.Layout.Sections)
	}
	// ...and swapped for the black box before anything sells: allowed with replace
	e3, ty3 := draftWithTypes(t, p, "Left", "Right")
	if _, err := p.events.ApplyVenueMap(ctx, organizer, e3.ID, inbound.ApplyVenueInput{VenueID: opera.ID, Assignments: []inbound.SectionAssignment{{Section: "ORCH", TicketTypeID: ty3[0].ID}}}); err != nil {
		t.Fatal(err)
	}
	res, err = p.events.ApplyVenueMap(ctx, organizer, e3.ID, inbound.ApplyVenueInput{VenueID: custom.ID, Replace: true, Assignments: []inbound.SectionAssignment{
		{Section: "L", TicketTypeID: ty3[0].ID}, {Section: "R", TicketTypeID: ty3[1].ID}}})
	if err != nil || res.Seats != custom.Seats {
		t.Fatalf("replace: %+v %v", res, err)
	}
	if l := typeByID(t, p, ty3[0].ID); l.Total != 6+8+10-1 || l.Available != l.Total {
		t.Fatalf("after replacing, the stock is that of the new map: %+v", l)
	}
	ev3, _ := p.events.View(ctx, &organizer, e3.ID)
	if ev3.VenueID != custom.ID {
		t.Fatalf("the event should point at its venue: %d", ev3.VenueID)
	}
	pv, _ := p.erepo.EventSeatMap(ctx, e3.ID)
	acc, skipped := 0, false
	for _, s := range pv.Seats {
		if s.Accessible {
			acc++
		}
		if s.Section == "L" && s.Row == "B" && s.Number == 4 {
			skipped = true
		}
	}
	if acc != 1 || skipped || len(pv.Layout.Stages) != 2 || pv.Layout.Stages[1].Shape != "polygon" {
		t.Fatalf("custom hall: %d accessible, pillar seat present %v, stages %+v", acc, skipped, pv.Layout.Stages)
	}
	// a seat drawn by hand may still be nudged, section-aware
	if n, err := p.events.SetSeatMap(ctx, organizer, e3.ID, ty3[1].ID, domain.SeatMapLayout{Width: 700, Height: 500},
		[]domain.SeatPosition{{Section: "R", Row: "A", Number: 1, X: 5, Y: 5}, {Section: "ZZ", Row: "A", Number: 1, X: 6, Y: 6}}); err != nil || n != 1 {
		t.Fatalf("nudging a seat: placed %d, %v", n, err)
	}

	_ = byID
	checkLedger(t, p, vip.ID)
}

func TestVenueRefusals(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	hall, err := p.venues.Create(ctx, organizer, inbound.VenueInput{Name: "Hall", Template: "hall"})
	if err != nil {
		t.Fatal(err)
	}
	e, ty := draftWithTypes(t, p, "A", "B")
	ga, _ := p.events.CreateTicketType(ctx, organizer, e.ID, inbound.TicketTypeInput{Name: "GA", Price: 1, Total: 5, Active: true})
	apply := func(a ...inbound.SectionAssignment) error {
		_, err := p.events.ApplyVenueMap(ctx, organizer, e.ID, inbound.ApplyVenueInput{VenueID: hall.ID, Assignments: a})
		return err
	}
	for name, c := range map[string]struct {
		err  error
		want error
	}{
		"no assignments":     {apply(), domain.ErrInvalid},
		"unknown section":    {apply(inbound.SectionAssignment{Section: "NOPE", TicketTypeID: ty[0].ID}), domain.ErrInvalid},
		"general admission":  {apply(inbound.SectionAssignment{Section: "MAIN", TicketTypeID: ga.ID}), domain.ErrInvalid},
		"a type of no event": {apply(inbound.SectionAssignment{Section: "MAIN", TicketTypeID: 99999}), domain.ErrInvalid},
		"twice":              {apply(inbound.SectionAssignment{Section: "MAIN", TicketTypeID: ty[0].ID}, inbound.SectionAssignment{Section: "MAIN", TicketTypeID: ty[1].ID}), domain.ErrInvalid},
	} {
		if !errors.Is(c.err, c.want) {
			t.Errorf("%s: %v", name, c.err)
		}
	}
	// somebody else's private venue does not exist; a stranger cannot apply to my event; a venue must exist
	if _, err := p.events.ApplyVenueMap(ctx, user(9), e.ID, inbound.ApplyVenueInput{VenueID: hall.ID}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("a stranger applying to a draft: %v", err)
	}
	if _, err := p.events.ApplyVenueMap(ctx, organizer, e.ID, inbound.ApplyVenueInput{VenueID: 99999, Assignments: []inbound.SectionAssignment{{Section: "MAIN", TicketTypeID: ty[0].ID}}}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("an unknown venue: %v", err)
	}
	other, _ := p.venues.Create(ctx, user(9), inbound.VenueInput{Name: "Private", Template: "hall"})
	if _, err := p.events.ApplyVenueMap(ctx, organizer, e.ID, inbound.ApplyVenueInput{VenueID: other.ID, Assignments: []inbound.SectionAssignment{{Section: "MAIN", TicketTypeID: ty[0].ID}}}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("another organizer's private venue: %v", err)
	}
	if n := typeByID(t, p, ty[0].ID).Total; n != 0 {
		t.Fatalf("a refused apply left %d tickets behind", n)
	}

	// what a venue can be made of
	for name, in := range map[string]inbound.VenueInput{
		"nothing":      {Name: "x1"},
		"both":         {Name: "x1", Template: "hall", Map: &domain.VenueMap{}},
		"no name":      {Template: "hall"},
		"bad template": {Name: "x1", Template: "castle"},
		"bad params":   {Name: "x1", Template: "hall", Params: map[string]int{"rows": 0}},
		"seats off the canvas": {Name: "x1", Map: &domain.VenueMap{Width: 50, Height: 50,
			Sections: []domain.SectionSpec{{Key: "A", Layout: "grid", Origin: domain.Point{X: 10, Y: 10}, Rows: 3, SeatsPerRow: 5}}}},
	} {
		if _, err := p.venues.Create(ctx, organizer, in); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	// preview saves nothing and says what would be made
	pv, err := p.venues.Preview(ctx, inbound.VenueInput{Template: "arena", Params: map[string]int{"rows": 2, "seats_per_row": 3}})
	if err != nil || pv.Total != 24 || len(pv.Sections) != 4 || pv.Sections[0].Seats != 6 || len(pv.View.Sections) != 4 {
		t.Fatalf("preview: %+v %v", pv, err)
	}
	if page, _ := p.venues.List(ctx, organizer, false, 0, 50); len(page.Items) != 1 {
		t.Fatalf("a preview or a refusal saved a venue: %d", len(page.Items))
	}
}

// The library: private venues, shared ones, copies; deleting a venue leaves its events' seats alone.
func TestVenueLibrary(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	v, err := p.venues.Create(ctx, organizer, inbound.VenueInput{Name: "Home hall", City: "Da Nang", Template: "hall"})
	if err != nil {
		t.Fatal(err)
	}
	stranger := user(9)
	if _, err := p.venues.Get(ctx, stranger, v.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("reading a private venue: %v", err)
	}
	if _, err := p.venues.SetShared(ctx, organizer, v.ID, true); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("an organizer sharing: %v", err)
	}
	if got, err := p.venues.SetShared(ctx, admin, v.ID, true); err != nil || !got.Shared {
		t.Fatalf("an admin sharing: %+v %v", got, err)
	}
	if got, err := p.venues.Get(ctx, stranger, v.ID); err != nil || got.Seats != 200 {
		t.Fatalf("reading a shared venue: %+v %v", got, err)
	}
	if _, err := p.venues.Update(ctx, stranger, v.ID, inbound.VenueInput{Name: "Mine now", Template: "hall"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("editing a shared venue of somebody else: %v", err)
	}
	if lib, _ := p.venues.List(ctx, stranger, true, 0, 10); len(lib.Items) != 1 {
		t.Fatalf("the library: %d", len(lib.Items))
	}
	if mine, _ := p.venues.List(ctx, stranger, false, 0, 10); len(mine.Items) != 0 {
		t.Fatal("the stranger has no venues of their own yet")
	}
	cp, err := p.venues.Copy(ctx, stranger, v.ID, "")
	if err != nil || cp.OwnerID != stranger.UserID || cp.Name != "Copy of Home hall" || cp.Shared || cp.Seats != 200 {
		t.Fatalf("copy: %+v %v", cp, err)
	}
	up, err := p.venues.Update(ctx, stranger, cp.ID, inbound.VenueInput{Name: "My hall", Template: "hall", Params: map[string]int{"rows": 4, "seats_per_row": 10}})
	if err != nil || up.Seats != 40 || up.Name != "My hall" {
		t.Fatalf("editing my copy: %+v %v", up, err)
	}
	if orig, _ := p.venues.Get(ctx, organizer, v.ID); orig.Seats != 200 {
		t.Fatal("editing a copy must not touch the original")
	}

	// an event made from a venue keeps its seats when the venue is edited or deleted
	e, ty := draftWithTypes(t, p, "All")
	if _, err := p.events.ApplyVenueMap(ctx, organizer, e.ID, inbound.ApplyVenueInput{VenueID: v.ID, Assignments: []inbound.SectionAssignment{{Section: "MAIN", TicketTypeID: ty[0].ID}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.venues.Update(ctx, organizer, v.ID, inbound.VenueInput{Name: "Home hall", Template: "hall", Params: map[string]int{"rows": 1, "seats_per_row": 5}}); err != nil {
		t.Fatal(err)
	}
	if err := p.venues.Delete(ctx, organizer, v.ID); err != nil {
		t.Fatal(err)
	}
	if err := p.venues.Delete(ctx, organizer, v.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleting twice: %v", err)
	}
	if got := typeByID(t, p, ty[0].ID); got.Total != 200 {
		t.Fatalf("the event lost seats: %+v", got)
	}
	pv, _ := p.erepo.EventSeatMap(ctx, e.ID)
	ev, _ := p.events.View(ctx, &organizer, e.ID)
	if len(pv.Seats) != 200 || len(pv.Layout.Sections) != 1 || ev.VenueID != 0 {
		t.Fatalf("after deleting the venue: %d seats, %d sections, venue %d", len(pv.Seats), len(pv.Layout.Sections), ev.VenueID)
	}
	// a draft's seat map is not public
	if _, err := service.NewCatalogService(p.erepo, 0).EventSeatMap(ctx, e.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the seat map of a draft: %v", err)
	}
}
