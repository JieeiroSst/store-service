package mfa

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
)

type fakeProvider struct {
	enabled   bool
	statusErr error
	verifyErr error
	verified  []string
	setupFor  string
}

func (f *fakeProvider) Status(context.Context, int) (domain.MFAStatus, error) {
	at := time.Unix(1_800_000_000, 0)
	if !f.enabled {
		return domain.MFAStatus{}, f.statusErr
	}
	return domain.MFAStatus{Enabled: true, EnabledAt: &at, RecoveryCodesRemaining: 9}, f.statusErr
}

func (f *fakeProvider) SetupTOTP(_ context.Context, _ int, account string) (domain.TOTPSetup, error) {
	f.setupFor = account
	return domain.TOTPSetup{Secret: "S", OtpauthURI: "otpauth://totp/x", QRCodePNG: []byte("png"), Issuer: "Keiki", Account: account, Digits: 6, Period: 30, Algorithm: "SHA1"}, nil
}

func (f *fakeProvider) EnableTOTP(context.Context, int, string) ([]string, error) {
	return []string{"aaaaa-bbbbb"}, nil
}

func (f *fakeProvider) DisableTOTP(context.Context, int, string) error { return nil }

func (f *fakeProvider) RegenerateRecoveryCodes(context.Context, int, string) ([]string, error) {
	return []string{"ccccc-ddddd"}, nil
}

func (f *fakeProvider) VerifyTOTP(_ context.Context, _ int, code string) error {
	f.verified = append(f.verified, code)
	return f.verifyErr
}

type fakeChallenges struct{ m map[string]domain.MFAChallenge }

func (f *fakeChallenges) Save(_ context.Context, tok string, c domain.MFAChallenge, _ time.Duration) error {
	f.m[tok] = c
	return nil
}

func (f *fakeChallenges) Get(_ context.Context, tok string) (domain.MFAChallenge, error) {
	c, ok := f.m[tok]
	if !ok {
		return domain.MFAChallenge{}, domain.ErrMFAChallenge
	}
	return c, nil
}

func (f *fakeChallenges) Delete(_ context.Context, tok string) (bool, error) {
	_, ok := f.m[tok]
	delete(f.m, tok)
	return ok, nil
}

type fakeSessions struct{ started []string }

func (f *fakeSessions) Start(_ context.Context, _ int, _ string, provider string) (dto.LoginResponse, error) {
	f.started = append(f.started, provider)
	return dto.LoginResponse{SessionToken: "access"}, nil
}

type fakeTokens struct{}

func (fakeTokens) GenerateAccessToken(context.Context, int, string, string, []string) (string, error) {
	return "", nil
}
func (fakeTokens) GenerateRefreshToken(context.Context) (string, error) { return "", nil }
func (fakeTokens) ParseAccessToken(_ context.Context, tok string) (domain.AccessClaims, error) {
	if tok != "good" {
		return domain.AccessClaims{}, errors.New("bad")
	}
	return domain.AccessClaims{UserID: 7, Username: "alice"}, nil
}

type fakeStore struct{}

func (fakeStore) SaveSession(context.Context, domain.Session, time.Duration, time.Duration) error {
	return nil
}
func (fakeStore) GetSessionByAccessToken(context.Context, string) (domain.Session, error) {
	return domain.Session{ID: "s", UserID: 7, Username: "alice"}, nil
}
func (fakeStore) GetSessionByRefreshToken(context.Context, string) (domain.Session, error) {
	return domain.Session{}, nil
}
func (fakeStore) DeleteTokens(context.Context, string, string) error { return nil }
func (fakeStore) DeleteSession(context.Context, string) error        { return nil }

type harness struct {
	svc      *Service
	provider *fakeProvider
	ch       *fakeChallenges
	sessions *fakeSessions
}

func newHarness(enabled bool) *harness {
	h := &harness{
		provider: &fakeProvider{enabled: enabled},
		ch:       &fakeChallenges{m: map[string]domain.MFAChallenge{}},
		sessions: &fakeSessions{},
	}
	h.svc = New(h.provider, h.ch, h.sessions, fakeTokens{}, fakeStore{}, Config{ChallengeTTL: 5 * time.Minute})
	return h
}

