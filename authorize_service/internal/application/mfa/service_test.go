package mfa

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
)

type fakeRepo struct {
	recs map[string]model.UserTOTP
}

func (f *fakeRepo) Get(_ context.Context, id string) (model.UserTOTP, error) {
	r, ok := f.recs[id]
	if !ok {
		return model.UserTOTP{}, common.ErrMFANotSetUp
	}
	return r, nil
}

func (f *fakeRepo) SavePending(_ context.Context, id, secret string) error {
	if r, ok := f.recs[id]; ok && r.Enabled {
		return common.ErrMFAAlreadyEnabled
	}
	f.recs[id] = model.UserTOTP{UserID: id, Secret: secret}
	return nil
}

func (f *fakeRepo) Enable(_ context.Context, id string, step int64, codes string, at time.Time) error {
	r := f.recs[id]
	if r.Enabled || r.LastUsedStep >= step {
		return common.ErrMFAInvalidCode
	}
	r.Enabled, r.LastUsedStep, r.RecoveryCodes, r.EnabledAt = true, step, codes, &at
	f.recs[id] = r
	return nil
}

func (f *fakeRepo) MarkUsed(_ context.Context, id string, step int64) (bool, error) {
	r := f.recs[id]
	if r.LastUsedStep >= step {
		return false, nil
	}
	r.LastUsedStep = step
	f.recs[id] = r
	return true, nil
}

func (f *fakeRepo) SwapRecoveryCodes(_ context.Context, id, old, new string) (bool, error) {
	r := f.recs[id]
	if r.RecoveryCodes != old {
		return false, nil
	}
	r.RecoveryCodes = new
	f.recs[id] = r
	return true, nil
}

func (f *fakeRepo) Delete(_ context.Context, id string) error {
	delete(f.recs, id)
	return nil
}

type plainBox struct{}

func (plainBox) Seal(p []byte) (string, error) { return "sealed:" + string(p), nil }
func (plainBox) Open(s string) ([]byte, error) { return []byte(s[len("sealed:"):]), nil }

type fakeQR struct{}

func (fakeQR) PNG(string, int) ([]byte, error) { return []byte("png"), nil }

type fakeThrottle struct{ failures map[string]int }

func (f *fakeThrottle) Blocked(_ context.Context, k string) (bool, error) { return f.failures[k] >= 5, nil }
func (f *fakeThrottle) RecordFailure(_ context.Context, k string) error  { f.failures[k]++; return nil }
func (f *fakeThrottle) Reset(_ context.Context, k string) error          { delete(f.failures, k); return nil }

type harness struct {
	svc   *Service
	repo  *fakeRepo
	clock time.Time
}

func newHarness() *harness {
	h := &harness{repo: &fakeRepo{recs: map[string]model.UserTOTP{}}, clock: time.Unix(1_800_000_000, 0)}
	h.svc = New(h.repo, plainBox{}, fakeQR{}, &fakeThrottle{failures: map[string]int{}}, Config{Issuer: "Keiki"})
	h.svc.now = func() time.Time { return h.clock }
	return h
}

func (h *harness) code(t *testing.T, secret string) string {
	key, err := decodeSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	return hotp(key, timeStep(h.clock))
}

func (h *harness) enroll(t *testing.T) (string, []string) {
	ctx := context.Background()
	setup, err := h.svc.SetupTOTP(ctx, "7", "alice")
	if err != nil {
		t.Fatal(err)
	}
	codes, err := h.svc.EnableTOTP(ctx, "7", h.code(t, setup.Secret))
	if err != nil {
		t.Fatal(err)
	}
	h.clock = h.clock.Add(totpPeriod * time.Second)
	return setup.Secret, codes
}

