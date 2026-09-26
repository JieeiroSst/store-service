package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

func activeHotel() domain.Hotel {
	return domain.Hotel{ID: 1, OwnerID: 50, Status: domain.HotelActive, Currency: "VND", WalletID: "hotel-wallet", FreeCancelHours: 48, Name: "Grand", CheckInTime: "14:00", CheckOutTime: "12:00", Address: "1 Beach Rd", PhoneNumber: "+84 236 123 456",
		RoomTypes: []domain.RoomType{{ID: 3, HotelID: 1, Name: "Deluxe", Capacity: 2, Active: true}, {ID: 4, HotelID: 1, Name: "Old", Capacity: 2}}}
}

func pending() domain.Reservation {
	exp := time.Now().Add(10 * time.Minute)
	return domain.Reservation{ID: 5, HotelID: 1, RoomTypeID: 3, GuestID: 7, Status: domain.ReservationPending, Currency: "VND",
		TotalAmount: "5000000.5000", Start: time.Now().AddDate(0, 0, 10), End: time.Now().AddDate(0, 0, 12), Rooms: 1, ExpiresAt: &exp}
}

func newResSvc(x domain.Reservation, w *fakeWallets, g *fakeGateway, hotel domain.Hotel) (*ReservationService, *memReservations, *fakeHotels) {
	repo := &memReservations{x: x}
	hotels := &fakeHotels{hotel: hotel}
	var wg outbound.WalletGateway
	var pg outbound.PaymentGateway
	if w != nil {
		wg = w
	}
	if g != nil {
		pg = g
	}
	svc := NewReservationService(repo, hotels, NewPaymentRail(wg, pg, nil, nil), ReservationOptions{HoldTTL: time.Minute})
	svc.now = func() time.Time { return time.Date(2027, 1, 10, 15, 0, 0, 0, time.UTC) }
	return svc, repo, hotels
}

func day(d int) time.Time { return time.Date(2027, 1, d, 0, 0, 0, 0, time.UTC) }

func TestReserveValidation(t *testing.T) {
	ok := inbound.ReserveCommand{HotelID: 1, RoomTypeID: 3, Start: day(12), End: day(14), Rooms: 1, Adults: 2}
	tests := []struct {
		name string
		who  inbound.Principal
		mod  func(*inbound.ReserveCommand)
		want error
	}{
		{"ok", guestP, nil, nil},
		{"the hotel's own manager", ownerP, nil, domain.ErrForbidden},
		{"admins may reserve like anyone", adminP, nil, nil},
		{"past check-in", guestP, func(c *inbound.ReserveCommand) { c.Start, c.End = day(1), day(3) }, domain.ErrInvalid},
		{"check-out not after check-in", guestP, func(c *inbound.ReserveCommand) { c.End = day(12) }, domain.ErrInvalid},
		{"too long a stay", guestP, func(c *inbound.ReserveCommand) { c.End = day(12).AddDate(0, 0, 31) }, domain.ErrInvalid},
		{"no rooms", guestP, func(c *inbound.ReserveCommand) { c.Rooms = 0 }, domain.ErrInvalid},
		{"too many rooms", guestP, func(c *inbound.ReserveCommand) { c.Rooms = 11 }, domain.ErrInvalid},
		{"no adult", guestP, func(c *inbound.ReserveCommand) { c.Adults = 0 }, domain.ErrInvalid},
		{"over capacity", guestP, func(c *inbound.ReserveCommand) { c.Adults, c.Children = 2, 1 }, domain.ErrInvalid},
		{"switched-off room type", guestP, func(c *inbound.ReserveCommand) { c.RoomTypeID = 4 }, domain.ErrNotFound},
		{"unknown room type", guestP, func(c *inbound.ReserveCommand) { c.RoomTypeID = 99 }, domain.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, _ := newResSvc(domain.Reservation{}, nil, nil, activeHotel())
			cmd := ok
			if tc.mod != nil {
				tc.mod(&cmd)
			}
			_, err := svc.Reserve(context.Background(), tc.who, cmd)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if tc.want == nil {
				p := repo.reserved
				if p == nil || p.Currency != "VND" || p.RequestID == "" || p.AddonsTotal != "0.0000" || !p.ExpiresAt.After(time.Now().Add(-time.Hour)) {
					t.Fatalf("defaults not applied: %+v", p)
				}
			} else if repo.reserved != nil {
				t.Fatal("repository called despite failure")
			}
		})
	}
	// An unpublished hotel cannot be reserved.
	h := activeHotel()
	h.Status = domain.HotelPending
	svc, _, _ := newResSvc(domain.Reservation{}, nil, nil, h)
	if _, err := svc.Reserve(context.Background(), guestP, ok); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("pending hotel: %v", err)
	}
}

