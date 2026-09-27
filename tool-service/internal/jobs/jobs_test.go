package jobs

import (
	"context"
	"errors"
	"testing"
	"time"
)

func wait(t *testing.T, m *Manager, id string) Job {
	t.Helper()
	for i := 0; i < 200; i++ {
		if j, _ := m.Get(id); j.Status != Running {
			return j
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("job did not finish")
	return Job{}
}

func TestLifecycle(t *testing.T) {
	m := New()
	ok := m.Start("x", time.Second, func(ctx context.Context, p func(string)) (any, error) { p("half"); return 42, nil })
	if j := wait(t, m, ok.ID); j.Status != Done || j.Result != 42 || j.FinishedAt == nil {
		t.Fatalf("%+v", j)
	}
	bad := m.Start("x", time.Second, func(ctx context.Context, p func(string)) (any, error) { return nil, errors.New("boom") })
	if j := wait(t, m, bad.ID); j.Status != Failed || j.Error != "boom" {
		t.Fatalf("%+v", j)
	}
	if _, found := m.Get("nope"); found {
		t.Fatal("unknown id must not be found")
	}
}

func TestCancelAndTimeout(t *testing.T) {
	m := New()
	j := m.Start("long", time.Minute, func(ctx context.Context, p func(string)) (any, error) { <-ctx.Done(); return nil, ctx.Err() })
	if !m.Cancel(j.ID) {
		t.Fatal("cancel failed")
	}
	if got := wait(t, m, j.ID); got.Status != Canceled {
		t.Fatalf("%+v", got)
	}
	if m.Cancel(j.ID) {
		t.Fatal("cannot cancel twice")
	}
	to := m.Start("slow", 50*time.Millisecond, func(ctx context.Context, p func(string)) (any, error) { <-ctx.Done(); return nil, ctx.Err() })
	if got := wait(t, m, to.ID); got.Status != Failed {
		t.Fatalf("timeout must fail the job: %+v", got)
	}
}

func TestPruneKeepsRunning(t *testing.T) {
	m := New()
	run := m.Start("keep", time.Minute, func(ctx context.Context, p func(string)) (any, error) { <-ctx.Done(); return nil, nil })
	for i := 0; i < keep+20; i++ {
		j := m.Start("q", time.Second, func(ctx context.Context, p func(string)) (any, error) { return nil, nil })
		wait(t, m, j.ID)
	}
	if _, ok := m.Get(run.ID); !ok {
		t.Fatal("running job was pruned")
	}
	if len(m.List()) > keep+1 {
		t.Fatalf("too many jobs retained: %d", len(m.List()))
	}
	m.Cancel(run.ID)
}
