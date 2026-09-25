package service

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

// recordingHomestays is a one-homestay repository that records what the service stores.
type recordingHomestays struct {
	outbound.HomestayRepository
	current domain.Homestay
	saved   *domain.Homestay
	review  *reviewCall
	listed  *domain.HomestayFilter
}

type reviewCall struct {
	status domain.HomestayStatus
	note   string
}

func (r *recordingHomestays) Get(context.Context, int64) (domain.Homestay, error) {
	return r.current, nil
}
func (r *recordingHomestays) Create(_ context.Context, h domain.Homestay, actor int64) (domain.Homestay, error) {
	h.HostID = actor
	r.saved = &h
	return h, nil
}
func (r *recordingHomestays) Update(_ context.Context, h domain.Homestay, _ int64) (domain.Homestay, error) {
	r.saved = &h
	return h, nil
}
func (r *recordingHomestays) SetStatus(_ context.Context, _ int64, s domain.HomestayStatus, _ int64) error {
	r.saved = &domain.Homestay{Status: s}
	return nil
}
func (r *recordingHomestays) SetReview(_ context.Context, _ int64, s domain.HomestayStatus, note string, _ int64) error {
	r.review = &reviewCall{s, note}
	return nil
}
func (r *recordingHomestays) List(_ context.Context, f domain.HomestayFilter) ([]domain.Homestay, error) {
	r.listed = &f
	return nil, nil
}

func ptr(s string) *string { return &s }

var (
	adminP = inbound.Principal{UserID: 1, Admin: true}
	ownerP = inbound.Principal{UserID: 50}
	otherP = inbound.Principal{UserID: 60}
)

func input(w *string) inbound.HomestayInput {
	return inbound.HomestayInput{Name: "Villa", Type: 1, Address: "1 Beach Rd", PhoneNumber: "+84 901 234 567", Guests: 2, WalletID: w}
}

func TestRegistrationNeedsApprovalUnlessAdmin(t *testing.T) {
	for _, tc := range []struct {
		who  inbound.Principal
		want domain.HomestayStatus
	}{{ownerP, domain.HomestayPending}, {adminP, domain.HomestayActive}} {
		repo := &recordingHomestays{}
		h, err := NewHomestayService(repo, nil, nil).Create(context.Background(), tc.who, input(nil))
		if err != nil || h.Status != tc.want || h.HostID != tc.who.UserID {
			t.Fatalf("%+v: status=%v host=%d err=%v", tc.who, h.Status, h.HostID, err)
		}
	}
	// A registrant cannot publish themself by sending a status.
	in := input(nil)
	in.Status = domain.HomestayActive
	if _, err := NewHomestayService(&recordingHomestays{}, nil, nil).Create(context.Background(), ownerP, in); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("self-published: %v", err)
	}
}