func TestReservePricesExtraServicesFromTheCatalogue(t *testing.T) {
	svc, repo, hotels := newResSvc(domain.Reservation{}, nil, nil, activeHotel())
	hotels.services = []domain.HotelService{
		{ID: 10, HotelID: 1, Name: "Airport transfer", Price: "1500000", Unit: domain.UnitStay},
		{ID: 11, HotelID: 1, Name: "Breakfast", Price: "250000.50", Unit: domain.UnitNight},
	}
	cmd := inbound.ReserveCommand{HotelID: 1, RoomTypeID: 3, Start: day(12), End: day(15), Rooms: 1, Adults: 2,
		Addons: []inbound.AddonRequest{{ServiceID: 10, Quantity: 2}, {ServiceID: 11, Quantity: 2}}}
	if _, err := svc.Reserve(context.Background(), guestP, cmd); err != nil {
		t.Fatal(err)
	}
	p := repo.reserved
	// 2 transfers once = 3,000,000; 2 breakfasts x 3 nights x 250,000.50 = 1,500,003.00
	if p.AddonsTotal != "4500003.0000" || len(p.Addons) != 2 || p.Addons[1].Amount != "1500003.0000" || p.Addons[0].Name != "Airport transfer" {
		t.Fatalf("addons: %+v total %s", p.Addons, p.AddonsTotal)
	}
	for name, add := range map[string][]inbound.AddonRequest{
		"not offered": {{ServiceID: 99, Quantity: 1}},
		"twice":       {{ServiceID: 10, Quantity: 1}, {ServiceID: 10, Quantity: 1}},
		"zero":        {{ServiceID: 10, Quantity: 0}},
		"too many":    {{ServiceID: 10, Quantity: 21}},
	} {
		cmd.Addons = add
		if _, err := svc.Reserve(context.Background(), guestP, cmd); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestPayWithWalletGoesToTheHotelsWalletAndIsIdempotent(t *testing.T) {
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "vnd"}}
	svc, _, _ := newResSvc(pending(), w, nil, activeHotel())
	svc.now = time.Now
	cmd := inbound.PayCommand{Method: domain.MethodWallet}
	x, err := svc.Pay(context.Background(), guestP, 5, cmd)
	if err != nil || x.Status != domain.ReservationPaid || x.PaymentRef != "t-reservation-5" {
		t.Fatalf("pay: %+v %v", x, err)
	}
	if len(w.to) != 1 || w.to[0] != "hotel-wallet" {
		t.Fatalf("money must go to the hotel's own wallet: %v", w.to)
	}
	if _, err := svc.Pay(context.Background(), guestP, 5, cmd); err != nil || len(w.transfers) != 1 {
		t.Fatalf("second pay must not charge again: %v transfers=%d", err, len(w.transfers))
	}
}

