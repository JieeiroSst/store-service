package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type memWaitlist struct {
	outbound.WaitlistRepository
	entries []domain.WaitlistEntry
	offered map[int64]int64 // entry -> reservation
	gone    map[int64]bool  // entries whose guest has left by the time the offer is recorded
	joinErr error
}

func (m *memWaitlist) Join(_ context.Context, e domain.WaitlistEntry, _ int) (domain.WaitlistEntry, error) {
	if m.joinErr != nil {
		return domain.WaitlistEntry{}, m.joinErr
	}
	e.ID = int64(len(m.entries) + 1)
	e.Status = domain.WaitlistWaiting
	m.entries = append(m.entries, e)
	return e, nil
}
func (m *memWaitlist) Waiting(context.Context, int64, time.Time, int) ([]domain.WaitlistEntry, error) {
	var out []domain.WaitlistEntry
	for _, e := range m.entries {
		if e.Status == domain.WaitlistWaiting {
			out = append(out, e)
		}
	}
	return out, nil
}
func (m *memWaitlist) MarkOffered(_ context.Context, id, res int64) (outbound.OfferResult, error) {
	if m.gone[id] {
		return outbound.OfferGone, nil
	}
	if m.offered == nil {
		m.offered = map[int64]int64{}
	}
	if m.offered[id] == res && res != 0 {
		return outbound.OfferAlready, nil // what a second replica would see
	}
	for i := range m.entries {
		if m.entries[i].ID == id {
			m.entries[i].Status = domain.WaitlistOffered
		}
	}
	m.offered[id] = res
	return outbound.OfferMade, nil
}

// rooms models an inventory of `left` rooms that Reserve takes from, like the real one.
type roomsRepo struct {
	*memReservations
	left    int
	reserve []outbound.ReserveParams
	failFor map[string]error // by request id
}

func (r *roomsRepo) RoomsLeft(context.Context, int64, int64, time.Time, time.Time) (int, error) {
	return r.left, nil
}
func (r *roomsRepo) Reserve(_ context.Context, p outbound.ReserveParams) (domain.Reservation, error) {
	if err := r.failFor[p.RequestID]; err != nil {
		return domain.Reservation{}, err
	}
	// Like the database: a request id that already has a reservation returns that reservation.
	for i, q := range r.reserve {
		if q.RequestID == p.RequestID {
			return domain.Reservation{ID: int64(101 + i), GuestID: p.GuestID, Status: domain.ReservationPending}, nil
		}
	}
	if r.left < p.Rooms {
		return domain.Reservation{}, domain.ErrDatesUnavailable
	}
	r.left -= p.Rooms
	r.reserve = append(r.reserve, p)
	return domain.Reservation{ID: int64(100 + len(r.reserve)), GuestID: p.GuestID, Status: domain.ReservationPending}, nil
}

func newWaitSvc(left int, entries ...domain.WaitlistEntry) (*WaitlistService, *memWaitlist, *roomsRepo, *memInbox) {
	wl := &memWaitlist{}
	for i, e := range entries {
		e.ID, e.Status = int64(i+1), domain.WaitlistWaiting
		if e.HotelID == 0 {
			e.HotelID, e.RoomTypeID = 1, 3
		}
		if e.Rooms == 0 {
			e.Rooms = 1
		}
		wl.entries = append(wl.entries, e)
	}
	rr := &roomsRepo{memReservations: &memReservations{}, left: left}
	inbox := &memInbox{}
	svc := NewWaitlistService(wl, rr, &fakeHotels{hotel: activeHotel()}, WaitlistOptions{OfferTTL: 15 * time.Minute, MaxPending: 3, Notifier: NewNotifier(inbox, nil)})
	svc.now = func() time.Time { return time.Date(2027, 1, 10, 9, 0, 0, 0, time.UTC) }
	return svc, wl, rr, inbox
}

func entry(guest int64) domain.WaitlistEntry {
	return domain.WaitlistEntry{GuestID: guest, Start: day(12), End: day(14), Adults: 2}
}

