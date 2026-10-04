package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	httpadapter "github.com/JIeeiroSst/networking-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/redisstore"
	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/snapshot"
	"github.com/JIeeiroSst/networking-service/internal/application"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
	"github.com/alicebob/miniredis/v2"
	"github.com/hashicorp/consul/api"
	goredis "github.com/redis/go-redis/v9"
)

type nopMetrics struct{}

func (nopMetrics) CheckRun(domain.CheckType, domain.HealthStatus) {}
func (nopMetrics) SessionInvalidated(string)                      {}
func (nopMetrics) BlockingQueryWoken()                            {}

type stack struct {
	srv    *httptest.Server
	client *api.Client
}

func eachBackend(t *testing.T, mutate func(*config.Config), fn func(t *testing.T, s *stack)) {
	for _, backend := range []string{"memory", "redis"} {
		t.Run(backend, func(t *testing.T) {
			var store port.StateStore
			if backend == "memory" {
				store = memory.NewStore()
			} else {
				store = newRedisStore(t, miniredis.RunT(t))
			}
			fn(t, newStack(t, store, mutate))
		})
	}
}

func newRedisStore(t *testing.T, mr *miniredis.Miniredis) *redisstore.Store {
	t.Helper()
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	s := redisstore.New(rdb, redisstore.Options{Prefix: "test:", PollInterval: 50 * time.Millisecond}, time.Now)
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		s.Close()
		_ = rdb.Close()
	})
	return s
}