func TestPayErrors(t *testing.T) {
	good := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	cases := []struct {
		name string
		w    *fakeWallets
		who  inbound.Principal
		mod  func(*domain.Reservation, *domain.Hotel)
		cmd  inbound.PayCommand
		want error
	}{
		{"no wallet", &fakeWallets{noWallet: true}, guestP, nil, inbound.PayCommand{Method: domain.MethodWallet}, domain.ErrInvalid},
		{"currency mismatch", &fakeWallets{wallet: domain.Wallet{ID: "w", Currency: "USD"}}, guestP, nil, inbound.PayCommand{Method: domain.MethodWallet}, domain.ErrInvalid},
		{"insufficient", &fakeWallets{wallet: domain.Wallet{ID: "w", Currency: "VND"}, transferErr: domain.ErrInsufficientFunds}, guestP, nil, inbound.PayCommand{Method: domain.MethodWallet}, domain.ErrInsufficientFunds},
		{"someone else's reservation", good, otherP, nil, inbound.PayCommand{Method: domain.MethodWallet}, domain.ErrNotFound},
		{"hotel without a wallet", good, guestP, func(_ *domain.Reservation, h *domain.Hotel) { h.WalletID = "" }, inbound.PayCommand{Method: domain.MethodWallet}, domain.ErrInvalid},
		{"unknown method", good, guestP, nil, inbound.PayCommand{Method: "cash"}, domain.ErrInvalid},
		{"gateway disabled", good, guestP, nil, inbound.PayCommand{Method: domain.MethodGateway, Provider: "stripe"}, domain.ErrInvalid},
		{"hold expired", good, guestP, func(x *domain.Reservation, _ *domain.Hotel) { past := time.Now().Add(-time.Hour); x.ExpiresAt = &past }, inbound.PayCommand{Method: domain.MethodWallet}, domain.ErrConflict},
		{"already canceled", good, guestP, func(x *domain.Reservation, _ *domain.Hotel) { x.Status = domain.ReservationCanceled }, inbound.PayCommand{Method: domain.MethodWallet}, domain.ErrConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			x, h := pending(), activeHotel()
			if tc.mod != nil {
				tc.mod(&x, &h)
			}
			svc, repo, _ := newResSvc(x, tc.w, nil, h)
			svc.now = time.Now
			_, err := svc.Pay(context.Background(), tc.who, 5, tc.cmd)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if repo.x.Status == domain.ReservationPaid {
				t.Fatal("must not be paid")
			}
		})
	}
}

func TestPayRefundsWhenReservationNoLongerPayable(t *testing.T) {
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	svc, repo, _ := newResSvc(pending(), w, nil, activeHotel())
	svc.now = time.Now
	repo.failPaid = domain.ErrConflict // the hold expired while the transfer was in flight
	_, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodWallet})
	if !errors.Is(err, domain.ErrConflict) || len(w.reversed) != 1 {
		t.Fatalf("want conflict + refund, got %v reversed=%v", err, w.reversed)
	}
	// A transient error must not refund: a retry finds the same transfer and confirms it.
	w2 := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	svc, repo, _ = newResSvc(pending(), w2, nil, activeHotel())
	svc.now = time.Now
	repo.failPaid = domain.ErrUpstreamUnavailable
	if _, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodWallet}); err == nil || len(w2.reversed) != 0 {
		t.Fatalf("transient: err=%v reversed=%v", err, w2.reversed)
	}
	if x, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodWallet}); err != nil || x.Status != domain.ReservationPaid {
		t.Fatalf("retry: %+v %v", x, err)
	}
}

func TestPayWithGateway(t *testing.T) {
	g := &fakeGateway{status: domain.GatewayPending}
	svc, repo, _ := newResSvc(pending(), nil, g, activeHotel())
	svc.now = time.Now
	x, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodGateway, Provider: "Stripe"})
	if err != nil || x.Status != domain.ReservationPending || x.PaymentRef != "77" {
		t.Fatalf("in flight: %+v %v", x, err)
	}
	if in := g.created[0]; in.Provider != "stripe" || in.Amount != 5000001 || in.PayerEmail != "g@x.com" {
		t.Fatalf("gateway request: %+v", in)
	}
	svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodGateway, Provider: "stripe"})
	if len(g.created) != 1 {
		t.Fatalf("retrying while in flight must not create a second payment: %d", len(g.created))
	}
	g.status = domain.GatewayCaptured // polling the reservation confirms it
	if got, err := svc.Get(context.Background(), guestP, 5); err != nil || got.Status != domain.ReservationPaid || repo.x.Status != domain.ReservationPaid {
		t.Fatalf("sync: %+v %v", got, err)
	}
}

// ---- the state diagram: pending -> paid | canceled | rejected, paid -> refunded

func TestGuestCancelsPendingReservation(t *testing.T) {
	svc, repo, _ := newResSvc(pending(), &fakeWallets{}, nil, activeHotel())
	x, err := svc.Cancel(context.Background(), guestP, 5)
	if err != nil || x.Status != domain.ReservationCanceled || len(repo.transitions) != 1 {
		t.Fatalf("cancel: %+v %v", x, err)
	}
	// Cancelling again changes nothing.
	if _, err := svc.Cancel(context.Background(), guestP, 5); err != nil || len(repo.transitions) != 1 {
		t.Fatalf("second cancel: %v %d", err, len(repo.transitions))
	}
	if _, err := svc.Cancel(context.Background(), otherP, 5); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger: %v", err)
	}
}

