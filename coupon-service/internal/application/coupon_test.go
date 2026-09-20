package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
)

type fakeCouponRepo struct {
	byID map[int64]*model.Coupon
}

func newFakeCouponRepo(coupons ...*model.Coupon) *fakeCouponRepo {
	r := &fakeCouponRepo{byID: map[int64]*model.Coupon{}}
	for _, c := range coupons {
		cp := *c
		r.byID[cp.ID] = &cp
	}
	return r
}

func (r *fakeCouponRepo) Create(_ context.Context, c *model.Coupon) (*model.Coupon, error) {
	r.byID[c.ID] = c
	return c, nil
}
func (r *fakeCouponRepo) GetByID(_ context.Context, id int64) (*model.Coupon, error) {
	c, ok := r.byID[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	cp := *c
	return &cp, nil
}
func (r *fakeCouponRepo) GetByCode(_ context.Context, code string) (*model.Coupon, error) {
	for _, c := range r.byID {
		if c.Code == code {
			cp := *c
			return &cp, nil
		}
	}
	return nil, port.ErrNotFound
}
func (r *fakeCouponRepo) List(_ context.Context) ([]model.Coupon, error) {
	var out []model.Coupon
	for _, c := range r.byID {
		out = append(out, *c)
	}
	return out, nil
}
func (r *fakeCouponRepo) Update(_ context.Context, c *model.Coupon) (*model.Coupon, error) {
	r.byID[c.ID] = c
	return c, nil
}
func (r *fakeCouponRepo) Delete(_ context.Context, id int64) error {
	delete(r.byID, id)
	return nil
}
func (r *fakeCouponRepo) IncrementUsage(_ context.Context, id int64) error {
	c, ok := r.byID[id]
	if !ok {
		return port.ErrNotFound
	}
	if !c.HasUsesLeft() {
		return port.ErrUsageLimitReached
	}
	c.CurrentUses++
	return nil
}
func (r *fakeCouponRepo) DeactivateStale(_ context.Context, now time.Time) (int64, error) {
	var n int64
	for _, c := range r.byID {
		if !c.IsActive {
			continue
		}
		exhausted := c.MaxUses != nil && c.CurrentUses >= *c.MaxUses
		if now.After(c.EndDate) || exhausted {
			c.IsActive = false
			n++
		}
	}
	return n, nil
}

type fakeRestrictionRepo struct {
	restrictions []model.CouponRestriction
}

func (r *fakeRestrictionRepo) Create(_ context.Context, res *model.CouponRestriction) (*model.CouponRestriction, error) {
	r.restrictions = append(r.restrictions, *res)
	return res, nil
}
func (r *fakeRestrictionRepo) GetByID(_ context.Context, id int64) (*model.CouponRestriction, error) {
	for _, res := range r.restrictions {
		if res.ID == id {
			return &res, nil
		}
	}
	return nil, port.ErrNotFound
}
func (r *fakeRestrictionRepo) ListByCoupon(_ context.Context, couponID int64) ([]model.CouponRestriction, error) {
	var out []model.CouponRestriction
	for _, res := range r.restrictions {
		if res.CouponID == couponID {
			out = append(out, res)
		}
	}
	return out, nil
}
func (r *fakeRestrictionRepo) Update(_ context.Context, res *model.CouponRestriction) (*model.CouponRestriction, error) {
	return res, nil
}
func (r *fakeRestrictionRepo) Delete(_ context.Context, id int64) error { return nil }

type fakeUsageRepo struct {
	usages []model.CouponUsage
}

func (r *fakeUsageRepo) Create(_ context.Context, u *model.CouponUsage) error {
	r.usages = append(r.usages, *u)
	return nil
}
func (r *fakeUsageRepo) GetByOrderAndCoupon(_ context.Context, orderID, couponID int64) (*model.CouponUsage, error) {
	for _, u := range r.usages {
		if u.OrderID == orderID && u.CouponID == couponID {
			return &u, nil
		}
	}
	return nil, port.ErrNotFound
}
func (r *fakeUsageRepo) ListByCoupon(_ context.Context, couponID int64) ([]model.CouponUsage, error) {
	return r.usages, nil
}
func (r *fakeUsageRepo) ListByUser(_ context.Context, userID int64) ([]model.CouponUsage, error) {
	return r.usages, nil
}

type fakeUserCouponRepo struct {
	byID map[int64]*model.UserCoupon
}

func newFakeUserCouponRepo(rows ...*model.UserCoupon) *fakeUserCouponRepo {
	r := &fakeUserCouponRepo{byID: map[int64]*model.UserCoupon{}}
	for i, row := range rows {
		cp := *row
		cp.ID = int64(i + 1)
		r.byID[cp.ID] = &cp
	}
	return r
}

func (r *fakeUserCouponRepo) Create(_ context.Context, uc *model.UserCoupon) (*model.UserCoupon, error) {
	uc.ID = int64(len(r.byID) + 1)
	r.byID[uc.ID] = uc
	return uc, nil
}
func (r *fakeUserCouponRepo) GetByID(_ context.Context, id int64) (*model.UserCoupon, error) {
	uc, ok := r.byID[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	return uc, nil
}
func (r *fakeUserCouponRepo) GetByUserAndCoupon(_ context.Context, userID, couponID int64) (*model.UserCoupon, error) {
	for _, uc := range r.byID {
		if uc.UserID == userID && uc.CouponID == couponID {
			return uc, nil
		}
	}
	return nil, port.ErrNotFound
}
func (r *fakeUserCouponRepo) Update(_ context.Context, uc *model.UserCoupon) (*model.UserCoupon, error) {
	r.byID[uc.ID] = uc
	return uc, nil
}
func (r *fakeUserCouponRepo) Delete(_ context.Context, id int64) error {
	delete(r.byID, id)
	return nil
}
func (r *fakeUserCouponRepo) ListByUser(_ context.Context, userID int64) ([]model.UserCoupon, error) {
	var out []model.UserCoupon
	for _, uc := range r.byID {
		if uc.UserID == userID {
			out = append(out, *uc)
		}
	}
	return out, nil
}
func (r *fakeUserCouponRepo) ListByCoupon(_ context.Context, couponID int64) ([]model.UserCoupon, error) {
	var out []model.UserCoupon
	for _, uc := range r.byID {
		if uc.CouponID == couponID {
			out = append(out, *uc)
		}
	}
	return out, nil
}

func activeCoupon() *model.Coupon {
	return &model.Coupon{
		ID:              1,
		Code:            "SAVE10",
		Type:            model.CouponPercentage,
		DiscountValue:   10,
		MinimumPurchase: 50000,
		IsActive:        true,
		StartDate:       time.Now().Add(-24 * time.Hour),
		EndDate:         time.Now().Add(24 * time.Hour),
	}
}

func TestValidateCoupon(t *testing.T) {
	t.Run("happy path applies discount", func(t *testing.T) {
		svc := NewCouponService(newFakeCouponRepo(activeCoupon()), &fakeRestrictionRepo{}, &fakeUsageRepo{}, newFakeUserCouponRepo())

		coupon, discount, err := svc.ValidateCoupon(context.Background(), port.ValidateCouponInput{
			Code: "SAVE10", UserID: 1, PurchaseAmount: 100000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if discount != 10000 {
			t.Errorf("discount = %v, want 10000", discount)
		}
		if coupon.Code != "SAVE10" {
			t.Errorf("coupon code = %v, want SAVE10", coupon.Code)
		}
	})

	t.Run("below minimum purchase", func(t *testing.T) {
		svc := NewCouponService(newFakeCouponRepo(activeCoupon()), &fakeRestrictionRepo{}, &fakeUsageRepo{}, newFakeUserCouponRepo())

		_, _, err := svc.ValidateCoupon(context.Background(), port.ValidateCouponInput{
			Code: "SAVE10", UserID: 1, PurchaseAmount: 1000,
		})
		if !errors.Is(err, port.ErrMinimumPurchaseNotMet) {
			t.Fatalf("err = %v, want ErrMinimumPurchaseNotMet", err)
		}
	})

	t.Run("inactive coupon", func(t *testing.T) {
		coupon := activeCoupon()
		coupon.IsActive = false
		svc := NewCouponService(newFakeCouponRepo(coupon), &fakeRestrictionRepo{}, &fakeUsageRepo{}, newFakeUserCouponRepo())

		_, _, err := svc.ValidateCoupon(context.Background(), port.ValidateCouponInput{
			Code: "SAVE10", UserID: 1, PurchaseAmount: 100000,
		})
		if !errors.Is(err, port.ErrCouponInactive) {
			t.Fatalf("err = %v, want ErrCouponInactive", err)
		}
	})

	t.Run("expired coupon", func(t *testing.T) {
		coupon := activeCoupon()
		coupon.EndDate = time.Now().Add(-time.Hour)
		svc := NewCouponService(newFakeCouponRepo(coupon), &fakeRestrictionRepo{}, &fakeUsageRepo{}, newFakeUserCouponRepo())

		_, _, err := svc.ValidateCoupon(context.Background(), port.ValidateCouponInput{
			Code: "SAVE10", UserID: 1, PurchaseAmount: 100000,
		})
		if !errors.Is(err, port.ErrCouponExpired) {
			t.Fatalf("err = %v, want ErrCouponExpired", err)
		}
	})

	t.Run("usage limit reached", func(t *testing.T) {
		coupon := activeCoupon()
		max := 1
		coupon.MaxUses = &max
		coupon.CurrentUses = 1
		svc := NewCouponService(newFakeCouponRepo(coupon), &fakeRestrictionRepo{}, &fakeUsageRepo{}, newFakeUserCouponRepo())

		_, _, err := svc.ValidateCoupon(context.Background(), port.ValidateCouponInput{
			Code: "SAVE10", UserID: 1, PurchaseAmount: 100000,
		})
		if !errors.Is(err, port.ErrUsageLimitReached) {
			t.Fatalf("err = %v, want ErrUsageLimitReached", err)
		}
	})

	t.Run("category restricted", func(t *testing.T) {
		coupon := activeCoupon()
		restrictions := &fakeRestrictionRepo{restrictions: []model.CouponRestriction{
			{CouponID: coupon.ID, RestrictionType: model.RestrictionCategory, RestrictedEntityID: 99, IsExclude: false},
		}}
		svc := NewCouponService(newFakeCouponRepo(coupon), restrictions, &fakeUsageRepo{}, newFakeUserCouponRepo())

		_, _, err := svc.ValidateCoupon(context.Background(), port.ValidateCouponInput{
			Code: "SAVE10", UserID: 1, PurchaseAmount: 100000, CategoryIDs: []int64{5},
		})
		if !errors.Is(err, port.ErrCouponRestricted) {
			t.Fatalf("err = %v, want ErrCouponRestricted", err)
		}

		_, _, err = svc.ValidateCoupon(context.Background(), port.ValidateCouponInput{
			Code: "SAVE10", UserID: 1, PurchaseAmount: 100000, CategoryIDs: []int64{99},
		})
		if err != nil {
			t.Fatalf("unexpected error for matching category: %v", err)
		}
	})

	t.Run("targeted coupon requires assignment", func(t *testing.T) {
		coupon := activeCoupon()
		userCoupons := newFakeUserCouponRepo(&model.UserCoupon{UserID: 2, CouponID: coupon.ID})
		svc := NewCouponService(newFakeCouponRepo(coupon), &fakeRestrictionRepo{}, &fakeUsageRepo{}, userCoupons)

		_, _, err := svc.ValidateCoupon(context.Background(), port.ValidateCouponInput{
			Code: "SAVE10", UserID: 1, PurchaseAmount: 100000,
		})
		if !errors.Is(err, port.ErrCouponNotAssignedToUser) {
			t.Fatalf("err = %v, want ErrCouponNotAssignedToUser", err)
		}

		_, _, err = svc.ValidateCoupon(context.Background(), port.ValidateCouponInput{
			Code: "SAVE10", UserID: 2, PurchaseAmount: 100000,
		})
		if err != nil {
			t.Fatalf("unexpected error for assigned user: %v", err)
		}
	})
}

func TestApplyCoupon(t *testing.T) {
	t.Run("records usage and increments count", func(t *testing.T) {
		coupon := activeCoupon()
		couponRepo := newFakeCouponRepo(coupon)
		usageRepo := &fakeUsageRepo{}
		svc := NewCouponService(couponRepo, &fakeRestrictionRepo{}, usageRepo, newFakeUserCouponRepo())

		usage, err := svc.ApplyCoupon(context.Background(), port.ApplyCouponInput{
			ValidateCouponInput: port.ValidateCouponInput{Code: "SAVE10", UserID: 1, PurchaseAmount: 100000},
			OrderID:             42,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if usage.DiscountAmount != 10000 {
			t.Errorf("discount = %v, want 10000", usage.DiscountAmount)
		}

		updated, _ := couponRepo.GetByID(context.Background(), coupon.ID)
		if updated.CurrentUses != 1 {
			t.Errorf("current_uses = %v, want 1", updated.CurrentUses)
		}
	})

	t.Run("rejects double redemption on the same order", func(t *testing.T) {
		coupon := activeCoupon()
		svc := NewCouponService(newFakeCouponRepo(coupon), &fakeRestrictionRepo{}, &fakeUsageRepo{}, newFakeUserCouponRepo())

		in := port.ApplyCouponInput{
			ValidateCouponInput: port.ValidateCouponInput{Code: "SAVE10", UserID: 1, PurchaseAmount: 100000},
			OrderID:             42,
		}
		if _, err := svc.ApplyCoupon(context.Background(), in); err != nil {
			t.Fatalf("first apply failed: %v", err)
		}
		if _, err := svc.ApplyCoupon(context.Background(), in); !errors.Is(err, port.ErrOrderAlreadyUsedCoupon) {
			t.Fatalf("err = %v, want ErrOrderAlreadyUsedCoupon", err)
		}
	})

	t.Run("marks targeted user-coupon as used", func(t *testing.T) {
		coupon := activeCoupon()
		userCoupons := newFakeUserCouponRepo(&model.UserCoupon{UserID: 1, CouponID: coupon.ID})
		svc := NewCouponService(newFakeCouponRepo(coupon), &fakeRestrictionRepo{}, &fakeUsageRepo{}, userCoupons)

		if _, err := svc.ApplyCoupon(context.Background(), port.ApplyCouponInput{
			ValidateCouponInput: port.ValidateCouponInput{Code: "SAVE10", UserID: 1, PurchaseAmount: 100000},
			OrderID:             1,
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		uc, err := userCoupons.GetByUserAndCoupon(context.Background(), 1, coupon.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !uc.IsUsed {
			t.Error("expected user coupon to be marked used")
		}
	})
}

func TestDeactivateStaleCoupons(t *testing.T) {
	stillValid := activeCoupon()
	stillValid.ID = 1
	stillValid.Code = "STILLVALID"

	expired := activeCoupon()
	expired.ID = 2
	expired.Code = "EXPIRED"
	expired.EndDate = time.Now().Add(-time.Hour)

	maxUses := 1
	exhausted := activeCoupon()
	exhausted.ID = 3
	exhausted.Code = "EXHAUSTED"
	exhausted.MaxUses = &maxUses
	exhausted.CurrentUses = 1

	couponRepo := newFakeCouponRepo(stillValid, expired, exhausted)
	svc := NewCouponService(couponRepo, &fakeRestrictionRepo{}, &fakeUsageRepo{}, newFakeUserCouponRepo())

	n, err := svc.DeactivateStaleCoupons(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("deactivated count = %d, want 2", n)
	}

	for id, wantActive := range map[int64]bool{1: true, 2: false, 3: false} {
		c, err := couponRepo.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.IsActive != wantActive {
			t.Errorf("coupon %d IsActive = %v, want %v", id, c.IsActive, wantActive)
		}
	}
}