func mfaToken(t *testing.T, err error) string {
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.CodeMFARequired {
		t.Fatalf("want mfa required, got %v", err)
	}
	if de.Metadata["expires_in"] != "300" {
		t.Fatalf("metadata = %v", de.Metadata)
	}
	return de.Metadata["mfa_token"]
}

func TestLoginWithoutMFAStartsSession(t *testing.T) {
	h := newHarness(false)
	resp, err := h.svc.Start(context.Background(), 7, "alice", domain.ProviderPassword)
	if err != nil || resp.SessionToken != "access" || len(h.sessions.started) != 1 {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
}

func TestLoginFailsClosedWhenAuthorizeIsDown(t *testing.T) {
	h := newHarness(false)
	h.provider.statusErr = domain.ErrServiceUnavailable
	if _, err := h.svc.Start(context.Background(), 7, "alice", domain.ProviderPassword); !errors.Is(err, domain.ErrServiceUnavailable) {
		t.Fatalf("err=%v", err)
	}
	if len(h.sessions.started) != 0 {
		t.Fatal("session must not start when the 2FA state is unknown")
	}
}

func TestLoginWithMFA(t *testing.T) {
	h := newHarness(true)
	ctx := context.Background()
	_, err := h.svc.Start(ctx, 7, "alice", domain.ProviderGoogle)
	token := mfaToken(t, err)
	if len(h.sessions.started) != 0 {
		t.Fatal("session started before second factor")
	}

	h.provider.verifyErr = domain.ErrMFAInvalidCode
	if _, err := h.svc.VerifyLogin(ctx, token, "000000"); !errors.Is(err, domain.ErrMFAInvalidCode) {
		t.Fatalf("wrong code err=%v", err)
	}
	h.provider.verifyErr = nil
	resp, err := h.svc.VerifyLogin(ctx, token, " 123456 ")
	if err != nil || resp.SessionToken != "access" || h.sessions.started[0] != domain.ProviderGoogle {
		t.Fatalf("verify resp=%+v err=%v started=%v", resp, err, h.sessions.started)
	}
	if h.provider.verified[1] != "123456" {
		t.Fatalf("code not trimmed: %q", h.provider.verified[1])
	}
	if _, err := h.svc.VerifyLogin(ctx, token, "123456"); !errors.Is(err, domain.ErrMFAChallenge) {
		t.Fatalf("challenge reuse err=%v", err)
	}
}

func TestVerifyDropsChallengeWhenMFAWasDisabled(t *testing.T) {
	h := newHarness(true)
	ctx := context.Background()
	_, err := h.svc.Start(ctx, 7, "alice", domain.ProviderPassword)
	token := mfaToken(t, err)
	h.provider.verifyErr = domain.ErrMFANotEnabled
	if _, err := h.svc.VerifyLogin(ctx, token, "123456"); !errors.Is(err, domain.ErrMFAChallenge) {
		t.Fatalf("err=%v", err)
	}
	if len(h.ch.m) != 0 {
		t.Fatal("challenge must be removed")
	}
}

func TestManagementUsesSessionUser(t *testing.T) {
	h := newHarness(true)
	ctx := context.Background()
	if _, err := h.svc.SetupTOTP(ctx, "bad"); !errors.Is(err, domain.ErrFailedToken) {
		t.Fatalf("bad token err=%v", err)
	}
	setup, err := h.svc.SetupTOTP(ctx, "good")
	if err != nil || setup.QRCode != "data:image/png;base64,cG5n" || setup.Account != "alice" || h.provider.setupFor != "alice" {
		t.Fatalf("setup=%+v err=%v", setup, err)
	}
	st, err := h.svc.Status(ctx, "good")
	if err != nil || !st.Enabled || st.Method != domain.MFAMethodTOTP || st.RecoveryCodesRemaining != 9 {
		t.Fatalf("status=%+v err=%v", st, err)
	}
	codes, err := h.svc.EnableTOTP(ctx, "good", "123456")
	if err != nil || len(codes.RecoveryCodes) != 1 {
		t.Fatalf("enable=%+v err=%v", codes, err)
	}
	if err := h.svc.DisableTOTP(ctx, "good", "aaaaa-bbbbb"); err != nil {
		t.Fatalf("disable err=%v", err)
	}
}
