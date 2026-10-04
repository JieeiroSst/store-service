package storetest

import (
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
)

type Factory func(t *testing.T, now func() time.Time) port.StateStore

func Run(t *testing.T, newStore Factory) {
	cases := map[string]func(*testing.T, Factory){
		"RegistrationIsIdempotent":        registrationIsIdempotent,
		"RegistrationIsAtomic":            registrationIsAtomic,
		"ReRegisterKeepsCheckStatus":      reRegisterKeepsCheckStatus,
		"UpdateCheckTracksCriticalSince":  updateCheckTracksCriticalSince,
		"DeregisterNodeCascades":          deregisterNodeCascades,
		"DeregisterServiceDropsChecks":    deregisterServiceDropsChecks,
		"KVCAS":                           kvCAS,
		"KVListAndDeleteTree":             kvListAndDeleteTree,
		"Locking":                         locking,
		"SessionDeleteBehavior":           sessionDeleteBehavior,
		"CriticalCheckInvalidatesSession": criticalCheckInvalidatesSession,
		"SessionRenewKeepsIndex":          sessionRenewKeepsIndex,
		"IntentionPairIsUnique":           intentionPairIsUnique,
		"SnapshotRoundTrip":               snapshotRoundTrip,
		"Stats":                           stats,
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) { tc(t, newStore) })
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func lastIndex(t *testing.T, s port.StateStore) uint64 {
	t.Helper()
	idx, err := s.LastIndex()
	must(t, err)
	return idx
}

func registerWeb(t *testing.T, s port.StateStore, node, id string, checks ...domain.Check) {
	t.Helper()
	must(t, s.EnsureRegistration(domain.CatalogRegistration{
		Node:    domain.Node{Name: node, Address: "10.0.0.1"},
		Service: &domain.Service{ID: id, Name: "web", Port: 80, Tags: []string{"v1"}},
		Checks:  checks,
	}))
}

func mustSession(t *testing.T, s port.StateStore, id, node string, b domain.SessionBehavior, checks ...string) {
	t.Helper()
	must(t, s.SessionCreate(domain.Session{ID: id, Node: node, Behavior: b, LockDelay: 15 * time.Second, NodeChecks: checks}))
}

func kv(t *testing.T, s port.StateStore, req domain.KVRequest, now time.Time) bool {
	t.Helper()
	ok, err := s.KVApply(req, now)
	must(t, err)
	return ok
}

func registrationIsIdempotent(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	registerWeb(t, s, "n1", "web-1", domain.Check{ID: "c1", ServiceID: "web-1", Status: domain.HealthPassing})
	before := lastIndex(t, s)
	registerWeb(t, s, "n1", "web-1", domain.Check{ID: "c1", ServiceID: "web-1", Status: domain.HealthPassing})
	if after := lastIndex(t, s); after != before {
		t.Fatalf("index moved on a no-op registration: %d -> %d", before, after)
	}
	_, entries, err := s.ServiceNodes("web")
	must(t, err)
	if len(entries) != 1 || len(entries[0].Checks) != 1 || entries[0].Checks[0].ServiceName != "web" {
		t.Fatalf("entries = %+v", entries)
	}
}

func registrationIsAtomic(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	err := s.EnsureRegistration(domain.CatalogRegistration{
		Node:   domain.Node{Name: "n1", Address: "10.0.0.1"},
		Checks: []domain.Check{{ID: "c1", ServiceID: "missing"}},
	})
	if !domain.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
	_, nodes, err := s.Nodes()
	must(t, err)
	if len(nodes) != 0 {
		t.Fatalf("node written despite failed registration: %+v", nodes)
	}
}

func reRegisterKeepsCheckStatus(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	registerWeb(t, s, "n1", "web-1", domain.Check{ID: "c1", ServiceID: "web-1", Type: domain.CheckTTL, TTL: time.Minute})
	_, err := s.UpdateCheck("n1", "c1", domain.HealthPassing, "ok", time.Now().Add(time.Minute))
	must(t, err)
	registerWeb(t, s, "n1", "web-1", domain.Check{ID: "c1", ServiceID: "web-1", Type: domain.CheckTTL, TTL: time.Minute})
	_, checks, err := s.Checks()
	must(t, err)
	if checks[0].Status != domain.HealthPassing {
		t.Fatalf("status = %s, want passing", checks[0].Status)
	}
}

