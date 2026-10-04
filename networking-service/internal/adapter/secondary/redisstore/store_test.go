package redisstore

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/storetest"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func newReplica(t *testing.T, mr *miniredis.Miniredis, now func() time.Time) (*Store, *goredis.Client) {
	t.Helper()
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	s := New(rdb, Options{Prefix: "test:", PollInterval: 50 * time.Millisecond}, now)
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		s.Close()
		_ = rdb.Close()
	})
	return s, rdb
}

func TestContract(t *testing.T) {
	storetest.Run(t, func(t *testing.T, now func() time.Time) port.StateStore {
		s, _ := newReplica(t, miniredis.RunT(t), now)
		return s
	})
}

func waitFired(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("%s: watch never fired", what)
	}
}

func TestWritesWakeOtherReplicas(t *testing.T) {
	mr := miniredis.RunT(t)
	a, _ := newReplica(t, mr, time.Now)
	b, _ := newReplica(t, mr, time.Now)

	ch := b.WatchCh()
	if _, err := a.KVApply(domain.KVRequest{Op: domain.KVSet, Pair: domain.KVPair{Key: "cfg", Value: []byte("v1")}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	waitFired(t, ch, "pub/sub")
	if _, p, err := b.KVGet("cfg"); err != nil || p == nil || string(p.Value) != "v1" {
		t.Fatalf("replica b read %+v, %v", p, err)
	}
}

func TestPollingCatchesMissedEvents(t *testing.T) {
	mr := miniredis.RunT(t)
	s, rdb := newReplica(t, mr, time.Now)
	ch := s.WatchCh()
	if err := rdb.Set(context.Background(), "test:index", 99, 0).Err(); err != nil {
		t.Fatal(err)
	}
	waitFired(t, ch, "poll")
}

func TestConcurrentCASAcrossReplicas(t *testing.T) {
	mr := miniredis.RunT(t)
	a, _ := newReplica(t, mr, time.Now)
	b, _ := newReplica(t, mr, time.Now)

	const perWorker, workers = 10, 4
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		s := a
		if i%2 == 1 {
			s = b
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < perWorker; {
				_, p, err := s.KVGet("counter")
				if err != nil {
					t.Error(err)
					return
				}
				var cur int
				var cas uint64
				if p != nil {
					cur, _ = strconv.Atoi(string(p.Value))
					cas = p.ModifyIndex
				}
				ok, err := s.KVApply(domain.KVRequest{Op: domain.KVCAS, CAS: cas,
					Pair: domain.KVPair{Key: "counter", Value: []byte(strconv.Itoa(cur + 1))}}, time.Now())
				if err != nil {
					t.Error(err)
					return
				}
				if ok {
					n++
				}
			}
		}()
	}
	wg.Wait()
	_, p, _ := a.KVGet("counter")
	if got, _ := strconv.Atoi(string(p.Value)); got != perWorker*workers {
		t.Fatalf("counter = %d, want %d (lost updates)", got, perWorker*workers)
	}
}

func TestLockIsExclusiveAcrossReplicas(t *testing.T) {
	mr := miniredis.RunT(t)
	a, _ := newReplica(t, mr, time.Now)
	b, _ := newReplica(t, mr, time.Now)
	if err := a.EnsureRegistration(domain.CatalogRegistration{Node: domain.Node{Name: "n1", Address: "10.0.0.1"}}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"sa", "sb"} {
		if err := a.SessionCreate(domain.Session{ID: id, Node: "n1", Behavior: domain.SessionRelease}); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	results := make([]bool, 2)
	for i, s := range []*Store{a, b} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := s.KVApply(domain.KVRequest{Op: domain.KVLock, Session: []string{"sa", "sb"}[i], Pair: domain.KVPair{Key: "leader"}}, time.Now())
			if err != nil {
				t.Error(err)
			}
			results[i] = ok
		}()
	}
	wg.Wait()
	if results[0] == results[1] {
		t.Fatalf("lock results = %v, want exactly one winner", results)
	}
}

func TestElectorSingleLeaderAndFailover(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	e1 := NewElector(rdb, "test:", 300*time.Millisecond)
	e2 := NewElector(rdb, "test:", 300*time.Millisecond)
	e1.Start()
	e2.Start()
	defer e2.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for !e1.IsLeader() && !e2.IsLeader() {
		if time.Now().After(deadline) {
			t.Fatal("nobody became leader")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if e1.IsLeader() && e2.IsLeader() {
		t.Fatal("two leaders")
	}
	if !e1.IsLeader() {
		e1, e2 = e2, e1
	}
	e1.Stop()
	deadline = time.Now().Add(2 * time.Second)
	for !e2.IsLeader() {
		if time.Now().After(deadline) {
			t.Fatal("no failover after the leader stopped")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
