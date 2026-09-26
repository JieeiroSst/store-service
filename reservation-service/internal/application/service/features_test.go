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

// ---- promo codes

func TestPromoCodeIsCheckedBeforeTheRoomsAreTouched(t *testing.T) {
	from, to := day(1), day(31)
	svc, repo, hotels := rushSvc(ReservationOptions{})
	hotels.promos = []domain.Promotion{
		{ID: 1, HotelID: 1, Code: "SPRING15", PercentOff: 15, MinNights: 2, Active: true, ValidFrom: &from, ValidTo: &to},
		{ID: 2, HotelID: 1, Code: "OLD", PercentOff: 10, MinNights: 1, Active: false},
		{ID: 3, HotelID: 1, Code: "GONE", AmountOff: "500000", MinNights: 1, Active: true, MaxUses: 5, UsedCount: 5},
		{ID: 4, HotelID: 1, Code: "LONGSTAY", PercentOff: 20, MinNights: 7, Active: true},
	}
	cmd := rushCmd             // 2 nights
	cmd.PromoCode = "spring15" // codes are matched without regard to case
	if _, err := svc.Reserve(context.Background(), guestP, cmd); err != nil {
		t.Fatal(err)
	}
	if repo.reserved.Promo == nil || repo.reserved.Promo.ID != 1 {
		t.Fatalf("the promo must be passed to the repository: %+v", repo.reserved.Promo)
	}
	for name, code := range map[string]string{"unknown": "NOPE", "switched off": "OLD", "used up": "GONE", "too short a stay": "LONGSTAY"} {
		repo.reserved = nil
		c := rushCmd
		c.PromoCode = code
		if _, err := svc.Reserve(context.Background(), guestP, c); !errors.Is(err, domain.ErrInvalid) || repo.reserved != nil {
			t.Errorf("%s: err %v, repository called: %v", name, err, repo.reserved != nil)
		}
	}
	// A code that is not valid for the check-in date is refused.
	c := rushCmd
	c.PromoCode, c.Start, c.End = "SPRING15", day(30), day(33) // starts inside; ends outside is fine, starting after valid_to is not
	if _, err := svc.Reserve(context.Background(), guestP, c); err != nil {
		t.Fatalf("check-in inside the window: %v", err)
	}
	c.Start, c.End = day(31).AddDate(0, 0, 2), day(31).AddDate(0, 0, 4)
	if _, err := svc.Reserve(context.Background(), guestP, c); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("check-in after valid_to: %v", err)
	}
}

func TestPromoCatalogueIsTheManagersOnly(t *testing.T) {
	ctx := context.Background()
	svc, repo := newCatalog(activeHotel())
	in := inbound.PromotionInput{Code: "SUMMER10", PercentOff: 10}
	if _, err := svc.CreatePromotion(ctx, otherP, 1, in); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other manager: %v", err)
	}
	if _, err := svc.Promotions(ctx, guestP, 1); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a guest must not read the codes: %v", err)
	}
	p, err := svc.CreatePromotion(ctx, ownerP, 1, in)
	if err != nil || !p.Active || p.MinNights != 1 || len(repo.promos) != 1 {
		t.Fatalf("create: %+v %v", p, err)
	}
	amt := "1000000"
	for name, bad := range map[string]inbound.PromotionInput{
		"no discount":           {Code: "ABC"},
		"both discounts":        {Code: "ABC", PercentOff: 10, AmountOff: amt},
		"over 100%":             {Code: "ABC", PercentOff: 101},
		"negative amount":       {Code: "ABC", AmountOff: "-5"},
		"short code":            {Code: "A", PercentOff: 5},
		"code with spaces":      {Code: "SUMMER 10", PercentOff: 5},
		"ends before it starts": {Code: "ABC", PercentOff: 5, ValidFrom: &[]time.Time{day(20)}[0], ValidTo: &[]time.Time{day(10)}[0]},
	} {
		if _, err := svc.CreatePromotion(ctx, ownerP, 1, bad); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	// max_uses cannot drop below the uses already taken.
	repo.promos = []domain.Promotion{{ID: 9, HotelID: 1, Code: "USED", PercentOff: 5, MinNights: 1, Active: true, MaxUses: 10, UsedCount: 6}}
	if _, err := svc.UpdatePromotion(ctx, ownerP, 1, 9, inbound.PromotionInput{Code: "USED", PercentOff: 5, MaxUses: 3}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("max_uses below used_count: %v", err)
	}
	if _, err := svc.UpdatePromotion(ctx, ownerP, 1, 404, in); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown promo: %v", err)
	}
}

// ---- the front desk

func TestFrontDeskSteps(t *testing.T) {
	// "today" in newResSvc is Jan 10.
	x := paidAt(day(10))
	x.End = day(12)
	svc, repo, _ := newResSvc(x, &fakeWallets{}, nil, activeHotel())
	ctx := context.Background()

	if _, err := svc.CheckIn(ctx, guestP, 5); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("guests cannot check themselves in: %v", err)
	}
	if _, err := svc.CheckIn(ctx, otherP, 5); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("another manager: %v", err)
	}
	if _, err := svc.CheckOut(ctx, ownerP, 5); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("checking out before checking in: %v", err)
	}
	got, err := svc.CheckIn(ctx, ownerP, 5)
	if err != nil || got.Stay != domain.StayCheckedIn || repo.stayActor != 50 {
		t.Fatalf("check in: %+v %v", got, err)
	}
	if _, err := svc.CheckIn(ctx, ownerP, 5); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("checking in twice: %v", err)
	}
	inbox := withInbox(svc)
	if got, err = svc.CheckOut(ctx, ownerP, 5); err != nil || got.Stay != domain.StayCheckedOut {
		t.Fatalf("check out: %+v %v", got, err)
	}
	if k := inbox.kinds(7); len(k) != 1 || k[0] != "reservation.checked_out" {
		t.Fatalf("the guest is thanked and invited to review: %v", k)
	}
}