func TestJoinOnlyWhenTheRoomsAreGone(t *testing.T) {
	ctx := context.Background()
	cmd := inbound.WaitCommand{HotelID: 1, RoomTypeID: 3, Start: day(12), End: day(14), Rooms: 1, Adults: 2}

	svc, wl, _, _ := newWaitSvc(0)
	e, err := svc.Join(ctx, guestP, cmd)
	if err != nil || e.Status != domain.WaitlistWaiting || len(wl.entries) != 1 {
		t.Fatalf("sold out: %+v %v", e, err)
	}
	svc, _, _, _ = newWaitSvc(2)
	if _, err := svc.Join(ctx, guestP, cmd); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("rooms are available: waiting makes no sense: %v", err)
	}

	svc, _, _, _ = newWaitSvc(0)
	for name, mod := range map[string]func(*inbound.WaitCommand){
		"past":          func(c *inbound.WaitCommand) { c.Start, c.End = day(1), day(3) },
		"empty stay":    func(c *inbound.WaitCommand) { c.End = c.Start },
		"no rooms":      func(c *inbound.WaitCommand) { c.Rooms = 0 },
		"over capacity": func(c *inbound.WaitCommand) { c.Adults = 3 },
		"unknown type":  func(c *inbound.WaitCommand) { c.RoomTypeID = 99 },
	} {
		c := cmd
		mod(&c)
		if _, err := svc.Join(ctx, guestP, c); err == nil {
			t.Errorf("%s: must be refused", name)
		}
	}
	if _, err := svc.Join(ctx, ownerP, cmd); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("the hotel's own manager: %v", err)
	}
}

func TestTheOldestGuestGetsTheRoomAndIsToldToPay(t *testing.T) {
	// Two guests wait; one room comes back.
	svc, wl, rr, inbox := newWaitSvc(1, entry(21), entry(22))
	n, err := svc.Promote(context.Background(), 3)
	if err != nil || n != 1 {
		t.Fatalf("offers: %d %v", n, err)
	}
	if len(rr.reserve) != 1 || rr.reserve[0].GuestID != 21 {
		t.Fatalf("first in line: %+v", rr.reserve)
	}
	p := rr.reserve[0]
	if p.RequestID != "waitlist-1" || p.Actor != 0 || p.MaxPending != 3 || time.Until(p.ExpiresAt) < 0 && false {
		t.Fatalf("offer params: %+v", p)
	}
	if want := time.Date(2027, 1, 10, 9, 15, 0, 0, time.UTC); !p.ExpiresAt.Equal(want) {
		t.Fatalf("the offer holds the rooms for 15 minutes: %v", p.ExpiresAt)
	}
	if wl.entries[0].Status != domain.WaitlistOffered || wl.entries[1].Status != domain.WaitlistWaiting || wl.offered[1] != 101 {
		t.Fatalf("statuses: %+v %v", wl.entries, wl.offered)
	}
	if k := inbox.kinds(21); len(k) != 1 || k[0] != "waitlist.offer" {
		t.Fatalf("the guest is told: %v", k)
	}
	if k := inbox.kinds(50); len(k) != 1 || k[0] != "reservation.created" {
		t.Fatalf("the manager sees a new reservation: %v", k)
	}
	if len(inbox.forUser(22)) != 0 {
		t.Fatal("the second guest keeps waiting quietly")
	}
}

func TestEntriesThatDoNotFitKeepTheirPlace(t *testing.T) {
	big := entry(21)
	big.Rooms = 2 // needs two rooms
	// One room is free: the guest ahead needs two and is skipped, the next one (one room) gets it.
	svc, wl, rr, _ := newWaitSvc(1, big, entry(22))
	if n, _ := svc.Promote(context.Background(), 3); n != 1 || rr.reserve[0].GuestID != 22 {
		t.Fatalf("offers %d, %+v", n, rr.reserve)
	}
	if wl.entries[0].Status != domain.WaitlistWaiting {
		t.Fatal("the skipped guest must keep their place")
	}
	// When a second room comes back, they are next.
	rr.left = 2
	if n, _ := svc.Promote(context.Background(), 3); n != 1 || rr.reserve[1].GuestID != 21 || rr.reserve[1].Rooms != 2 {
		t.Fatalf("second round: %d %+v", n, rr.reserve)
	}
}

