package session

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
)

type memRepo struct {
	rows map[string]domain.LoginSession
}

func (m *memRepo) Create(_ context.Context, s domain.LoginSession) error {
	m.rows[s.ID] = s
	return nil
}

func (m *memRepo) Get(_ context.Context, id string) (domain.LoginSession, error) {
	s, ok := m.rows[id]
	if !ok {
		return domain.LoginSession{}, domain.ErrSessionNotFound
	}
	return s, nil
}

func (m *memRepo) Touch(_ context.Context, id string, last, exp time.Time) error {
	s := m.rows[id]
	s.LastActiveAt, s.ExpiresAt = last, exp
	m.rows[id] = s
	return nil
}

func (m *memRepo) ListActive(_ context.Context, userID int, now time.Time) ([]domain.LoginSession, error) {
	var out []domain.LoginSession
	for _, s := range m.rows {
		if s.UserID == userID && s.RevokedAt == nil && s.ExpiresAt.After(now) {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastActiveAt.After(out[j].LastActiveAt) })
	return out, nil
}

func (m *memRepo) Revoke(_ context.Context, ids []string, reason string, at time.Time) error {
	for _, id := range ids {
		s := m.rows[id]
		if s.RevokedAt == nil {
			s.RevokedAt, s.RevokeReason = &at, reason
			m.rows[id] = s
		}
	}
	return nil
}

func (m *memRepo) rowsByProvider(p string) string {
	for id, r := range m.rows {
		if r.Provider == p {
			return id
		}
	}
	return ""
}

type memStore struct {
	access, refresh, byID map[string]domain.Session
}

func newMemStore() *memStore {
	return &memStore{access: map[string]domain.Session{}, refresh: map[string]domain.Session{}, byID: map[string]domain.Session{}}
}

func (m *memStore) SaveSession(_ context.Context, s domain.Session, _, _ time.Duration) error {
	m.access[s.AccessToken], m.refresh[s.RefreshToken] = s, s
	if s.ID != "" {
		m.byID[s.ID] = s
	}
	return nil
}

func (m *memStore) GetSessionByAccessToken(_ context.Context, t string) (domain.Session, error) {
	s, ok := m.access[t]
	if !ok {
		return s, errors.New("miss")
	}
	return s, nil
}

func (m *memStore) GetSessionByRefreshToken(_ context.Context, t string) (domain.Session, error) {
	s, ok := m.refresh[t]
	if !ok {
		return s, errors.New("miss")
	}
	return s, nil
}

func (m *memStore) DeleteTokens(_ context.Context, a, r string) error {
	delete(m.access, a)
	delete(m.refresh, r)
	return nil
}

func (m *memStore) DeleteSession(ctx context.Context, id string) error {
	if s, ok := m.byID[id]; ok {
		_ = m.DeleteTokens(ctx, s.AccessToken, s.RefreshToken)
	}
	delete(m.byID, id)
	return nil
}

type seqTokens struct{ n int }

func (g *seqTokens) GenerateAccessToken(_ context.Context, userID int, _, _ string, _ []string) (string, error) {
	g.n++
	return "at-" + strconv.Itoa(userID) + "-" + strconv.Itoa(g.n), nil
}

func (g *seqTokens) GenerateRefreshToken(context.Context) (string, error) {
	g.n++
	return "rt-" + strconv.Itoa(g.n), nil
}

func (g *seqTokens) ParseAccessToken(_ context.Context, t string) (domain.AccessClaims, error) {
	return domain.AccessClaims{}, nil
}

type noRoles struct{ output.Authorizer }

func (noRoles) GetUserRoles(context.Context, int) (domain.UserRoles, error) {
	return domain.UserRoles{PrimaryRole: "user"}, nil
}

type harness struct {
	svc   *Service
	repo  *memRepo
	store *memStore
	clock time.Time
}