func TestCheckInWindowAndNoShow(t *testing.T) {
	ctx := context.Background()
	// The stay starts tomorrow: too early to check in, too early to call it a no-show.
	svc, _, _ := newResSvc(paidAt(day(11)), &fakeWallets{}, nil, activeHotel())
	if _, err := svc.CheckIn(ctx, ownerP, 5); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("early check-in: %v", err)
	}
	if _, err := svc.NoShow(ctx, ownerP, 5); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("early no-show: %v", err)
	}
	// It started yesterday and the guest never came.
	late := paidAt(day(9))
	late.End = day(12)
	svc, _, _ = newResSvc(late, &fakeWallets{}, nil, activeHotel())
	if got, err := svc.NoShow(ctx, ownerP, 5); err != nil || got.Stay != domain.StayNoShow {
		t.Fatalf("no-show: %+v %v", got, err)
	}
	// The stay is over: it is too late to check in.
	over := paidAt(day(5))
	over.End = day(8)
	svc, _, _ = newResSvc(over, &fakeWallets{}, nil, activeHotel())
	if _, err := svc.CheckIn(ctx, ownerP, 5); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("check-in after the stay: %v", err)
	}
	// Only paid reservations have a stay.
	svc, _, _ = newResSvc(pending(), &fakeWallets{}, nil, activeHotel())
	if _, err := svc.CheckIn(ctx, ownerP, 5); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("pending: %v", err)
	}
}

func TestHistoryNamesWhoActedWithoutLeakingUserIDs(t *testing.T) {
	svc, repo, _ := newResSvc(pending(), &fakeWallets{}, nil, activeHotel())
	repo.events = []domain.ReservationEvent{
		{Actor: 7, Event: "created", ToStatus: domain.ReservationPending},
		{Actor: 7, Event: "paid", FromStatus: domain.ReservationPending, ToStatus: domain.ReservationPaid},
		{Actor: 50, Event: "checked_in"},
		{Actor: 0, Event: "canceled", Note: "hold expired"},
	}
	h, err := svc.History(context.Background(), guestP, 5)
	if err != nil || len(h) != 4 {
		t.Fatalf("history: %v %v", h, err)
	}
	if h[0].By != "guest" || h[2].By != "hotel" || h[3].By != "system" || h[1].From != domain.ReservationPending {
		t.Fatalf("who: %+v", h)
	}
	if _, err := svc.History(context.Background(), otherP, 5); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger: %v", err)
	}
	if _, err := svc.History(context.Background(), ownerP, 5); err != nil {
		t.Fatalf("the hotel's manager: %v", err)
	}
}

// ---- report, wishlist, search filters

func TestReportIsTheManagersAndBounded(t *testing.T) {
	ctx := context.Background()
	svc, repo := newCatalog(activeHotel())
	if _, err := svc.Report(ctx, guestP, 1, day(1), day(8)); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("guest: %v", err)
	}
	if _, err := svc.Report(ctx, ownerP, 1, day(1), day(8)); err != nil || !repo.reportFrom.Equal(day(1)) {
		t.Fatalf("manager: %v", err)
	}
	if _, err := svc.Report(ctx, ownerP, 1, day(8), day(8)); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("empty range: %v", err)
	}
	if _, err := svc.Report(ctx, ownerP, 1, day(1), day(1).AddDate(0, 0, 93)); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("a year of data in one go: %v", err)
	}
}

type memWishlist struct {
	outbound.WishlistRepository
	added []int64
}

func (m *memWishlist) Add(_ context.Context, _, hotel int64) error {
	m.added = append(m.added, hotel)
	return nil
}
func (m *memWishlist) List(context.Context, int64, int64, int) ([]domain.Hotel, error) {
	return []domain.Hotel{{ID: 9}, {ID: 8}, {ID: 7}}, nil
}

func TestOnlyPublishedHotelsCanBeSaved(t *testing.T) {
	ctx := context.Background()
	repo := &fakeHotels{hotel: activeHotel()}
	wl := &memWishlist{}
	svc := NewWishlistService(wl, repo)
	if err := svc.Add(ctx, guestP, 1); err != nil || len(wl.added) != 1 {
		t.Fatalf("published: %v", err)
	}
	repo.hotel.Status = domain.HotelPending
	if err := svc.Add(ctx, guestP, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("pending: %v", err)
	}
	pg, err := svc.List(ctx, guestP, 0, 2)
	if err != nil || len(pg.Items) != 2 || pg.NextID != 8 {
		t.Fatalf("page: %+v %v", pg, err)
	}
}

func TestSearchFiltersAreValidated(t *testing.T) {
	svc := NewHotelService(&fakeHotels{}, nil, nil)
	ctx := context.Background()
	for name, f := range map[string]domain.HotelFilter{
		"negative rate": {MinRate: "-1"},
		"junk rate":     {MaxRate: "cheap"},
	} {
		if _, err := svc.Search(ctx, f); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	many := make([]string, 21)
	if _, err := svc.Search(ctx, domain.HotelFilter{Amenities: many}); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("too many amenities: %v", err)
	}
	repo := &fakeHotels{}
	NewHotelService(repo, nil, nil).Search(ctx, domain.HotelFilter{Amenities: []string{"Spa"}, MinRate: "1000000", MaxRate: "9000000", Sort: "rating"})
	if f := repo.listed; len(f.Amenities) != 1 || f.MinRate != "1000000" || f.MaxRate != "9000000" {
		t.Fatalf("filters must reach the repository: %+v", f)
	}
}
