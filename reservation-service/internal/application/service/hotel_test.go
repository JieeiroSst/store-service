package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

func ptr(s string) *string { return &s }

func hotelInput(w *string) inbound.HotelInput {
	return inbound.HotelInput{Name: "Grand Palace", Stars: 5, City: "Da Nang", Address: "1 Beach Rd",
		PhoneNumber: "+84 236 123 456", Currency: "vnd", WalletID: w}
}

func TestOnlyFiveStarHotelsAreRegistered(t *testing.T) {
	svc := NewHotelService(&fakeHotels{}, nil, nil)
	for _, stars := range []int{0, 3, 4, 6} {
		in := hotelInput(nil)
		in.Stars = stars
		if _, err := svc.Register(context.Background(), ownerP, in); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%d stars: %v", stars, err)
		}
	}
	if _, err := svc.Register(context.Background(), ownerP, hotelInput(nil)); err != nil {
		t.Fatalf("5 stars: %v", err)
	}
}

func TestRegistrationNeedsVerificationUnlessAdmin(t *testing.T) {
	for _, tc := range []struct {
		who  inbound.Principal
		want domain.HotelStatus
	}{{ownerP, domain.HotelPending}, {adminP, domain.HotelActive}} {
		repo := &fakeHotels{}
		h, err := NewHotelService(repo, nil, nil).Register(context.Background(), tc.who, hotelInput(nil))
		if err != nil || h.Status != tc.want || h.OwnerID != tc.who.UserID {
			t.Fatalf("%+v: status=%v owner=%d err=%v", tc.who, h.Status, h.OwnerID, err)
		}
		if h.Currency != "VND" || h.CheckInTime != "14:00" || h.CheckOutTime != "12:00" || h.FreeCancelHours != 48 {
			t.Fatalf("defaults: %+v", h)
		}
	}
	in := hotelInput(nil)
	in.Status = domain.HotelActive
	if _, err := NewHotelService(&fakeHotels{}, nil, nil).Register(context.Background(), ownerP, in); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("a manager cannot publish themself: %v", err)
	}
}