func TestRegistrationValidation(t *testing.T) {
	svc := NewHomestayService(&recordingHomestays{}, nil, nil)
	for name, mod := range map[string]func(*inbound.HomestayInput){
		"no type":     func(i *inbound.HomestayInput) { i.Type = 0 },
		"bad type":    func(i *inbound.HomestayInput) { i.Type = 4 },
		"no address":  func(i *inbound.HomestayInput) { i.Address = "  " },
		"no phone":    func(i *inbound.HomestayInput) { i.PhoneNumber = "" },
		"letters":     func(i *inbound.HomestayInput) { i.PhoneNumber = "call me maybe" },
		"no name":     func(i *inbound.HomestayInput) { i.Name = "" },
		"zero guests": func(i *inbound.HomestayInput) { i.Guests = 0 },
	} {
		in := input(nil)
		mod(&in)
		if _, err := svc.Create(context.Background(), ownerP, in); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	for _, typ := range []int{1, 2, 3} { // homestay, hotel, rental house
		in := input(nil)
		in.Type = typ
		if _, err := svc.Create(context.Background(), ownerP, in); err != nil {
			t.Errorf("type %d: %v", typ, err)
		}
	}
}

func TestHomestayWalletChecks(t *testing.T) {
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
			repo := &recordingHomestays{}
			var wg outbound.WalletGateway
			if tc.wallets != nil {
				wg = tc.wallets
			}
			_, err := NewHomestayService(repo, wg, nil).Create(context.Background(), tc.who, input(tc.wallet))
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

func TestUpdateOwnershipAndWallet(t *testing.T) {
	repo := &recordingHomestays{current: domain.Homestay{ID: 1, HostID: 50, Status: domain.HomestayActive, WalletID: "w-old"}}
	svc := NewHomestayService(repo, &fakeWallets{ownerID: 50}, nil)
	ctx := context.Background()

	// Not mentioned: wallet and status are kept, so editing the address cannot cut off payments.
	if _, err := svc.Update(ctx, ownerP, 1, input(nil)); err != nil || repo.saved.WalletID != "w-old" || repo.saved.Status != domain.HomestayActive {
		t.Fatalf("keep: %v %+v", err, repo.saved)
	}
	// The owner can pause and resume.
	in := input(nil)
	in.Status = domain.HomestayInactive
	if _, err := svc.Update(ctx, ownerP, 1, in); err != nil || repo.saved.Status != domain.HomestayInactive {
		t.Fatalf("pause: %v", err)
	}
	// Changing the wallet is checked against the OWNER, even when an admin does it.
	if _, err := svc.Update(ctx, ownerP, 1, input(ptr("w-new"))); err != nil || repo.saved.WalletID != "w-new" {
		t.Fatalf("change: %v", err)
	}
	// Someone else's homestay: visible ones say forbidden, unpublished ones do not exist.
	if _, err := svc.Update(ctx, otherP, 1, input(nil)); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other user: %v", err)
	}
	repo.current.Status = domain.HomestayPending
	if _, err := svc.Update(ctx, otherP, 1, input(nil)); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("other user, unpublished: %v", err)
	}
	// Admins can edit anything.
	repo.current.Status = domain.HomestayActive
	if _, err := svc.Update(ctx, adminP, 1, input(nil)); err != nil {
		t.Fatalf("admin: %v", err)
	}
	// Only "active"/"inactive" may be requested.
	in.Status = domain.HomestayPending
	if _, err := svc.Update(ctx, ownerP, 1, in); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("pending via update: %v", err)
	}
}

func TestApprovalWorkflow(t *testing.T) {
	ctx := context.Background()
	repo := &recordingHomestays{current: domain.Homestay{ID: 1, HostID: 50, Status: domain.HomestayPending}}
	svc := NewHomestayService(repo, nil, nil)

	if _, err := svc.Approve(ctx, ownerP, 1); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("owner approving themself: %v", err)
	}
	if _, err := svc.Reject(ctx, adminP, 1, "  "); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("reject needs a reason: %v", err)
	}
	if _, err := svc.Reject(ctx, adminP, 1, "photos are of another property"); err != nil || repo.review.status != domain.HomestayRejected || repo.review.note == "" {
		t.Fatalf("reject: %v %+v", err, repo.review)
	}

	// A pending listing cannot be switched on by its owner, and a rejected one is resubmitted by fixing it.
	in := input(nil)
	in.Status = domain.HomestayActive
	if _, err := svc.Update(ctx, ownerP, 1, in); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("switch on before approval: %v", err)
	}
	repo.current.Status = domain.HomestayRejected
	if _, err := svc.Update(ctx, ownerP, 1, input(nil)); err != nil || repo.saved.Status != domain.HomestayPending {
		t.Fatalf("resubmit: %v %+v", err, repo.saved)
	}
	if _, err := svc.Update(ctx, adminP, 1, input(nil)); err != nil || repo.saved.Status != domain.HomestayRejected {
		t.Fatalf("an admin's edit must not resubmit for the owner: %v %+v", err, repo.saved)
	}

	if _, err := svc.Approve(ctx, adminP, 1); err != nil || repo.review.status != domain.HomestayActive || repo.review.note != "" {
		t.Fatalf("approve: %v %+v", err, repo.review)
	}
	// Already published: neither approve nor reject applies.
	repo.current.Status = domain.HomestayActive
	if _, err := svc.Approve(ctx, adminP, 1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("approve twice: %v", err)
	}
	if _, err := svc.Reject(ctx, adminP, 1, "late"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("reject published: %v", err)
	}
}