func paidAt(start time.Time) domain.Reservation {
	x := pending()
	x.Status, x.PaymentMethod, x.PaymentRef, x.Start, x.End = domain.ReservationPaid, domain.MethodWallet, "t-1", start, start.AddDate(0, 0, 2)
	return x
}

func TestPaidReservationIsRefundedWithinTheFreeCancellationWindow(t *testing.T) {
	// "now" is Jan 10 15:00; check-in Jan 14 00:00 is 81h away, the window is 48h.
	w := &fakeWallets{}
	svc, repo, _ := newResSvc(paidAt(day(14)), w, nil, activeHotel())
	x, err := svc.Cancel(context.Background(), guestP, 5)
	if err != nil || x.Status != domain.ReservationRefunded || len(w.reversed) != 1 || w.reversed[0] != "t-1" {
		t.Fatalf("refund: %+v %v reversed=%v", x, err, w.reversed)
	}
	if repo.transitions[0] != domain.ReservationRefunded {
		t.Fatal("paid -> refunded")
	}
}

func TestGuestCannotRefundInsideTheWindowButTheHotelCan(t *testing.T) {
	// check-in Jan 11 00:00 is 9h away: inside the 48h window.
	w := &fakeWallets{}
	svc, repo, _ := newResSvc(paidAt(day(11)), w, nil, activeHotel())
	if _, err := svc.Cancel(context.Background(), guestP, 5); !errors.Is(err, domain.ErrConflict) || len(w.reversed) != 0 || repo.x.Status != domain.ReservationPaid {
		t.Fatalf("guest inside the window: %v reversed=%v", err, w.reversed)
	}
	// The hotel's manager and admins can refund any time (a goodwill gesture, or the hotel cannot honour it).
	if x, err := svc.Cancel(context.Background(), ownerP, 5); err != nil || x.Status != domain.ReservationRefunded || len(w.reversed) != 1 {
		t.Fatalf("manager: %+v %v", x, err)
	}
}

func TestGatewayRefundOnCancel(t *testing.T) {
	x := paidAt(day(14))
	x.PaymentMethod, x.PaymentRef = domain.MethodGateway, "77"
	g := &fakeGateway{}
	svc, _, _ := newResSvc(x, nil, g, activeHotel())
	if got, err := svc.Cancel(context.Background(), guestP, 5); err != nil || got.Status != domain.ReservationRefunded || len(g.refunded) != 1 || g.refunded[0] != 77 {
		t.Fatalf("gateway refund: %+v %v %v", got, err, g.refunded)
	}
}

func TestRejectOnlyPendingByManagerOrAdmin(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newResSvc(pending(), &fakeWallets{}, nil, activeHotel())
	if _, err := svc.Reject(ctx, guestP, 5, "no"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("the guest rejecting: %v", err)
	}
	if _, err := svc.Reject(ctx, otherP, 5, "no"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger: %v", err)
	}
	if _, err := svc.Reject(ctx, ownerP, 5, " "); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("reason required: %v", err)
	}
	x, err := svc.Reject(ctx, ownerP, 5, "fully booked for a private event")
	if err != nil || x.Status != domain.ReservationRejected || x.StatusNote == "" {
		t.Fatalf("reject: %+v %v", x, err)
	}
	// Paid reservations are cancelled (refunded), not rejected.
	svc, repo, _ = newResSvc(paidAt(day(14)), &fakeWallets{}, nil, activeHotel())
	if _, err := svc.Reject(ctx, adminP, 5, "x"); !errors.Is(err, domain.ErrConflict) || len(repo.transitions) != 0 {
		t.Fatalf("reject paid: %v", err)
	}
}

func TestManagerSeesReservationsAtTheirHotelOnly(t *testing.T) {
	svc, _, _ := newResSvc(pending(), nil, nil, activeHotel())
	if _, err := svc.Get(context.Background(), ownerP, 5); err != nil {
		t.Fatalf("manager: %v", err)
	}
	if _, err := svc.Get(context.Background(), otherP, 5); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("other manager: %v", err)
	}
}