func TestHotelValidation(t *testing.T) {
	svc := NewHotelService(&fakeHotels{}, nil, nil)
	for name, mod := range map[string]func(*inbound.HotelInput){
		"no name":     func(i *inbound.HotelInput) { i.Name = " " },
		"no city":     func(i *inbound.HotelInput) { i.City = "" },
		"no address":  func(i *inbound.HotelInput) { i.Address = "" },
		"no phone":    func(i *inbound.HotelInput) { i.PhoneNumber = "" },
		"bad phone":   func(i *inbound.HotelInput) { i.PhoneNumber = "call me" },
		"no currency": func(i *inbound.HotelInput) { i.Currency = "" },
		"bad time":    func(i *inbound.HotelInput) { i.CheckInTime = "25:00" },
		"bad window":  func(i *inbound.HotelInput) { n := -1; i.FreeCancelHours = &n },
	} {
		in := hotelInput(nil)
		mod(&in)
		if _, err := svc.Register(context.Background(), ownerP, in); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestHotelWalletChecks(t *testing.T) {
	tests := []struct {
		name    string
		wallets *fakeWallets // nil = wallet payment disabled
		who     inbound.Principal
		wallet  *string
		want    error
	}{
		{"none given", &fakeWallets{}, ownerP, nil, nil},
		{"own wallet", &fakeWallets{ownerID: 50}, ownerP, ptr("w-ok"), nil},
		{"someone else's wallet", &fakeWallets{ownerID: 99}, ownerP, ptr("w-x"), domain.ErrForbidden},
		{"admin may set any", &fakeWallets{ownerID: 99}, adminP, ptr("w-x"), nil},
		{"unknown wallet", &fakeWallets{missing: map[string]bool{"w-x": true}}, ownerP, ptr("w-x"), domain.ErrInvalid},
		{"frozen wallet", &fakeWallets{ownerID: 50, status: map[string]string{"w-f": "FROZEN"}}, ownerP, ptr("w-f"), domain.ErrInvalid},
		{"wallets disabled", nil, ownerP, ptr("w-ok"), domain.ErrInvalid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeHotels{}
			var wg outbound.WalletGateway
			if tc.wallets != nil {
				wg = tc.wallets
			}
			_, err := NewHotelService(repo, wg, nil).Register(context.Background(), tc.who, hotelInput(tc.wallet))
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if tc.want != nil && repo.saved != nil {
				t.Fatal("nothing must be stored")
			}
			if tc.want == nil && tc.wallet != nil && repo.saved.WalletID != *tc.wallet {
				t.Fatalf("saved wallet = %q", repo.saved.WalletID)
			}
		})
	}
}

func TestUpdateOwnershipAndKeepingWhatIsNotMentioned(t *testing.T) {
	repo := &fakeHotels{hotel: domain.Hotel{ID: 1, OwnerID: 50, Status: domain.HotelActive, Currency: "VND",
		WalletID: "w-old", CheckInTime: "15:00", CheckOutTime: "11:00", FreeCancelHours: 72}}
	svc := NewHotelService(repo, &fakeWallets{ownerID: 50}, nil)
	ctx := context.Background()

	// Wallet, times, window and status are kept when not mentioned.
	if _, err := svc.Update(ctx, ownerP, 1, hotelInput(nil)); err != nil {
		t.Fatal(err)
	}
	s := repo.saved
	if s.WalletID != "w-old" || s.Status != domain.HotelActive || s.CheckInTime != "15:00" || s.CheckOutTime != "11:00" || s.FreeCancelHours != 72 {
		t.Fatalf("kept: %+v", s)
	}
	// The manager can pause and resume, and change the wallet to one of their own.
	in := hotelInput(ptr("w-new"))
	in.Status = domain.HotelInactive
	if _, err := svc.Update(ctx, ownerP, 1, in); err != nil || repo.saved.Status != domain.HotelInactive || repo.saved.WalletID != "w-new" {
		t.Fatalf("pause+wallet: %v %+v", err, repo.saved)
	}
	// Currency is fixed once rates and payments exist.
	in = hotelInput(nil)
	in.Currency = "USD"
	if _, err := svc.Update(ctx, ownerP, 1, in); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("currency change: %v", err)
	}
	// Someone else's hotel: a published one is forbidden, an unpublished one does not exist.
	if _, err := svc.Update(ctx, otherP, 1, hotelInput(nil)); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other user: %v", err)
	}
	repo.hotel.Status = domain.HotelPending
	if _, err := svc.Update(ctx, otherP, 1, hotelInput(nil)); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("other user, unpublished: %v", err)
	}
	// Admins can edit anything.
	repo.hotel.Status = domain.HotelActive
	if _, err := svc.Update(ctx, adminP, 1, hotelInput(nil)); err != nil {
		t.Fatalf("admin: %v", err)
	}
}

func TestVerificationWorkflow(t *testing.T) {
	ctx := context.Background()
	repo := &fakeHotels{hotel: domain.Hotel{ID: 1, OwnerID: 50, Status: domain.HotelPending, Currency: "VND"}}
	svc := NewHotelService(repo, nil, nil)

	if _, err := svc.Approve(ctx, ownerP, 1); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("manager verifying themself: %v", err)
	}
	if _, err := svc.Reject(ctx, adminP, 1, "  "); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("reject needs a reason: %v", err)
	}
	if _, err := svc.Reject(ctx, adminP, 1, "not a 5-star property"); err != nil || repo.review.status != domain.HotelRejected || repo.review.note == "" {
		t.Fatalf("reject: %v %+v", err, repo.review)
	}
	// A pending hotel cannot be switched on by its manager; a rejected one is resubmitted by fixing it.
	in := hotelInput(nil)
	in.Status = domain.HotelActive
	if _, err := svc.Update(ctx, ownerP, 1, in); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("switch on before verification: %v", err)
	}
	repo.hotel.Status = domain.HotelRejected
	if _, err := svc.Update(ctx, ownerP, 1, hotelInput(nil)); err != nil || repo.saved.Status != domain.HotelPending {
		t.Fatalf("resubmit: %v %+v", err, repo.saved)
	}
	if _, err := svc.Update(ctx, adminP, 1, hotelInput(nil)); err != nil || repo.saved.Status != domain.HotelRejected {
		t.Fatalf("an admin's edit must not resubmit for the manager: %v %+v", err, repo.saved)
	}
	if _, err := svc.Approve(ctx, adminP, 1); err != nil || repo.review.status != domain.HotelActive || repo.review.note != "" {
		t.Fatalf("approve: %v %+v", err, repo.review)
	}
	repo.hotel.Status = domain.HotelActive
	if _, err := svc.Approve(ctx, adminP, 1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("approve twice: %v", err)
	}
	// Pausing is only for published hotels (otherwise pause-then-resume would skip verification).
	repo.hotel.Status = domain.HotelPending
	if err := svc.Deactivate(ctx, ownerP, 1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("pause pending: %v", err)
	}
}