func updateCheckTracksCriticalSince(t *testing.T, newStore Factory) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := newStore(t, func() time.Time { return now })
	registerWeb(t, s, "n1", "web-1", domain.Check{ID: "c1", ServiceID: "web-1", Status: domain.HealthPassing})

	before := lastIndex(t, s)
	c, err := s.UpdateCheck("n1", "c1", domain.HealthPassing, "", time.Time{})
	must(t, err)
	if lastIndex(t, s) != before {
		t.Fatal("index moved on an unchanged check update")
	}

	c, err = s.UpdateCheck("n1", "c1", domain.HealthCritical, "down", time.Time{})
	must(t, err)
	if !c.CriticalSince.Equal(now) {
		t.Fatalf("CriticalSince = %v", c.CriticalSince)
	}
	now = now.Add(time.Minute)
	c, _ = s.UpdateCheck("n1", "c1", domain.HealthCritical, "still down", time.Time{})
	if !c.CriticalSince.Equal(now.Add(-time.Minute)) {
		t.Fatal("CriticalSince moved while the check stayed critical")
	}
	c, _ = s.UpdateCheck("n1", "c1", domain.HealthPassing, "up", time.Time{})
	if !c.CriticalSince.IsZero() {
		t.Fatal("CriticalSince kept after recovery")
	}
	if _, err := s.UpdateCheck("n1", "nope", domain.HealthPassing, "", time.Time{}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func deregisterNodeCascades(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	registerWeb(t, s, "n1", "web-1", domain.Check{ID: "serfHealth", Status: domain.HealthPassing})
	mustSession(t, s, "s1", "n1", domain.SessionRelease, "serfHealth")
	must(t, s.DeregisterNode("n1"))
	_, svcs, _ := s.Services()
	_, checks, _ := s.Checks()
	if len(svcs) != 0 || len(checks) != 0 {
		t.Fatal("services or checks survived node deregistration")
	}
	if _, sess, _ := s.SessionGet("s1"); sess != nil {
		t.Fatal("session survived node deregistration")
	}
	if err := s.DeregisterNode("n1"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func deregisterServiceDropsChecks(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	registerWeb(t, s, "n1", "web-1", domain.Check{ID: "c1", ServiceID: "web-1"}, domain.Check{ID: "node", Status: domain.HealthPassing})
	must(t, s.DeregisterService("n1", "web-1"))
	_, checks, _ := s.Checks()
	if len(checks) != 1 || checks[0].ID != "node" {
		t.Fatalf("checks = %+v, want only the node check", checks)
	}
}

func kvCAS(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	now := time.Now()
	if !kv(t, s, domain.KVRequest{Op: domain.KVCAS, Pair: domain.KVPair{Key: "a", Value: []byte("1")}}, now) {
		t.Fatal("cas=0 on a missing key must create it")
	}
	if kv(t, s, domain.KVRequest{Op: domain.KVCAS, Pair: domain.KVPair{Key: "a", Value: []byte("2")}}, now) {
		t.Fatal("cas=0 on an existing key must fail")
	}
	_, p, err := s.KVGet("a")
	must(t, err)
	if !kv(t, s, domain.KVRequest{Op: domain.KVCAS, CAS: p.ModifyIndex, Pair: domain.KVPair{Key: "a", Value: []byte("3")}}, now) {
		t.Fatal("cas with the current index must succeed")
	}
	_, p2, _ := s.KVGet("a")
	if string(p2.Value) != "3" || p2.CreateIndex != p.CreateIndex || p2.ModifyIndex <= p.ModifyIndex {
		t.Fatalf("pair after cas = %+v", p2)
	}
	if kv(t, s, domain.KVRequest{Op: domain.KVDelCAS, CAS: p.ModifyIndex, Pair: domain.KVPair{Key: "a"}}, now) {
		t.Fatal("delete-cas with a stale index must fail")
	}
	if !kv(t, s, domain.KVRequest{Op: domain.KVDelCAS, CAS: p2.ModifyIndex, Pair: domain.KVPair{Key: "a"}}, now) {
		t.Fatal("delete-cas with the current index must succeed")
	}
	if _, p, _ := s.KVGet("a"); p != nil {
		t.Fatal("key survived delete-cas")
	}
}

func kvListAndDeleteTree(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	for _, k := range []string{"app/b/c", "app/a", "other", "apple"} {
		kv(t, s, domain.KVRequest{Op: domain.KVSet, Pair: domain.KVPair{Key: k, Value: []byte(k)}}, time.Now())
	}
	_, list, err := s.KVList("app/")
	must(t, err)
	if len(list) != 2 || list[0].Key != "app/a" || list[1].Key != "app/b/c" {
		t.Fatalf("list = %+v", list)
	}
	_, all, _ := s.KVList("")
	if len(all) != 4 {
		t.Fatalf("all = %d keys", len(all))
	}
	before := lastIndex(t, s)
	kv(t, s, domain.KVRequest{Op: domain.KVDelTree, Pair: domain.KVPair{Key: "app/"}}, time.Now())
	_, left, _ := s.KVList("")
	if len(left) != 2 || left[0].Key != "apple" || left[1].Key != "other" {
		t.Fatalf("left = %+v", left)
	}
	if lastIndex(t, s) <= before {
		t.Fatal("delete-tree did not move the index")
	}
	before = lastIndex(t, s)
	kv(t, s, domain.KVRequest{Op: domain.KVDelete, Pair: domain.KVPair{Key: "missing"}}, time.Now())
	if lastIndex(t, s) != before {
		t.Fatal("deleting a missing key moved the index")
	}
}

func locking(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	registerWeb(t, s, "n1", "web-1")
	mustSession(t, s, "s1", "n1", domain.SessionRelease)
	mustSession(t, s, "s2", "n1", domain.SessionRelease)
	now := time.Now()
	lock := func(sess string) bool {
		return kv(t, s, domain.KVRequest{Op: domain.KVLock, Session: sess, Pair: domain.KVPair{Key: "leader", Value: []byte(sess)}}, now)
	}

	if !lock("s1") {
		t.Fatal("s1 could not take a free lock")
	}
	if lock("s2") {
		t.Fatal("s2 took a lock held by s1")
	}
	if !lock("s1") {
		t.Fatal("re-acquire by the holder must succeed")
	}
	if _, p, _ := s.KVGet("leader"); p.LockIndex != 1 || p.Session != "s1" {
		t.Fatalf("pair = %+v, want LockIndex 1 held by s1", p)
	}
	if kv(t, s, domain.KVRequest{Op: domain.KVUnlock, Session: "s2", Pair: domain.KVPair{Key: "leader"}}, now) {
		t.Fatal("s2 released a lock it does not hold")
	}

	must(t, s.SessionDestroy("s1", now))
	if lock("s2") {
		t.Fatal("lock taken during lock-delay")
	}
	now = now.Add(16 * time.Second)
	if !lock("s2") {
		t.Fatal("lock not available after lock-delay")
	}
	if _, p, _ := s.KVGet("leader"); p.LockIndex != 2 || p.Session != "s2" {
		t.Fatalf("pair = %+v, want LockIndex 2 held by s2", p)
	}

	if _, err := s.KVApply(domain.KVRequest{Op: domain.KVLock, Session: "nope", Pair: domain.KVPair{Key: "x"}}, now); !errors.Is(err, domain.ErrInvalidSession) {
		t.Fatalf("err = %v, want invalid session", err)
	}
}

func sessionDeleteBehavior(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	registerWeb(t, s, "n1", "web-1")
	mustSession(t, s, "s1", "n1", domain.SessionDelete)
	kv(t, s, domain.KVRequest{Op: domain.KVLock, Session: "s1", Pair: domain.KVPair{Key: "eph"}}, time.Now())
	must(t, s.SessionDestroy("s1", time.Now()))
	if _, p, _ := s.KVGet("eph"); p != nil {
		t.Fatal("key of a delete-behavior session survived")
	}
	if _, list, _ := s.KVList("eph"); len(list) != 0 {
		t.Fatal("deleted key still listed")
	}
}

func criticalCheckInvalidatesSession(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	registerWeb(t, s, "n1", "web-1", domain.Check{ID: "serfHealth", Status: domain.HealthPassing})
	mustSession(t, s, "s1", "n1", domain.SessionRelease, "serfHealth")
	kv(t, s, domain.KVRequest{Op: domain.KVLock, Session: "s1", Pair: domain.KVPair{Key: "k"}}, time.Now())

	_, err := s.UpdateCheck("n1", "serfHealth", domain.HealthCritical, "", time.Time{})
	must(t, err)
	if _, sess, _ := s.SessionGet("s1"); sess != nil {
		t.Fatal("session survived its check going critical")
	}
	if _, p, _ := s.KVGet("k"); p.Session != "" {
		t.Fatal("lock not released")
	}
	if err := s.SessionCreate(domain.Session{ID: "s2", Node: "n1", NodeChecks: []string{"serfHealth"}}); !domain.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func sessionRenewKeepsIndex(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	registerWeb(t, s, "n1", "web-1")
	mustSession(t, s, "s1", "n1", domain.SessionRelease)
	before := lastIndex(t, s)
	exp := time.Now().Add(time.Minute).Truncate(time.Second)
	sess, err := s.SessionRenew("s1", exp)
	must(t, err)
	if !sess.Expires.Equal(exp) || lastIndex(t, s) != before {
		t.Fatalf("renew: expires %v, index %d -> %d", sess.Expires, before, lastIndex(t, s))
	}
	if _, got, _ := s.SessionGet("s1"); !got.Expires.Equal(exp) {
		t.Fatal("renewed expiry not stored")
	}
	if _, err := s.SessionRenew("nope", exp); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func intentionPairIsUnique(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	must(t, s.IntentionUpsert(domain.Intention{ID: "1", SourceName: "a", DestinationName: "b", Action: domain.IntentionAllow, Precedence: 9}))
	must(t, s.IntentionUpsert(domain.Intention{ID: "2", SourceName: "*", DestinationName: "*", Action: domain.IntentionDeny, Precedence: 5}))
	err := s.IntentionUpsert(domain.Intention{ID: "3", SourceName: "a", DestinationName: "b", Action: domain.IntentionDeny})
	if !domain.IsInvalid(err) {
		t.Fatalf("err = %v, want duplicate", err)
	}
	must(t, s.IntentionUpsert(domain.Intention{ID: "1", SourceName: "a", DestinationName: "b", Action: domain.IntentionDeny, Precedence: 9}))
	_, list, err := s.IntentionList()
	must(t, err)
	if len(list) != 2 || list[0].ID != "1" || list[0].Action != domain.IntentionDeny {
		t.Fatalf("list = %+v", list)
	}
	must(t, s.IntentionDelete("2"))
	if err := s.IntentionDelete("2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func snapshotRoundTrip(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	registerWeb(t, s, "n1", "web-1", domain.Check{ID: "c1", ServiceID: "web-1", Status: domain.HealthPassing})
	kv(t, s, domain.KVRequest{Op: domain.KVSet, Pair: domain.KVPair{Key: "k", Value: []byte("v")}}, time.Now())
	snap, err := s.Snapshot()
	must(t, err)

	r := newStore(t, time.Now)
	must(t, r.Restore(snap))
	if lastIndex(t, r) != snap.Index {
		t.Fatalf("index %d, want %d", lastIndex(t, r), snap.Index)
	}
	if _, p, _ := r.KVGet("k"); p == nil || string(p.Value) != "v" {
		t.Fatal("kv not restored")
	}
	if _, list, _ := r.KVList(""); len(list) != 1 {
		t.Fatal("kv listing not restored")
	}
	if _, e, _ := r.ServiceNodes("web"); len(e) != 1 || len(e[0].Checks) != 1 {
		t.Fatalf("catalog not restored: %+v", e)
	}
}

func stats(t *testing.T, newStore Factory) {
	s := newStore(t, time.Now)
	registerWeb(t, s, "n1", "web-1", domain.Check{ID: "c1", ServiceID: "web-1", Status: domain.HealthPassing})
	registerWeb(t, s, "n1", "web-2", domain.Check{ID: "c2", ServiceID: "web-2"})
	kv(t, s, domain.KVRequest{Op: domain.KVSet, Pair: domain.KVPair{Key: "k"}}, time.Now())
	st, err := s.Stats()
	must(t, err)
	if st.Nodes != 1 || st.Services != 1 || st.Instances != 2 || st.Keys != 1 ||
		st.Checks[domain.HealthPassing] != 1 || st.Checks[domain.HealthCritical] != 1 {
		t.Fatalf("stats = %+v", st)
	}
}