func TestReleaseExpired(t *testing.T) {
	x := pending()
	past := time.Now().Add(-2 * time.Minute)
	x.ExpiresAt = &past
	svc, repo, _ := newResSvc(x, &fakeWallets{}, nil, activeHotel())
	svc.now = time.Now
	if n, err := svc.ReleaseExpired(context.Background()); err != nil || n != 1 || repo.x.Status != domain.ReservationCanceled || repo.x.StatusNote != "hold expired" {
		t.Fatalf("plain expiry: n=%d err=%v %+v", n, err, repo.x)
	}
	// A gateway payment captured after the hold ended still confirms the reservation.
	x.PaymentMethod, x.PaymentRef = domain.MethodGateway, "77"
	svc, repo, _ = newResSvc(x, nil, &fakeGateway{status: domain.GatewayCaptured}, activeHotel())
	svc.now = time.Now
	if n, _ := svc.ReleaseExpired(context.Background()); n != 0 || repo.x.Status != domain.ReservationPaid {
		t.Fatalf("captured: n=%d status=%v", n, repo.x.Status)
	}
	// A payment still in flight gets one more hold period (1 minute here).
	recent := time.Now().Add(-30 * time.Second)
	x.ExpiresAt = &recent
	svc, repo, _ = newResSvc(x, nil, &fakeGateway{status: domain.GatewayPending}, activeHotel())
	svc.now = time.Now
	if n, _ := svc.ReleaseExpired(context.Background()); n != 0 || repo.x.Status != domain.ReservationPending {
		t.Fatalf("in flight: n=%d status=%v", n, repo.x.Status)
	}
	longAgo := time.Now().Add(-10 * time.Minute)
	x.ExpiresAt = &longAgo
	svc, repo, _ = newResSvc(x, nil, &fakeGateway{status: domain.GatewayPending}, activeHotel())
	svc.now = time.Now
	if n, _ := svc.ReleaseExpired(context.Background()); n != 1 || repo.x.Status != domain.ReservationCanceled {
		t.Fatalf("stale: n=%d status=%v", n, repo.x.Status)
	}
}

// ---- a rush on the last room

func rushSvc(opts ReservationOptions) (*ReservationService, *memReservations, *fakeHotels) {
	svc, repo, hotels := newResSvc(domain.Reservation{}, nil, nil, activeHotel())
	opts.HoldTTL = time.Minute
	svc.opts = opts
	return svc, repo, hotels
}

var rushCmd = inbound.ReserveCommand{HotelID: 1, RoomTypeID: 3, Start: day(12), End: day(14), Rooms: 1, Adults: 2}

func TestFirstSoldOutAnswerProtectsTheDatabaseFromTheRest(t *testing.T) {
	cache := NewSoldOutCache(time.Minute)
	svc, repo, _ := rushSvc(ReservationOptions{SoldOut: cache})
	repo.reserveErr = domain.ErrDatesUnavailable // the database says the room is gone

	if _, err := svc.Reserve(context.Background(), guestP, rushCmd); !errors.Is(err, domain.ErrDatesUnavailable) {
		t.Fatalf("first: %v", err)
	}
	repo.reserved = nil
	if !svc.SoldOut(context.Background(), rushCmd) {
		t.Fatal("the answer must be remembered")
	}
	// Everyone after that is refused without touching the repository, even if it would now succeed.
	repo.reserveErr = nil
	for i := 0; i < 100; i++ {
		if _, err := svc.Reserve(context.Background(), inbound.Principal{UserID: int64(1000 + i)}, rushCmd); !errors.Is(err, domain.ErrDatesUnavailable) {
			t.Fatalf("user %d: %v", i, err)
		}
	}
	if repo.reserved != nil {
		t.Fatal("the database must not be asked again while sold out is remembered")
	}
	// A different room type or range is not affected; and once rooms come back the answer is dropped.
	other := rushCmd
	other.End = day(15)
	if svc.SoldOut(context.Background(), other) {
		t.Fatal("other range")
	}
	cache.Release(3)
	if _, err := svc.Reserve(context.Background(), guestP, rushCmd); err != nil || repo.reserved == nil {
		t.Fatalf("after release: %v", err)
	}
}

func TestFreedRoomsClearSoldOut(t *testing.T) {
	cache := NewSoldOutCache(time.Minute)
	cache.MarkSoldOut(3, day(12), day(14), 1)
	svc, _, _ := newResSvc(pending(), &fakeWallets{}, nil, activeHotel())
	svc.opts.SoldOut = cache
	if _, err := svc.Cancel(context.Background(), guestP, 5); err != nil {
		t.Fatal(err)
	}
	if svc.SoldOut(context.Background(), rushCmd) {
		t.Fatal("cancelling frees the rooms, so 'sold out' must be forgotten at once")
	}
}