func TestViewHidesUnpublishedFromOthers(t *testing.T) {
	ctx := context.Background()
	repo := &fakeHotels{hotel: domain.Hotel{ID: 1, OwnerID: 50, Status: domain.HotelPending,
		RoomTypes: []domain.RoomType{{ID: 1, Active: true}, {ID: 2, Active: false}}}}
	svc := NewHotelService(repo, nil, nil)
	if _, err := svc.View(ctx, nil, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("anonymous: %v", err)
	}
	if _, err := svc.View(ctx, &otherP, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("other user: %v", err)
	}
	if h, err := svc.View(ctx, &ownerP, 1); err != nil || len(h.RoomTypes) != 2 {
		t.Fatalf("manager sees everything: %v %d", err, len(h.RoomTypes))
	}
	repo.hotel.Status = domain.HotelActive
	if h, err := svc.View(ctx, nil, 1); err != nil || len(h.RoomTypes) != 1 {
		t.Fatalf("the public sees active room types only: %v %d", err, len(h.RoomTypes))
	}
}

func TestMineAndQueue(t *testing.T) {
	ctx := context.Background()
	repo := &fakeHotels{}
	svc := NewHotelService(repo, nil, nil)
	svc.Mine(ctx, ownerP, nil, 0)
	if f := repo.listed; f.OwnerID != 50 || !f.IncludeInactive || f.Limit != defaultPageSize+1 || f.Sort != "newest" {
		t.Fatalf("mine: %+v", f)
	}
	if _, err := svc.ForReview(ctx, ownerP, 0, nil, 0); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("queue for non-admin: %v", err)
	}
	svc.ForReview(ctx, adminP, 0, nil, 0)
	if f := repo.listed; f.Status != domain.HotelPending || f.Sort != "oldest" {
		t.Fatalf("queue defaults to pending, oldest first: %+v", f)
	}
	// A cursor from another list is refused.
	if _, err := svc.ForReview(ctx, adminP, 0, &domain.HotelCursor{Sort: "newest", ID: 3}, 0); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("cursor of another list: %v", err)
	}
	// One extra row means there is a next page, and the cursor is the last item shown.
	repo.list = []domain.Hotel{{ID: 3}, {ID: 2}, {ID: 1}}
	pg, err := svc.Mine(ctx, ownerP, nil, 2)
	if err != nil || len(pg.Items) != 2 || pg.Next == nil || pg.Next.ID != 2 || pg.Next.Sort != "newest" {
		t.Fatalf("mine page: %+v %v", pg, err)
	}
}

// ---- search and cursor pagination

func catalogue() []domain.Hotel {
	// Recommended order with the tiers below: 6, 4, 2, 7, 5, 3, 1.
	return []domain.Hotel{
		{ID: 1, OwnerID: 10, Rating: 3.0, ReviewCount: 10},
		{ID: 2, OwnerID: 10, Rating: 4.5, ReviewCount: 30},
		{ID: 3, OwnerID: 10, Rating: 3.5, ReviewCount: 10},
		{ID: 4, OwnerID: 10, Rating: 4.8, ReviewCount: 40},
		{ID: 5, OwnerID: 10, Rating: 4.0, ReviewCount: 20},
		{ID: 6, OwnerID: 20, Rating: 4.9, ReviewCount: 50}, // platinum manager on top
		{ID: 7, OwnerID: 10, Rating: 4.2, ReviewCount: 25},
	}
}

func hotelIDs(hs []domain.Hotel) []int64 {
	out := make([]int64, len(hs))
	for i, h := range hs {
		out[i] = h.ID
	}
	return out
}

