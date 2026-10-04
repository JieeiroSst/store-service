package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/networking-service/internal/domain"
)

type nopMetrics struct{}

func (nopMetrics) CheckRun(domain.CheckType, domain.HealthStatus) {}
func (nopMetrics) SessionInvalidated(string)                      {}
func (nopMetrics) BlockingQueryWoken()                            {}

type fakeProber struct {
	mu     sync.Mutex
	status domain.HealthStatus
	calls  int
}

func (f *fakeProber) Probe(context.Context, domain.Check) (domain.HealthStatus, string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.status, "probed"
}

type fakeLeader struct{ leader bool }

func (l *fakeLeader) IsLeader() bool { return l.leader }

type fixture struct {
	cfg        *config.Config
	store      *memory.Store
	catalog    *CatalogService
	kv         *KVService
	sessions   *SessionService
	intentions *IntentionService
	runner     *HealthRunner
	leader     *fakeLeader
	prober     *fakeProber
	now        time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	cfg := config.FromEnv()
	cfg.Agent.NodeName = "agent-1"
	cfg.Blocking.DefaultWait = time.Second
	cfg.Blocking.MaxWait = 2 * time.Second
	f := &fixture{cfg: cfg, prober: &fakeProber{status: domain.HealthPassing}}
	f.now = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := func() time.Time { return f.now }
	f.store = memory.NewStoreWithClock(clock)

	blocker := NewBlocker(cfg, nopMetrics{})
	f.catalog = NewCatalogService(f.store, blocker, cfg)
	f.catalog.now = clock
	f.kv = NewKVService(f.store, blocker, cfg)
	f.kv.now = clock
	f.sessions = NewSessionService(f.store, blocker, nopMetrics{}, f.catalog)
	f.sessions.now = clock
	f.intentions = NewIntentionService(f.store, blocker, cfg)
	f.leader = &fakeLeader{leader: true}
	f.runner = NewHealthRunner(f.store, f.prober, nopMetrics{}, f.leader, f.sessions, cfg)
	f.runner.now = clock
	if err := f.catalog.RegisterSelf(); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *fixture) settle() {
	f.runner.wg.Wait()
}