func TestRetryWithTheSameRequestIDIsNotShownSoldOut(t *testing.T) {
	// The winner's own retry must get their reservation back, not "sold out".
	cache := NewSoldOutCache(time.Minute)
	cache.MarkSoldOut(3, day(12), day(14), 1)
	svc, repo, _ := rushSvc(ReservationOptions{SoldOut: cache})
	repo.byRequest = &domain.Reservation{ID: 9, GuestID: 7, RequestID: "mine", Status: domain.ReservationPending}
	cmd := rushCmd
	cmd.RequestID = "mine"
	x, err := svc.Reserve(context.Background(), guestP, cmd)
	if err != nil || x.ID != 9 || repo.reserved != nil {
		t.Fatalf("retry: %+v %v", x, err)
	}
	// Someone else using that request id gets nothing of it.
	if _, err := svc.Reserve(context.Background(), otherP, cmd); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("other user: %v", err)
	}
}

func TestOneAccountCannotHammerTheEndpoint(t *testing.T) {
	svc, repo, _ := rushSvc(ReservationOptions{Limiter: NewUserLimiter(1, 2)})
	for i := 0; i < 2; i++ {
		if _, err := svc.Reserve(context.Background(), guestP, rushCmd); err != nil {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	repo.reserved = nil
	if _, err := svc.Reserve(context.Background(), guestP, rushCmd); !errors.Is(err, domain.ErrBusy) || repo.reserved != nil {
		t.Fatalf("third attempt must be shed before the database: %v", err)
	}
	if _, err := svc.Reserve(context.Background(), otherP, rushCmd); err != nil {
		t.Fatalf("another account is unaffected: %v", err)
	}
}

func TestBusyBulkheadShedsInsteadOfQueueing(t *testing.T) {
	bh := NewBulkhead(1, 10*time.Millisecond)
	svc, repo, _ := rushSvc(ReservationOptions{Bulkhead: bh})
	leave, _ := bh.Enter(context.Background()) // someone else occupies the only slot
	defer leave()
	if _, err := svc.Reserve(context.Background(), guestP, rushCmd); !errors.Is(err, domain.ErrBusy) || repo.reserved != nil {
		t.Fatalf("want busy without touching the database: %v", err)
	}
}

func TestHoldCapIsPassedToTheRepository(t *testing.T) {
	svc, repo, _ := rushSvc(ReservationOptions{MaxPending: 3})
	if _, err := svc.Reserve(context.Background(), guestP, rushCmd); err != nil || repo.reserved.MaxPending != 3 {
		t.Fatalf("%v %+v", err, repo.reserved)
	}
}

func TestProbeAnswersBeforeAnyoneIsIdentified(t *testing.T) {
	cache := NewSoldOutCache(time.Minute)
	svc, repo, _ := rushSvc(ReservationOptions{SoldOut: cache})
	zero := 0
	repo.left = &zero // the last room was just taken; nobody has been refused yet, so nothing is cached
	if !svc.SoldOut(context.Background(), rushCmd) {
		t.Fatal("the probe must find the range sold out")
	}
	// ...and remembers it: no more looks while it is fresh.
	before := atomic.LoadInt32(&repo.probes)
	for i := 0; i < 1000; i++ {
		if !svc.SoldOut(context.Background(), rushCmd) {
			t.Fatal("remembered")
		}
	}
	if atomic.LoadInt32(&repo.probes) != before {
		t.Fatal("no further probes while sold out is remembered")
	}
	// Two rooms asked for when one is free: the range is not sold out for the smaller request.
	cache.Release(3)
	one := 1
	repo.left = &one
	two := rushCmd
	two.Rooms = 2
	if !svc.SoldOut(context.Background(), two) || svc.SoldOut(context.Background(), rushCmd) {
		t.Fatal("threshold per number of rooms")
	}
}

func TestConcurrentProbesShareOneLookup(t *testing.T) {
	cache := NewSoldOutCache(time.Minute)
	svc, repo, _ := rushSvc(ReservationOptions{SoldOut: cache})
	repo.probeDelay = 100 * time.Millisecond // slow enough that all callers arrive while it runs
	var wg sync.WaitGroup
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.SoldOut(context.Background(), rushCmd)
		}()
	}
	wg.Wait()
	if n := atomic.LoadInt32(&repo.probes); n > 3 {
		t.Fatalf("500 identical questions must cost about one lookup, cost %d", n)
	}
	// Rooms seen free are trusted only for a moment.
	if !cache.RecentlyAvailable(3, rushCmd.Start, rushCmd.End, 1) {
		t.Fatal("recently available")
	}
	time.Sleep(availableFor + 20*time.Millisecond)
	if cache.RecentlyAvailable(3, rushCmd.Start, rushCmd.End, 1) {
		t.Fatal("must expire quickly")
	}
}