func TestSetupReturnsAuthenticatorPayload(t *testing.T) {
	h := newHarness()
	setup, err := h.svc.SetupTOTP(context.Background(), "7", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if setup.Account != "alice" || setup.Issuer != "Keiki" || setup.Digits != 6 || setup.Period != 30 || string(setup.QRCodePNG) != "png" {
		t.Fatalf("setup = %+v", setup)
	}
	if st, _ := h.svc.Status(context.Background(), "7"); st.Enabled {
		t.Fatal("pending secret must not count as enabled")
	}
}

func TestRequiresUserID(t *testing.T) {
	h := newHarness()
	if _, err := h.svc.SetupTOTP(context.Background(), "", "alice"); !errors.Is(err, common.ErrInvalidUserID) {
		t.Fatalf("err = %v", err)
	}
	if _, err := h.svc.VerifyTOTP(context.Background(), "", "123456", true); !errors.Is(err, common.ErrInvalidUserID) {
		t.Fatalf("err = %v", err)
	}
}

func TestEnableRejectsWrongCode(t *testing.T) {
	h := newHarness()
	ctx := context.Background()
	if _, err := h.svc.SetupTOTP(ctx, "7", "alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := h.svc.EnableTOTP(ctx, "7", "000000"); !errors.Is(err, common.ErrMFAInvalidCode) {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyFlow(t *testing.T) {
	h := newHarness()
	ctx := context.Background()
	secret, codes := h.enroll(t)
	if len(codes) != recoveryCodeCount {
		t.Fatalf("codes = %v", codes)
	}
	if _, err := h.svc.SetupTOTP(ctx, "7", "alice"); !errors.Is(err, common.ErrMFAAlreadyEnabled) {
		t.Fatalf("re-setup err = %v", err)
	}

	code := h.code(t, secret)
	if v, err := h.svc.VerifyTOTP(ctx, "7", code, true); err != nil || v.UsedRecoveryCode {
		t.Fatalf("verify v=%+v err=%v", v, err)
	}
	if _, err := h.svc.VerifyTOTP(ctx, "7", code, true); !errors.Is(err, common.ErrMFAInvalidCode) {
		t.Fatalf("replay err = %v", err)
	}
	if _, err := h.svc.VerifyTOTP(ctx, "7", codes[0], false); !errors.Is(err, common.ErrMFAInvalidCode) {
		t.Fatalf("recovery code must be refused when not allowed, err = %v", err)
	}
	if v, err := h.svc.VerifyTOTP(ctx, "7", codes[0], true); err != nil || !v.UsedRecoveryCode {
		t.Fatalf("recovery v=%+v err=%v", v, err)
	}
	st, _ := h.svc.Status(ctx, "7")
	if !st.Enabled || st.RecoveryCodesRemaining != recoveryCodeCount-1 || st.EnabledAt == nil {
		t.Fatalf("status = %+v", st)
	}
}

func TestVerifyThrottles(t *testing.T) {
	h := newHarness()
	ctx := context.Background()
	secret, _ := h.enroll(t)
	for i := 0; i < 5; i++ {
		_, _ = h.svc.VerifyTOTP(ctx, "7", "000000", true)
	}
	if _, err := h.svc.VerifyTOTP(ctx, "7", h.code(t, secret), true); !errors.Is(err, common.ErrMFATooManyAttempts) {
		t.Fatalf("err = %v", err)
	}
}

func TestDisableAndRegenerate(t *testing.T) {
	h := newHarness()
	ctx := context.Background()
	secret, old := h.enroll(t)

	fresh, err := h.svc.RegenerateRecoveryCodes(ctx, "7", h.code(t, secret))
	if err != nil || len(fresh) != recoveryCodeCount {
		t.Fatalf("regenerate err = %v", err)
	}
	if err := h.svc.DisableTOTP(ctx, "7", old[0]); !errors.Is(err, common.ErrMFAInvalidCode) {
		t.Fatalf("old recovery code err = %v", err)
	}
	if err := h.svc.DisableTOTP(ctx, "7", fresh[0]); err != nil {
		t.Fatalf("disable err = %v", err)
	}
	if _, err := h.svc.VerifyTOTP(ctx, "7", "123456", true); !errors.Is(err, common.ErrMFANotEnabled) {
		t.Fatalf("after disable err = %v", err)
	}
}