func newStack(t *testing.T, store port.StateStore, mutate func(*config.Config)) *stack {
	t.Helper()
	cfg := config.FromEnv()
	cfg.Agent.NodeName = "agent-1"
	cfg.Agent.NodeAddress = "10.1.0.5"
	cfg.ACL.DefaultPolicy = "allow"
	cfg.Snapshot.DataDir = ""
	if mutate != nil {
		mutate(cfg)
	}
	blocker := application.NewBlocker(cfg, nopMetrics{})
	catalog := application.NewCatalogService(store, blocker, cfg)
	sessions := application.NewSessionService(store, blocker, nopMetrics{}, catalog)
	storage, _ := snapshot.New(cfg)
	h := httpadapter.NewHandler(
		catalog,
		application.NewKVService(store, blocker, cfg),
		sessions,
		application.NewIntentionService(store, blocker, cfg),
		application.NewSnapshotter(store, storage, cfg),
		application.NewAuthorizer(cfg),
	)
	if err := catalog.RegisterSelf(); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(httpadapter.NewRouter(h))
	t.Cleanup(srv.Close)
	client, err := api.NewClient(&api.Config{Address: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	return &stack{srv: srv, client: client}
}

func TestKV(t *testing.T) {
	eachBackend(t, nil, func(t *testing.T, s *stack) {
		kv := s.client.KV()

		if _, err := kv.Put(&api.KVPair{Key: "config/app", Value: []byte(`{"port":8080}`), Flags: 42}, nil); err != nil {
			t.Fatal(err)
		}
		_, _ = kv.Put(&api.KVPair{Key: "config/db/url", Value: []byte("postgres://")}, nil)

		p, meta, err := kv.Get("config/app", nil)
		if err != nil || p == nil {
			t.Fatalf("get: %v %v", p, err)
		}
		if string(p.Value) != `{"port":8080}` || p.Flags != 42 || meta.LastIndex == 0 {
			t.Fatalf("pair = %+v, meta = %+v", p, meta)
		}
		if missing, _, err := kv.Get("nope", nil); err != nil || missing != nil {
			t.Fatalf("missing key: %v %v", missing, err)
		}

		list, _, err := kv.List("config/", nil)
		if err != nil || len(list) != 2 {
			t.Fatalf("list = %v, %v", list, err)
		}
		keys, _, err := kv.Keys("config/", "/", nil)
		if err != nil || strings.Join(keys, ",") != "config/app,config/db/" {
			t.Fatalf("keys = %v, %v", keys, err)
		}

		ok, _, err := kv.CAS(&api.KVPair{Key: "config/app", Value: []byte("x"), ModifyIndex: p.ModifyIndex - 1}, nil)
		if err != nil || ok {
			t.Fatalf("stale CAS = %v, %v", ok, err)
		}
		ok, _, err = kv.CAS(&api.KVPair{Key: "config/app", Value: []byte("y"), ModifyIndex: p.ModifyIndex}, nil)
		if err != nil || !ok {
			t.Fatalf("CAS = %v, %v", ok, err)
		}

		if _, err := kv.DeleteTree("config/", nil); err != nil {
			t.Fatal(err)
		}
		if list, _, _ := kv.List("config/", nil); len(list) != 0 {
			t.Fatalf("left after DeleteTree: %v", list)
		}
	})
}

func TestKVBlockingQuery(t *testing.T) {
	eachBackend(t, nil, func(t *testing.T, s *stack) {
		kv := s.client.KV()
		_, meta, _ := kv.Get("feature/flag", nil)

		got := make(chan *api.KVPair, 1)
		go func() {
			p, _, err := kv.Get("feature/flag", &api.QueryOptions{WaitIndex: meta.LastIndex, WaitTime: 5 * time.Second})
			if err != nil {
				t.Error(err)
			}
			got <- p
		}()
		time.Sleep(100 * time.Millisecond)
		_, _ = kv.Put(&api.KVPair{Key: "feature/flag", Value: []byte("on")}, nil)
		select {
		case p := <-got:
			if p == nil || string(p.Value) != "on" {
				t.Fatalf("watch saw %+v", p)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("blocking query did not return after the write")
		}
	})
}

func TestLeaderElection(t *testing.T) {
	eachBackend(t, nil, func(t *testing.T, s *stack) {
		l1, err := s.client.LockKey("service/worker/leader")
		if err != nil {
			t.Fatal(err)
		}
		lost1, err := l1.Lock(nil)
		if err != nil || lost1 == nil {
			t.Fatalf("l1 lock: %v", err)
		}

		l2, _ := s.client.LockOpts(&api.LockOptions{Key: "service/worker/leader", LockWaitTime: time.Second})
		acquired := make(chan struct{})
		go func() {
			if ch, err := l2.Lock(nil); err == nil && ch != nil {
				close(acquired)
			}
		}()
		select {
		case <-acquired:
			t.Fatal("l2 got the lock while l1 held it")
		case <-time.After(300 * time.Millisecond):
		}
		if err := l1.Unlock(); err != nil {
			t.Fatal(err)
		}
		select {
		case <-acquired:
		case <-time.After(5 * time.Second):
			t.Fatal("l2 never got the lock after l1 released it")
		}
		_ = l2.Unlock()
	})
}

func TestServiceDiscovery(t *testing.T) {
	eachBackend(t, nil, func(t *testing.T, s *stack) {
		agent := s.client.Agent()

		for i, id := range []string{"web-1", "web-2"} {
			err := agent.ServiceRegister(&api.AgentServiceRegistration{
				ID: id, Name: "web", Tags: []string{"v1", "primary"}[:i+1], Port: 8080 + i, Address: "10.0.0." + string(rune('1'+i)),
				Meta:  map[string]string{"version": "1"},
				Check: &api.AgentServiceCheck{CheckID: id + "-ttl", TTL: "30s"},
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		if err := agent.UpdateTTL("web-1-ttl", "ok", api.HealthPassing); err != nil {
			t.Fatal(err)
		}

		services, err := agent.Services()
		if err != nil || len(services) != 2 || services["web-1"].Port != 8080 || services["web-1"].Meta["version"] != "1" {
			t.Fatalf("agent services = %+v, %v", services, err)
		}

		names, _, err := s.client.Catalog().Services(nil)
		if err != nil || strings.Join(names["web"], ",") != "primary,v1" {
			t.Fatalf("catalog services = %v, %v", names, err)
		}

		passing, meta, err := s.client.Health().Service("web", "", true, nil)
		if err != nil || len(passing) != 1 || passing[0].Service.ID != "web-1" {
			t.Fatalf("passing = %+v, %v", passing, err)
		}
		if passing[0].Node.Node != "agent-1" || passing[0].Checks.AggregatedStatus() != api.HealthPassing {
			t.Fatalf("entry = %+v", passing[0])
		}
		if meta.LastIndex == 0 || !meta.KnownLeader {
			t.Fatalf("meta = %+v", meta)
		}

		tagged, _, err := s.client.Catalog().Service("web", "primary", nil)
		if err != nil || len(tagged) != 1 || tagged[0].ServiceID != "web-2" || tagged[0].ServicePort != 8081 {
			t.Fatalf("tagged = %+v, %v", tagged, err)
		}

		checks, _, err := s.client.Health().Checks("web", nil)
		if err != nil || len(checks) != 2 {
			t.Fatalf("checks = %+v, %v", checks, err)
		}
		for _, c := range checks {
			if c.Type != "ttl" || c.ServiceName != "web" {
				t.Fatalf("check = %+v", c)
			}
		}

		if err := agent.EnableServiceMaintenance("web-1", "deploy"); err != nil {
			t.Fatal(err)
		}
		if passing, _, _ := s.client.Health().Service("web", "", true, nil); len(passing) != 0 {
			t.Fatalf("instance in maintenance still passing: %+v", passing)
		}
		_ = agent.DisableServiceMaintenance("web-1")

		if err := agent.ServiceDeregister("web-2"); err != nil {
			t.Fatal(err)
		}
		if all, _, _ := s.client.Health().Service("web", "", false, nil); len(all) != 1 {
			t.Fatalf("after deregister: %+v", all)
		}
	})
}

func TestCatalogRegisterExternalService(t *testing.T) {
	eachBackend(t, nil, func(t *testing.T, s *stack) {
		_, err := s.client.Catalog().Register(&api.CatalogRegistration{
			Node: "rds", Address: "db.example.internal", NodeMeta: map[string]string{"external": "true"},
			Service: &api.AgentService{ID: "postgres", Service: "postgres", Port: 5432},
			Check: &api.AgentCheck{
				CheckID: "pg-tcp", Name: "pg tcp", ServiceID: "postgres", Status: api.HealthPassing,
				Definition: api.HealthCheckDefinition{TCP: "db.example.internal:5432", IntervalDuration: 10 * time.Second},
			},
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		nodes, _, err := s.client.Catalog().Nodes(&api.QueryOptions{NodeMeta: map[string]string{"external": "true"}})
		if err != nil || len(nodes) != 1 || nodes[0].Node != "rds" {
			t.Fatalf("nodes = %+v, %v", nodes, err)
		}
		checks, _, _ := s.client.Health().Node("rds", nil)
		if len(checks) != 1 || checks[0].Definition.TCP != "db.example.internal:5432" || checks[0].Definition.IntervalDuration != 10*time.Second {
			t.Fatalf("checks = %+v", checks)
		}

		if _, err := s.client.Catalog().Deregister(&api.CatalogDeregistration{Node: "rds"}, nil); err != nil {
			t.Fatal(err)
		}
		if node, _, _ := s.client.Catalog().Node("rds", nil); node != nil {
			t.Fatalf("node still there: %+v", node)
		}
	})
}

func TestUnsupportedCheckIsRejected(t *testing.T) {
	eachBackend(t, nil, func(t *testing.T, s *stack) {
		err := s.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
			Name: "grpc", Check: &api.AgentServiceCheck{GRPC: "localhost:9090", Interval: "10s"},
		})
		if err == nil || !strings.Contains(err.Error(), "400") {
			t.Fatalf("err = %v, want 400", err)
		}
	})
}

func TestSessions(t *testing.T) {
	eachBackend(t, nil, func(t *testing.T, s *stack) {
		sess := s.client.Session()
		id, _, err := sess.Create(&api.SessionEntry{Name: "worker", TTL: "15s", Behavior: api.SessionBehaviorDelete}, nil)
		if err != nil {
			t.Fatal(err)
		}
		info, _, err := sess.Info(id, nil)
		if err != nil || info == nil || info.Node != "agent-1" || info.TTL != "15s" || info.LockDelay != 15*time.Second {
			t.Fatalf("info = %+v, %v", info, err)
		}
		if len(info.NodeChecks) != 1 || info.NodeChecks[0] != "serfHealth" {
			t.Fatalf("default checks = %v", info.NodeChecks)
		}
		if _, _, err := sess.Renew(id, nil); err != nil {
			t.Fatal(err)
		}

		ok, _, err := s.client.KV().Acquire(&api.KVPair{Key: "eph", Session: id}, nil)
		if err != nil || !ok {
			t.Fatalf("acquire = %v, %v", ok, err)
		}
		if _, err := sess.Destroy(id, nil); err != nil {
			t.Fatal(err)
		}
		if p, _, _ := s.client.KV().Get("eph", nil); p != nil {
			t.Fatal("delete-behavior session left its key")
		}
		if renewed, _, err := sess.Renew(id, nil); err != nil || renewed != nil {
			t.Fatalf("renew of destroyed session = %+v, %v", renewed, err)
		}

		noChecks, _, err := sess.CreateNoChecks(&api.SessionEntry{Name: "free"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		info, _, _ = sess.Info(noChecks, nil)
		if len(info.NodeChecks) != 0 {
			t.Fatalf("CreateNoChecks got checks %v", info.NodeChecks)
		}
	})
}

func TestIntentions(t *testing.T) {
	eachBackend(t, func(c *config.Config) { c.Intentions.DefaultAllow = false }, func(t *testing.T, s *stack) {
		conn := s.client.Connect()

		allowed, _, err := conn.IntentionCheck(&api.IntentionCheck{Source: "web", Destination: "db"}, nil)
		if err != nil || allowed {
			t.Fatalf("default deny: allowed = %v, %v", allowed, err)
		}
		if _, err := conn.IntentionUpsert(&api.Intention{SourceName: "web", DestinationName: "db", Action: api.IntentionActionAllow}, nil); err != nil {
			t.Fatal(err)
		}
		allowed, _, _ = conn.IntentionCheck(&api.IntentionCheck{Source: "web", Destination: "db"}, nil)
		if !allowed {
			t.Fatal("web -> db not allowed after upsert")
		}

		ixn, _, err := conn.IntentionGetExact("web", "db", nil)
		if err != nil || ixn.Precedence != 9 || ixn.Action != api.IntentionActionAllow {
			t.Fatalf("exact = %+v, %v", ixn, err)
		}
		matches, _, err := conn.IntentionMatch(&api.IntentionMatch{By: api.IntentionMatchDestination, Names: []string{"db"}}, nil)
		if err != nil || len(matches["db"]) != 1 {
			t.Fatalf("match = %+v, %v", matches, err)
		}
		if _, err := conn.IntentionDeleteExact("web", "db", nil); err != nil {
			t.Fatal(err)
		}
		list, _, _ := conn.Intentions(nil)
		if len(list) != 0 {
			t.Fatalf("left = %+v", list)
		}
	})
}

func TestAgentSelfAndStatus(t *testing.T) {
	eachBackend(t, nil, func(t *testing.T, s *stack) {
		name, err := s.client.Agent().NodeName()
		if err != nil || name != "agent-1" {
			t.Fatalf("node name = %q, %v", name, err)
		}
		leader, err := s.client.Status().Leader()
		if err != nil || leader != "10.1.0.5:8300" {
			t.Fatalf("leader = %q, %v", leader, err)
		}
		dcs, err := s.client.Catalog().Datacenters()
		if err != nil || len(dcs) != 1 || dcs[0] != "dc1" {
			t.Fatalf("dcs = %v, %v", dcs, err)
		}
	})
}

func TestACL(t *testing.T) {
	eachBackend(t, func(c *config.Config) {
		c.ACL.DefaultPolicy = "deny"
		c.ACL.ManagementToken = "root-token"
		c.ACL.ReadTokens = []string{"read-token"}
	}, func(t *testing.T, s *stack) {
		if _, err := s.client.KV().Put(&api.KVPair{Key: "k"}, nil); err == nil || !strings.Contains(err.Error(), "403") {
			t.Fatalf("anonymous write: %v", err)
		}
		if _, err := s.client.KV().Put(&api.KVPair{Key: "k", Value: []byte("v")}, &api.WriteOptions{Token: "read-token"}); err == nil {
			t.Fatal("read token could write")
		}
		if _, err := s.client.KV().Put(&api.KVPair{Key: "k", Value: []byte("v")}, &api.WriteOptions{Token: "root-token"}); err != nil {
			t.Fatal(err)
		}
		p, _, err := s.client.KV().Get("k", &api.QueryOptions{Token: "read-token"})
		if err != nil || string(p.Value) != "v" {
			t.Fatalf("read with read token: %v, %v", p, err)
		}

		resp, err := http.Get(s.srv.URL + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("/health = %v, %v", resp, err)
		}
		resp.Body.Close()
	})
}

func TestWrongDatacenter(t *testing.T) {
	eachBackend(t, nil, func(t *testing.T, s *stack) {
		_, _, err := s.client.KV().Get("k", &api.QueryOptions{Datacenter: "dc9"})
		if err == nil || !strings.Contains(err.Error(), "No path to datacenter") {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestKVValueLimit(t *testing.T) {
	eachBackend(t, func(c *config.Config) { c.KV.MaxValueBytes = 8 }, func(t *testing.T, s *stack) {
		_, err := s.client.KV().Put(&api.KVPair{Key: "big", Value: []byte("123456789")}, nil)
		if err == nil || !strings.Contains(err.Error(), "413") {
			t.Fatalf("err = %v, want 413", err)
		}
	})
}

func TestReplicasShareStateThroughRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	a := newStack(t, newRedisStore(t, mr), nil)
	b := newStack(t, newRedisStore(t, mr), nil)

	err := a.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID: "pay-1", Name: "payment", Address: "10.0.0.7", Port: 9000,
		Check: &api.AgentServiceCheck{CheckID: "pay-ttl", TTL: "30s", Status: api.HealthPassing},
	})
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := b.client.Health().Service("payment", "", true, nil)
	if err != nil || len(entries) != 1 || entries[0].Service.Address != "10.0.0.7" {
		t.Fatalf("replica b sees %+v, %v", entries, err)
	}

	_, meta, _ := b.client.KV().Get("feature", nil)
	got := make(chan string, 1)
	go func() {
		p, _, err := b.client.KV().Get("feature", &api.QueryOptions{WaitIndex: meta.LastIndex, WaitTime: 5 * time.Second})
		if err != nil || p == nil {
			got <- ""
			return
		}
		got <- string(p.Value)
	}()
	time.Sleep(100 * time.Millisecond)
	if _, err := a.client.KV().Put(&api.KVPair{Key: "feature", Value: []byte("on")}, nil); err != nil {
		t.Fatal(err)
	}
	select {
	case v := <-got:
		if v != "on" {
			t.Fatalf("blocking query on b returned %q", v)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("write on a did not wake a blocking query on b")
	}

	l1, _ := a.client.LockKey("job/leader")
	if _, err := l1.Lock(nil); err != nil {
		t.Fatal(err)
	}
	l2, _ := b.client.LockOpts(&api.LockOptions{Key: "job/leader", LockTryOnce: true, LockWaitTime: 200 * time.Millisecond})
	if ch, err := l2.Lock(nil); err != nil || ch != nil {
		t.Fatalf("replica b took a lock held through replica a: %v", err)
	}
	_ = l1.Unlock()
}

func TestRedisOutageIsAnErrorNotAnEmptyAnswer(t *testing.T) {
	mr := miniredis.RunT(t)
	s := newStack(t, newRedisStore(t, mr), nil)
	mr.Close()
	for _, path := range []string{"/v1/health/service/web", "/v1/kv/cfg", "/v1/catalog/services"} {
		resp, err := http.Get(s.srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("%s with Redis down = %d, want 500", path, resp.StatusCode)
		}
	}
}