func TestSoldOutNeverHidesAnExistingRequestFromItsOwner(t *testing.T) {
	cache := NewSoldOutCache(time.Minute)
	cache.MarkSoldOut(3, day(12), day(14), 1)
	svc, repo, _ := rushSvc(ReservationOptions{SoldOut: cache})
	repo.byRequest = &domain.Reservation{ID: 9, GuestID: 7, RequestID: "mine"}

	known := rushCmd
	known.RequestID = "mine" // exists: could be the winner retrying, so the shortcut must not answer
	fresh := rushCmd
	fresh.RequestID = "brand-new" // a new id is no reason to let it through
	if svc.SoldOut(context.Background(), known) || !svc.SoldOut(context.Background(), fresh) || !svc.SoldOut(context.Background(), rushCmd) {
		t.Fatal("only a request id that already exists may skip the sold-out answer")
	}
	long := rushCmd
	long.RequestID = string(make([]byte, 129))
	if _, err := svc.Reserve(context.Background(), guestP, long); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("oversized request id: %v", err)
	}
}

// ---- cancellation fees

func tieredHotel() domain.Hotel {
	h := activeHotel()
	h.CancellationTiers = []domain.CancellationTier{{HoursBefore: 72, FeePercent: 0}, {HoursBefore: 24, FeePercent: 50}}
	return h
}

// "now" in newResSvc is Jan 10 15:00. Check-in on Jan 12 00:00 is 33h away: the 50% tier.
func TestCancellingInTheMiddleTierCostsAFeeAndPaysTheRestBack(t *testing.T) {
	x := paidAt(day(12))
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	svc, repo, _ := newResSvc(x, w, nil, tieredHotel())

	got, err := svc.Cancel(context.Background(), guestP, 5)
	if err != nil || got.Status != domain.ReservationRefunded {
		t.Fatalf("cancel: %+v %v", got, err)
	}
	// 5,000,000.50 rounds to 5,000,001; half is 2,500,001 (rounded half up), the rest 2,500,000.
	if repo.x.FeeAmount != "2500001.0000" || repo.x.RefundAmount != "2500000.0000" {
		t.Fatalf("fee %s refund %s", repo.x.FeeAmount, repo.x.RefundAmount)
	}
	if len(w.reversed) != 1 || len(w.fees) != 1 || w.fees[0] != 2500001 || w.to[len(w.to)-1] != "hotel-wallet" {
		t.Fatalf("wallet: reversed=%v fees=%v to=%v", w.reversed, w.fees, w.to)
	}
	// Repeating the call after a failure part-way is safe: the fee has a fixed reference.
	if _, ok := w.transfers["reservation-5-fee"]; !ok {
		t.Fatalf("fee reference: %v", w.transfers)
	}
}

func TestGatewayPaymentsAreRefundedPartially(t *testing.T) {
	x := paidAt(day(12))
	x.PaymentMethod, x.PaymentRef = domain.MethodGateway, "77"
	g := &fakeGateway{}
	svc, repo, _ := newResSvc(x, nil, g, tieredHotel())
	if _, err := svc.Cancel(context.Background(), guestP, 5); err != nil {
		t.Fatal(err)
	}
	if g.partial[77] != 2500000 || len(g.refunded) != 0 || repo.x.FeeAmount != "2500001.0000" {
		t.Fatalf("partial refund: %v full refunds %v fee %s", g.partial, g.refunded, repo.x.FeeAmount)
	}
}