func TestBlockingQueryWakesOnWrite(t *testing.T) {
	f := newFixture(t)
	_, m, _ := f.kv.Get(context.Background(), QueryOptions{}, "cfg")

	done := make(chan *domain.KVPair)
	go func() {
		p, _, _ := f.kv.Get(context.Background(), QueryOptions{MinIndex: m.LastIndex, Wait: 5 * time.Second}, "cfg")
		done <- p
	}()
	time.Sleep(50 * time.Millisecond)
	if _, err := f.kv.Put(domain.KVPair{Key: "cfg", Value: []byte("v")}, nil, "", ""); err != nil {
		t.Fatal(err)
	}
	select {
	case p := <-done:
		if p == nil || string(p.Value) != "v" {
			t.Fatalf("woke with %+v", p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("blocking query did not wake on write")
	}
}

func TestBlockingQueryTimesOut(t *testing.T) {
	f := newFixture(t)
	_, m, _ := f.kv.Get(context.Background(), QueryOptions{}, "cfg")
	start := time.Now()
	_, m2, _ := f.kv.Get(context.Background(), QueryOptions{MinIndex: m.LastIndex, Wait: 100 * time.Millisecond}, "cfg")
	if el := time.Since(start); el < 100*time.Millisecond || el > time.Second {
		t.Fatalf("returned after %s", el)
	}
	if m2.LastIndex != m.LastIndex {
		t.Fatalf("index changed without a write: %d -> %d", m.LastIndex, m2.LastIndex)
	}
}

func TestBlockingQueryReturnsAtOnceForFutureIndex(t *testing.T) {
	f := newFixture(t)
	start := time.Now()
	f.kv.Get(context.Background(), QueryOptions{MinIndex: 1 << 40, Wait: 5 * time.Second}, "cfg")
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("blocked on an index the server never reached (state reset)")
	}
}

func TestAgentServiceRegistrationDefaults(t *testing.T) {
	f := newFixture(t)
	err := f.catalog.AgentRegisterService(AgentServiceRegistration{
		Service: domain.Service{Name: "api", Port: 8080},
		Checks:  []domain.Check{{TTL: 30 * time.Second}, {TCP: "localhost:1", Interval: time.Second}},
	})
	if err != nil {
		t.Fatal(err)
	}
	svc, err := f.catalog.AgentService("api")
	if err != nil || svc.Node != "agent-1" {
		t.Fatalf("svc = %+v, err = %v", svc, err)
	}
	ids := map[string]domain.Check{}
	agentChecks, _ := f.catalog.AgentChecks()
	for _, c := range agentChecks {
		ids[c.ID] = c
	}
	if c, ok := ids["service:api:1"]; !ok || c.Type != domain.CheckTTL || c.Status != domain.HealthCritical {
		t.Fatalf("checks = %+v", ids)
	}
	if _, ok := ids["service:api:2"]; !ok {
		t.Fatalf("checks = %+v", ids)
	}

	err = f.catalog.AgentRegisterService(AgentServiceRegistration{
		Service: domain.Service{Name: "api", Port: 8080},
		Checks:  []domain.Check{{TTL: 30 * time.Second}}, ReplaceExistingChecks: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var left []string
	agentChecks, _ = f.catalog.AgentChecks()
	for _, c := range agentChecks {
		if c.ServiceID == "api" {
			left = append(left, c.ID)
		}
	}
	if len(left) != 1 || left[0] != "service:api" {
		t.Fatalf("checks after replace = %v", left)
	}
}

func TestTTLCheckLifecycle(t *testing.T) {
	f := newFixture(t)
	_ = f.catalog.AgentRegisterService(AgentServiceRegistration{
		Service: domain.Service{Name: "api"},
		Checks:  []domain.Check{{ID: "ttl", TTL: 10 * time.Second}},
	})
	if err := f.catalog.UpdateTTL("ttl", domain.HealthPassing, "alive"); err != nil {
		t.Fatal(err)
	}
	healthy := func() int {
		e, _, _ := f.catalog.ServiceNodes(context.Background(), QueryOptions{}, "api", nil, true)
		return len(e)
	}
	if healthy() != 1 {
		t.Fatal("instance not passing after TTL pass")
	}

	f.now = f.now.Add(11 * time.Second)
	f.runner.Tick(context.Background())
	if healthy() != 0 {
		t.Fatal("instance still passing after its TTL expired")
	}
	c, _ := f.catalog.AgentChecks()
	for _, chk := range c {
		if chk.ID == "ttl" && chk.Output != "TTL expired (last output before timeout follows): alive" {
			t.Fatalf("output = %q", chk.Output)
		}
	}

	var invalid *domain.InvalidError
	if err := f.catalog.UpdateTTL(domain.SerfHealthID, domain.HealthPassing, ""); !errors.As(err, &invalid) {
		t.Fatalf("UpdateTTL on a non-TTL check: err = %v", err)
	}
}

func TestHealthRunnerProbesAndReaps(t *testing.T) {
	f := newFixture(t)
	f.prober.status = domain.HealthCritical
	_ = f.catalog.AgentRegisterService(AgentServiceRegistration{
		Service: domain.Service{Name: "api"},
		Checks: []domain.Check{{
			ID: "http", HTTP: "http://api/health", Interval: 10 * time.Second,
			DeregisterCriticalServiceAfter: time.Minute,
		}},
	})

	f.runner.Tick(context.Background())
	f.settle()
	f.runner.Tick(context.Background())
	f.settle()
	if f.prober.calls != 1 {
		t.Fatalf("probe calls = %d, want 1", f.prober.calls)
	}

	f.now = f.now.Add(61 * time.Second)
	f.runner.Tick(context.Background())
	f.settle()
	if _, err := f.catalog.AgentService("api"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("service not reaped after DeregisterCriticalServiceAfter")
	}
}

func TestSessionTTLReaping(t *testing.T) {
	f := newFixture(t)
	id, err := f.sessions.Create(domain.Session{TTL: 10 * time.Second, NodeChecks: []string{domain.SerfHealthID}})
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := f.kv.Put(domain.KVPair{Key: "lock"}, nil, id, ""); !ok {
		t.Fatal("acquire failed")
	}
	f.now = f.now.Add(15 * time.Second)
	_ = f.sessions.ReapExpired()
	if s, _, _ := f.sessions.Info(context.Background(), QueryOptions{}, id); s == nil {
		t.Fatal("session reaped before 2x TTL")
	}
	f.now = f.now.Add(6 * time.Second)
	_ = f.sessions.ReapExpired()
	if s, _, _ := f.sessions.Info(context.Background(), QueryOptions{}, id); s != nil {
		t.Fatal("session not reaped after 2x TTL")
	}
	if p, _, _ := f.kv.Get(context.Background(), QueryOptions{}, "lock"); p.Session != "" {
		t.Fatal("lock still held by an expired session")
	}
}

func TestIntentionPrecedence(t *testing.T) {
	f := newFixture(t)
	f.intentions.defaultAllow = true
	mustUpsert := func(src, dst string, a domain.IntentionAction) {
		t.Helper()
		if err := f.intentions.UpsertExact(domain.Intention{SourceName: src, DestinationName: dst, Action: a}); err != nil {
			t.Fatal(err)
		}
	}
	mustUpsert("*", "db", domain.IntentionDeny)
	mustUpsert("api", "db", domain.IntentionAllow)
	mustUpsert("*", "*", domain.IntentionDeny)

	cases := []struct {
		src, dst string
		want     bool
	}{
		{"api", "db", true},
		{"web", "db", false},
		{"web", "cache", false},
	}
	for _, c := range cases {
		got, _, err := f.intentions.Check(c.src, c.dst)
		if err != nil || got != c.want {
			t.Errorf("Check(%s -> %s) = %v, %v; want %v", c.src, c.dst, got, err, c.want)
		}
	}

	mustUpsert("api", "db", domain.IntentionDeny)
	if list, _, _ := f.intentions.List(context.Background(), QueryOptions{}); len(list) != 3 {
		t.Fatalf("intentions = %d, want 3", len(list))
	}
	if ok, _, _ := f.intentions.Check("api", "db"); ok {
		t.Fatal("update did not take effect")
	}
}

func TestAuthorizer(t *testing.T) {
	cfg := config.FromEnv()
	cfg.ACL.DefaultPolicy = "deny"
	cfg.ACL.ManagementToken = "root"
	cfg.ACL.ReadTokens = []string{"reader"}
	a := NewAuthorizer(cfg)

	cases := []struct {
		token string
		write bool
		ok    bool
	}{
		{"", false, false},
		{"root", true, true},
		{"reader", false, true},
		{"reader", true, false},
		{"wrong", false, false},
	}
	for _, c := range cases {
		err := a.Authorize(c.token, c.write)
		if (err == nil) != c.ok {
			t.Errorf("Authorize(%q, write=%v) = %v", c.token, c.write, err)
		}
	}
}

func TestKVKeysWithSeparator(t *testing.T) {
	f := newFixture(t)
	for _, k := range []string{"app/a", "app/b/c", "app/b/d", "apple"} {
		_, _ = f.kv.Put(domain.KVPair{Key: k}, nil, "", "")
	}
	keys, _, _ := f.kv.Keys(context.Background(), QueryOptions{}, "app/", "/")
	want := []string{"app/a", "app/b/"}
	if len(keys) != len(want) || keys[0] != want[0] || keys[1] != want[1] {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
}

func TestFollowerRunsNoChecks(t *testing.T) {
	f := newFixture(t)
	f.leader.leader = false
	_ = f.catalog.AgentRegisterService(AgentServiceRegistration{
		Service: domain.Service{Name: "api"},
		Checks:  []domain.Check{{ID: "http", HTTP: "http://api/health", Interval: time.Second}},
	})
	if err := f.runner.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	f.settle()
	if f.prober.calls != 0 {
		t.Fatalf("follower probed %d times", f.prober.calls)
	}
	f.leader.leader = true
	_ = f.runner.Tick(context.Background())
	f.settle()
	if f.prober.calls != 1 {
		t.Fatalf("leader probed %d times, want 1", f.prober.calls)
	}
}

type memStorage struct{ snap *domain.Snapshot }

func (m *memStorage) Load() (*domain.Snapshot, error) { return m.snap, nil }
func (m *memStorage) Save(s domain.Snapshot) error    { m.snap = &s; return nil }

func TestSnapshotNeverOverwritesLiveState(t *testing.T) {
	f := newFixture(t)
	_, _ = f.kv.Put(domain.KVPair{Key: "live", Value: []byte("new")}, nil, "", "")
	stale := &memStorage{snap: &domain.Snapshot{Index: 1, KV: []domain.KVPair{{Key: "live", Value: []byte("old")}}}}
	if err := NewSnapshotter(f.store, stale, f.cfg).Restore(); err != nil {
		t.Fatal(err)
	}
	if p, _, _ := f.kv.Get(context.Background(), QueryOptions{}, "live"); string(p.Value) != "new" {
		t.Fatalf("live state overwritten by a stale snapshot: %q", p.Value)
	}

	empty := memory.NewStore()
	if err := NewSnapshotter(empty, stale, f.cfg).Restore(); err != nil {
		t.Fatal(err)
	}
	if _, p, _ := empty.KVGet("live"); p == nil || string(p.Value) != "old" {
		t.Fatal("snapshot not restored into an empty store")
	}
}