func newHarness(max int) *harness {
	h := &harness{repo: &memRepo{rows: map[string]domain.LoginSession{}}, store: newMemStore(), clock: time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)}
	h.svc = New(h.repo, h.store, &seqTokens{}, noRoles{}, Config{AccessTTL: 15 * time.Minute, RefreshTTL: 24 * time.Hour, MaxPerUser: max})
	h.svc.now = func() time.Time { return h.clock }
	return h
}

func (h *harness) tick() { h.clock = h.clock.Add(time.Minute) }

func TestStartRecordsSessionWithClientInfo(t *testing.T) {
	h := newHarness(0)
	ctx := domain.WithClientInfo(context.Background(), domain.ClientInfo{IP: "1.2.3.4", UserAgent: "Firefox"})
	resp, err := h.svc.Start(ctx, 7, "quan", domain.ProviderGoogle)
	if err != nil {
		t.Fatal(err)
	}
	cur := h.store.access[resp.SessionToken]
	row := h.repo.rows[cur.ID]
	if cur.ID == "" || row.UserID != 7 || row.Provider != domain.ProviderGoogle || row.IP != "1.2.3.4" || row.UserAgent != "Firefox" || !row.ExpiresAt.Equal(h.clock.Add(24*time.Hour)) {
		t.Fatalf("%+v %+v", cur, row)
	}
}

func TestRefreshKeepsSessionIDAndSlidesExpiry(t *testing.T) {
	h := newHarness(0)
	login, _ := h.svc.Start(context.Background(), 7, "quan", domain.ProviderPassword)
	id := h.store.access[login.SessionToken].ID
	h.tick()

	ref, err := h.svc.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if h.store.access[ref.NewSessionToken].ID != id {
		t.Fatal("refresh changed session id")
	}
	if _, ok := h.store.access[login.SessionToken]; ok {
		t.Fatal("old access token still valid")
	}
	if _, err := h.svc.Refresh(context.Background(), login.RefreshToken); !errors.Is(err, domain.ErrRefreshTokenInvalid) {
		t.Fatalf("reused refresh token: %v", err)
	}
	if row := h.repo.rows[id]; !row.LastActiveAt.Equal(h.clock) || !row.ExpiresAt.Equal(h.clock.Add(24*time.Hour)) {
		t.Fatalf("%+v", row)
	}
}

func TestRefreshRejectsRevokedSession(t *testing.T) {
	h := newHarness(0)
	login, _ := h.svc.Start(context.Background(), 7, "quan", domain.ProviderPassword)
	id := h.store.access[login.SessionToken].ID
	_ = h.repo.Revoke(context.Background(), []string{id}, domain.RevokeByUser, h.clock)

	if _, err := h.svc.Refresh(context.Background(), login.RefreshToken); !errors.Is(err, domain.ErrRefreshTokenInvalid) {
		t.Fatalf("%v", err)
	}
	if _, ok := h.store.byID[id]; ok {
		t.Fatal("tokens of revoked session were not purged")
	}
}

func TestRefreshMigratesLegacySession(t *testing.T) {
	h := newHarness(0)
	legacy := domain.Session{UserID: 7, Username: "quan", AccessToken: "old-at", RefreshToken: "old-rt"}
	_ = h.store.SaveSession(context.Background(), legacy, 0, 0)

	ref, err := h.svc.Refresh(context.Background(), "old-rt")
	if err != nil {
		t.Fatal(err)
	}
	if id := h.store.access[ref.NewSessionToken].ID; id == "" || h.repo.rows[id].UserID != 7 {
		t.Fatalf("legacy session not tracked: %q", id)
	}
	if _, ok := h.store.refresh["old-rt"]; ok {
		t.Fatal("legacy refresh token still valid")
	}
}