func TestDeactivateOnlyPublished(t *testing.T) {
	ctx := context.Background()
	repo := &recordingHomestays{current: domain.Homestay{ID: 1, HostID: 50, Status: domain.HomestayPending}}
	svc := NewHomestayService(repo, nil, nil)
	// Pausing a pending listing would let its owner switch it on later and skip approval.
	if err := svc.Deactivate(ctx, ownerP, 1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("pending: %v", err)
	}
	repo.current.Status = domain.HomestayActive
	if err := svc.Deactivate(ctx, ownerP, 1); err != nil || repo.saved.Status != domain.HomestayInactive {
		t.Fatalf("active: %v", err)
	}
	if err := svc.Deactivate(ctx, otherP, 1); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other: %v", err)
	}
}

func TestViewHidesUnpublishedFromOthers(t *testing.T) {
	ctx := context.Background()
	repo := &recordingHomestays{current: domain.Homestay{ID: 1, HostID: 50, Status: domain.HomestayPending}}
	svc := NewHomestayService(repo, nil, nil)
	if _, err := svc.View(ctx, nil, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("anonymous: %v", err)
	}
	if _, err := svc.View(ctx, &otherP, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("other user: %v", err)
	}
	for _, who := range []inbound.Principal{ownerP, adminP} {
		if _, err := svc.View(ctx, &who, 1); err != nil {
			t.Fatalf("%+v: %v", who, err)
		}
	}
	repo.current.Status = domain.HomestayActive
	if _, err := svc.View(ctx, nil, 1); err != nil {
		t.Fatalf("published: %v", err)
	}
}

func TestMineAndReviewQueue(t *testing.T) {
	ctx := context.Background()
	repo := &recordingHomestays{}
	svc := NewHomestayService(repo, nil, nil)

	svc.Mine(ctx, ownerP, 0, 0)
	if f := repo.listed; f.HostID != 50 || !f.IncludeInactive || f.Limit != defaultPageSize {
		t.Fatalf("mine: %+v", f)
	}
	if _, err := svc.ForReview(ctx, ownerP, 0, 0, 0); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("queue for non-admin: %v", err)
	}
	svc.ForReview(ctx, adminP, 0, 0, 0)
	if f := repo.listed; f.Status != domain.HomestayPending || f.Sort != "oldest" {
		t.Fatalf("queue defaults to pending, oldest first: %+v", f)
	}
	if _, err := svc.ForReview(ctx, adminP, 9, 0, 0); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("bad status: %v", err)
	}
}

func TestOwnerManagesRatesAndAvailabilityOnlyForOwnHomestay(t *testing.T) {
	ctx := context.Background()
	repo := &recordingHomestays{current: domain.Homestay{ID: 1, HostID: 50, Status: domain.HomestayActive}}
	svc := NewHomestayService(repo, nil, nil)
	rate := domain.Rate{HomestayID: 1, Model: domain.ModelDay, Price: "100", Currency: "VND"}
	if err := svc.SetRate(ctx, otherP, rate); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other user's rate: %v", err)
	}
	if err := svc.DeleteRate(ctx, otherP, 1, domain.ModelDay); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other user's delete: %v", err)
	}
	if err := svc.SetAvailability(ctx, otherP, inbound.SetAvailabilityCommand{HomestayID: 1}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other user's calendar: %v", err)
	}
}