func TestRecommendedPagesWalkTheWholeRankingOnce(t *testing.T) {
	repo := &fakeHotels{list: catalogue()}
	svc := NewHotelService(repo, nil, NewManagerTiers(&fakeLoyalty{tiers: map[int64]int{10: 1, 20: 4}}, 1<<40))
	ctx := context.Background()

	whole, err := svc.Search(ctx, domain.HotelFilter{Limit: 100})
	if err != nil || whole.Next != nil {
		t.Fatalf("whole: %v %v", err, whole.Next)
	}
	var got []int64
	var cur *domain.HotelCursor
	for pages := 0; ; pages++ {
		pg, err := svc.Search(ctx, domain.HotelFilter{Limit: 3, After: cur})
		if err != nil || pages > 10 {
			t.Fatalf("%v (pages %d)", err, pages)
		}
		got = append(got, hotelIDs(pg.Items)...)
		if pg.Next == nil {
			break
		}
		cur = pg.Next
	}
	if !reflect.DeepEqual(got, hotelIDs(whole.Items)) || len(got) != 7 || got[0] != 6 {
		t.Fatalf("paged %v, whole %v", got, hotelIDs(whole.Items))
	}
}

func TestCursorSurvivesNewHotels(t *testing.T) {
	repo := &fakeHotels{list: catalogue()}
	svc := NewHotelService(repo, nil, nil)
	first, _ := svc.Search(context.Background(), domain.HotelFilter{Limit: 2})
	repo.list = append(repo.list, domain.Hotel{ID: 99, OwnerID: 10, Rating: 5, ReviewCount: 200}) // appears between two requests
	second, err := svc.Search(context.Background(), domain.HotelFilter{Limit: 2, After: first.Next})
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range second.Items {
		if h.ID == 99 {
			t.Fatalf("the new hotel must not jump back into page 2: %v", hotelIDs(second.Items))
		}
		for _, seen := range first.Items {
			if h.ID == seen.ID {
				t.Fatalf("page 2 repeats page 1: %v %v", hotelIDs(second.Items), hotelIDs(first.Items))
			}
		}
	}
}

func TestDatabaseSortsPassCursorAndFetchOneExtra(t *testing.T) {
	repo := &fakeHotels{list: []domain.Hotel{{ID: 9, Rating: 4.5, ReviewCount: 3}, {ID: 8, Rating: 4.5, ReviewCount: 2}, {ID: 7, Rating: 4.0, ReviewCount: 9}}}
	svc := NewHotelService(repo, nil, nil)
	ctx := context.Background()
	pg, err := svc.Search(ctx, domain.HotelFilter{Sort: "rating", Limit: 2})
	if err != nil || repo.listed.Limit != 3 || len(pg.Items) != 2 {
		t.Fatalf("one extra row tells whether there is more: %v limit=%d items=%d", err, repo.listed.Limit, len(pg.Items))
	}
	want := &domain.HotelCursor{Sort: "rating", ID: 8, Rating: 4.5, Count: 2}
	if !reflect.DeepEqual(pg.Next, want) {
		t.Fatalf("next = %+v, want %+v", pg.Next, want)
	}
	if _, err := svc.Search(ctx, domain.HotelFilter{Sort: "rating", Limit: 2, After: pg.Next}); err != nil || !reflect.DeepEqual(repo.listed.After, want) {
		t.Fatalf("cursor must reach the repository: %v %+v", err, repo.listed.After)
	}
	repo.list = repo.list[:2]
	if pg, _ = svc.Search(ctx, domain.HotelFilter{Sort: "newest", Limit: 2}); pg.Next != nil {
		t.Fatalf("no next page expected: %+v", pg.Next)
	}
}

func TestSearchValidation(t *testing.T) {
	svc := NewHotelService(&fakeHotels{list: catalogue()}, nil, nil)
	ctx := context.Background()
	for name, f := range map[string]domain.HotelFilter{
		"cursor for another sort": {Sort: "newest", After: &domain.HotelCursor{Sort: "rating", ID: 1}},
		"default sort vs newest":  {After: &domain.HotelCursor{Sort: "newest", ID: 1}},
		"unknown sort":            {Sort: "price"},
	} {
		if _, err := svc.Search(ctx, f); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	// The public can never widen the search to unpublished hotels or another manager's list.
	repo := &fakeHotels{}
	NewHotelService(repo, nil, nil).Search(ctx, domain.HotelFilter{OwnerID: 5, Status: domain.HotelPending, IncludeInactive: true})
	if f := repo.listed; f.OwnerID != 0 || f.Status != 0 || f.IncludeInactive {
		t.Fatalf("filter was not sanitised: %+v", f)
	}
}