func TestListRevokeAndLogout(t *testing.T) {
	h := newHarness(0)
	ctx := context.Background()
	a, _ := h.svc.Start(ctx, 7, "quan", domain.ProviderPassword)
	h.tick()
	b, _ := h.svc.Start(ctx, 7, "quan", domain.ProviderFacebook)
	h.tick()
	other, _ := h.svc.Start(ctx, 8, "mai", domain.ProviderPassword)

	list, err := h.svc.List(ctx, b.SessionToken)
	if err != nil || len(list.Sessions) != 2 || !list.Sessions[0].Current || list.Sessions[0].Provider != domain.ProviderFacebook || list.Sessions[1].Current {
		t.Fatalf("%v %+v", err, list)
	}

	otherID := h.store.access[other.SessionToken].ID
	if err := h.svc.Revoke(ctx, b.SessionToken, otherID); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("revoked another user's session: %v", err)
	}

	aID := h.store.access[a.SessionToken].ID
	if err := h.svc.Revoke(ctx, b.SessionToken, aID); err != nil {
		t.Fatal(err)
	}
	if _, ok := h.store.access[a.SessionToken]; ok || h.repo.rows[aID].RevokeReason != domain.RevokeByUser {
		t.Fatalf("session a not revoked: %+v", h.repo.rows[aID])
	}

	if err := h.svc.End(ctx, b.SessionToken); err != nil {
		t.Fatal(err)
	}
	if _, err := h.svc.List(ctx, b.SessionToken); !errors.Is(err, domain.ErrFailedToken) {
		t.Fatalf("logged-out token still works: %v", err)
	}
	if h.repo.rows[h.repo.rowsByProvider(domain.ProviderFacebook)].RevokeReason != domain.RevokeLogout {
		t.Fatal("logout did not record revoke reason")
	}
}

func TestRevokeOthersKeepsCurrent(t *testing.T) {
	h := newHarness(0)
	ctx := context.Background()
	_, _ = h.svc.Start(ctx, 7, "quan", domain.ProviderPassword)
	_, _ = h.svc.Start(ctx, 7, "quan", domain.ProviderGoogle)
	cur, _ := h.svc.Start(ctx, 7, "quan", domain.ProviderTwitter)

	resp, err := h.svc.RevokeOthers(ctx, cur.SessionToken, false)
	if err != nil || resp.Revoked != 2 {
		t.Fatalf("%v %+v", err, resp)
	}
	list, _ := h.svc.List(ctx, cur.SessionToken)
	if len(list.Sessions) != 1 || !list.Sessions[0].Current {
		t.Fatalf("%+v", list)
	}

	if resp, err := h.svc.RevokeOthers(ctx, cur.SessionToken, true); err != nil || resp.Revoked != 1 {
		t.Fatalf("%v %+v", err, resp)
	}
	if _, ok := h.store.access[cur.SessionToken]; ok {
		t.Fatal("current session survived include_current")
	}
}

func TestMaxPerUserEvictsLeastRecentlyActive(t *testing.T) {
	h := newHarness(2)
	ctx := context.Background()
	first, _ := h.svc.Start(ctx, 7, "quan", domain.ProviderPassword)
	h.tick()
	second, _ := h.svc.Start(ctx, 7, "quan", domain.ProviderPassword)
	h.tick()
	third, _ := h.svc.Start(ctx, 7, "quan", domain.ProviderPassword)

	if _, ok := h.store.access[first.SessionToken]; ok {
		t.Fatal("oldest session not evicted")
	}
	for _, tok := range []string{second.SessionToken, third.SessionToken} {
		if _, ok := h.store.access[tok]; !ok {
			t.Fatal("recent session evicted")
		}
	}
	evicted := 0
	for _, r := range h.repo.rows {
		if r.RevokeReason == domain.RevokeEvicted {
			evicted++
		}
	}
	if evicted != 1 {
		t.Fatalf("evicted=%d", evicted)
	}
}

func TestRevokeAllForLockedAccount(t *testing.T) {
	h := newHarness(0)
	ctx := context.Background()
	a, _ := h.svc.Start(ctx, 7, "quan", domain.ProviderPassword)
	b, _ := h.svc.Start(ctx, 7, "quan", domain.ProviderGoogle)
	if err := h.svc.RevokeAll(ctx, 7, domain.RevokeAccountLocked); err != nil {
		t.Fatal(err)
	}
	for _, tok := range []string{a.SessionToken, b.SessionToken} {
		if _, ok := h.store.access[tok]; ok {
			t.Fatal("session survived account lock")
		}
	}
}