func TestOffersThatCannotBeMadeAreSkippedNotFatal(t *testing.T) {
	svc, wl, rr, _ := newWaitSvc(5, entry(21), entry(22), entry(23))
	rr.failFor = map[string]error{
		"waitlist-1": domain.ErrDatesUnavailable, // someone reserved them first
		"waitlist-2": domain.ErrConflict,         // this guest already holds too many unpaid reservations
	}
	if n, err := svc.Promote(context.Background(), 3); err != nil || n != 1 || rr.reserve[0].GuestID != 23 {
		t.Fatalf("offers %d %v %+v", n, err, rr.reserve)
	}
	if wl.entries[0].Status != domain.WaitlistWaiting || wl.entries[1].Status != domain.WaitlistWaiting {
		t.Fatal("skipped guests stay on the list")
	}
}

func TestAGuestWhoLeftDuringTheOfferGivesTheRoomBack(t *testing.T) {
	svc, wl, rr, inbox := newWaitSvc(1, entry(21))
	wl.gone = map[int64]bool{1: true}
	rr.memReservations.x = domain.Reservation{ID: 101, Status: domain.ReservationPending}
	if n, _ := svc.Promote(context.Background(), 3); n != 0 {
		t.Fatalf("no offer stands: %d", n)
	}
	if len(rr.memReservations.transitions) != 1 || rr.memReservations.transitions[0] != domain.ReservationCanceled {
		t.Fatalf("the hold must be cancelled: %v", rr.memReservations.transitions)
	}
	if len(inbox.all) != 0 {
		t.Fatal("nobody is told about an offer that did not happen")
	}
}

func TestFreedRoomsAreOfferedToTheWaitingList(t *testing.T) {
	// A guest cancels; the rooms go to whoever waits, without anyone calling the waiting list by hand.
	waitSvc, _, wr, _ := newWaitSvc(1, entry(21))
	res, repo, _ := newResSvc(pending(), &fakeWallets{}, nil, activeHotel())
	res.opts.Promoter = waitSvc
	if _, err := res.Cancel(context.Background(), guestP, 5); err != nil {
		t.Fatal(err)
	}
	_ = repo
	if len(wr.reserve) != 1 || wr.reserve[0].GuestID != 21 {
		t.Fatalf("the waiting guest was not offered the freed room: %+v", wr.reserve)
	}
	// So does a hold that expires.
	waitSvc, _, wr, _ = newWaitSvc(1, entry(21))
	x := pending()
	past := time.Now().Add(-2 * time.Minute)
	x.ExpiresAt = &past
	res, _, _ = newResSvc(x, &fakeWallets{}, nil, activeHotel())
	res.now = time.Now
	res.opts.Promoter = waitSvc
	res.ReleaseExpired(context.Background())
	if len(wr.reserve) != 1 {
		t.Fatalf("expiry: %+v", wr.reserve)
	}
	// And so does new inventory from the manager.
	waitSvc, _, wr, _ = newWaitSvc(1, entry(21))
	cat, _ := newCatalog(activeHotel())
	cat.WithPromoter(waitSvc)
	rate := "4500000"
	if err := cat.SetInventory(context.Background(), ownerP, inbound.InventoryCommand{HotelID: 1, RoomTypeID: 3, From: day(12), To: day(14), Total: i(5), Rate: &rate}); err != nil {
		t.Fatal(err)
	}
	if len(wr.reserve) != 1 {
		t.Fatalf("inventory: %+v", wr.reserve)
	}
}

func TestASecondReplicaMakingTheSameOfferDoesNotCancelIt(t *testing.T) {
	// Every replica runs the waiting list round; two of them meet on the same entry. The second finds the offer
	// already recorded and must leave the reservation alone (it used to cancel it as if the guest had left).
	svc, wl, rr, inbox := newWaitSvc(2, entry(21))
	rr.memReservations.x = domain.Reservation{ID: 101, GuestID: 21, Status: domain.ReservationPending}
	if n, _ := svc.Promote(context.Background(), 3); n != 1 {
		t.Fatalf("first replica: %d", n)
	}
	wl.entries[0].Status = domain.WaitlistWaiting // as the second replica saw it when it listed the entries
	if n, _ := svc.Promote(context.Background(), 3); n != 0 {
		t.Fatalf("second replica must not count another offer: %d", n)
	}
	if len(rr.memReservations.transitions) != 0 || rr.memReservations.x.Status != domain.ReservationPending {
		t.Fatalf("the offered reservation was cancelled: %v %v", rr.memReservations.transitions, rr.memReservations.x.Status)
	}
	if k := inbox.kinds(21); len(k) != 1 {
		t.Fatalf("the guest is told once: %v", k)
	}
}