func TestTiersAroundTheBoundaries(t *testing.T) {
	// Check-in Jan 13 00:00 is 57h away: between 24h and 72h, half. Jan 14 is 81h: free. Jan 11 is 9h: non-refundable.
	for _, tc := range []struct {
		checkIn time.Time
		fee     string
		err     bool
	}{{day(14), "0.0000", false}, {day(13), "2500001.0000", false}, {day(11), "", true}} {
		w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
		svc, repo, _ := newResSvc(paidAt(tc.checkIn), w, nil, tieredHotel())
		_, err := svc.Cancel(context.Background(), guestP, 5)
		if tc.err {
			if !errors.Is(err, domain.ErrConflict) || repo.x.Status != domain.ReservationPaid || len(w.reversed) != 0 {
				t.Fatalf("%v: must be refused untouched: %v", tc.checkIn, err)
			}
			continue
		}
		if err != nil || repo.x.FeeAmount != tc.fee {
			t.Fatalf("%v: fee %q err %v, want %s", tc.checkIn, repo.x.FeeAmount, err, tc.fee)
		}
	}
}

func TestTheHotelNeverPaysAFeeToCancelForItself(t *testing.T) {
	w := &fakeWallets{}
	svc, repo, _ := newResSvc(paidAt(day(11)), w, nil, tieredHotel()) // 9h out: non-refundable for the guest
	if _, err := svc.Cancel(context.Background(), ownerP, 5); err != nil {
		t.Fatal(err)
	}
	if repo.x.FeeAmount != "0.0000" || repo.x.RefundAmount != "5000001.0000" || len(w.fees) != 0 {
		t.Fatalf("hotel-initiated: fee %s refund %s fees %v", repo.x.FeeAmount, repo.x.RefundAmount, w.fees)
	}
}

func TestAFeeThatCannotBeCollectedDoesNotBlockTheCancellation(t *testing.T) {
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}, feeErr: domain.ErrInsufficientFunds}
	svc, repo, _ := newResSvc(paidAt(day(12)), w, nil, tieredHotel())
	if _, err := svc.Cancel(context.Background(), guestP, 5); err != nil {
		t.Fatal(err)
	}
	if repo.x.Status != domain.ReservationRefunded || repo.x.FeeAmount != "0.0000" || repo.x.RefundAmount != "5000001.0000" {
		t.Fatalf("fee waived: %+v", repo.x)
	}
	// A wallet that is merely unreachable is a failure to retry, not to waive.
	w2 := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}, feeErr: domain.ErrUpstreamUnavailable}
	svc, repo, _ = newResSvc(paidAt(day(12)), w2, nil, tieredHotel())
	if _, err := svc.Cancel(context.Background(), guestP, 5); !errors.Is(err, domain.ErrUpstreamUnavailable) || repo.x.Status != domain.ReservationPaid {
		t.Fatalf("unreachable wallet: %v status %v", err, repo.x.Status)
	}
}

func TestCancellationQuote(t *testing.T) {
	svc, _, _ := newResSvc(paidAt(day(12)), &fakeWallets{}, nil, tieredHotel())
	q, err := svc.CancellationQuote(context.Background(), guestP, 5)
	if err != nil || q.FeePercent != 50 || q.Fee != "2500001.0000" || q.Refund != "2500000.0000" || !q.Cancellable || len(q.Policy) != 2 {
		t.Fatalf("quote: %+v %v", q, err)
	}
	// The manager would pay no fee; a stranger cannot ask.
	if q, _ := svc.CancellationQuote(context.Background(), ownerP, 5); q.FeePercent != 0 || q.Refund != "5000001.0000" {
		t.Fatalf("manager quote: %+v", q)
	}
	if _, err := svc.CancellationQuote(context.Background(), otherP, 5); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("stranger: %v", err)
	}
	svc, _, _ = newResSvc(paidAt(day(11)), &fakeWallets{}, nil, tieredHotel())
	if q, _ := svc.CancellationQuote(context.Background(), guestP, 5); q.Cancellable || q.Reason == "" || q.FeePercent != 100 {
		t.Fatalf("non-refundable quote: %+v", q)
	}
	svc, _, _ = newResSvc(pending(), &fakeWallets{}, nil, tieredHotel())
	if q, _ := svc.CancellationQuote(context.Background(), guestP, 5); !q.Cancellable || q.Fee != "0.0000" {
		t.Fatalf("pending quote: %+v", q)
	}
}
