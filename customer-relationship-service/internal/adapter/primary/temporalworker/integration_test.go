package temporalworker

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/orchestrator"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	tc "github.com/JIeeiroSst/customer-relationship-service/internal/infrastructure/temporal"
	"github.com/JIeeiroSst/customer-relationship-service/internal/workflows"
	"go.uber.org/fx/fxtest"
)

// crm is an in-memory contract lifecycle use case.
type crm struct {
	mu       sync.Mutex
	state    port.ContractLifecycleState
	expired  time.Time
	reminded []time.Duration
}

func (c *crm) State(context.Context, uint) (port.ContractLifecycleState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state, nil
}

func (c *crm) Expire(_ context.Context, _ uint, at time.Time) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state.Status != model.ContractApproved || c.state.EndDate == nil || c.state.EndDate.After(at) {
		return false, nil
	}
	c.state.Status, c.expired = model.ContractExpired, time.Now()
	return true, nil
}

func (c *crm) Remind(_ context.Context, _ uint, remaining time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reminded = append(c.reminded, remaining)
	return nil
}

func (c *crm) Resync(context.Context) (int, error) { return 0, nil }

func (c *crm) get() (port.ContractLifecycleState, time.Time, []time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state, c.expired, append([]time.Duration(nil), c.reminded...)
}

func eventually(t *testing.T, what string, within time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// TestAgainstARealTemporalServer runs the worker and the orchestrator against
// a live server: TEMPORAL_TEST_ADDRESS=localhost:7233 go test ./...
func TestAgainstARealTemporalServer(t *testing.T) {
	addr := os.Getenv("TEMPORAL_TEST_ADDRESS")
	if addr == "" {
		t.Skip("TEMPORAL_TEST_ADDRESS not set")
	}
	cfg := &config.Config{Temporal: config.TemporalConfig{
		Address:   addr,
		Namespace: "default",
		TaskQueue: "crm-test-" + time.Now().Format("150405.000"),
		Reminders: []time.Duration{6 * time.Second, 3 * time.Second},
	}}

	lc := fxtest.NewLifecycle(t)
	client, err := tc.New(lc, cfg)
	if err != nil {
		t.Fatal(err)
	}
	c := &crm{}
	Register(lc, client, NewActivities(c), c)
	lc.RequireStart()
	defer lc.RequireStop()
	orch := orchestrator.New(client)
	ctx := context.Background()

	// Approved, ends in 9 seconds; reminders at 3s (6s before) and 6s (3s before).
	end := time.Now().Add(9 * time.Second)
	c.mu.Lock()
	c.state = port.ContractLifecycleState{Exists: true, Status: model.ContractApproved, EndDate: &end}
	c.mu.Unlock()

	// Sync is idempotent: SignalWithStart on a running workflow must not fail or duplicate it.
	for i := 0; i < 3; i++ {
		if err := orch.Sync(ctx, 42); err != nil {
			t.Fatalf("sync %d: %v", i, err)
		}
	}

	eventually(t, "both reminders", 30*time.Second, func() bool { _, _, r := c.get(); return len(r) == 2 })
	_, _, reminded := c.get()
	if reminded[0] != 6*time.Second || reminded[1] != 3*time.Second {
		t.Fatalf("reminders = %v", reminded)
	}

	eventually(t, "expiry", 30*time.Second, func() bool { s, _, _ := c.get(); return s.Status == model.ContractExpired })
	_, expiredAt, _ := c.get()
	if d := expiredAt.Sub(end); d < 0 || d > 5*time.Second {
		t.Fatalf("expired %v after the end date, want within a few seconds", d)
	}

	// Renewed: extended by a few seconds and signalled, it expires again on the new date.
	newEnd := time.Now().Add(4 * time.Second)
	c.mu.Lock()
	c.state = port.ContractLifecycleState{Exists: true, Status: model.ContractApproved, EndDate: &newEnd}
	c.mu.Unlock()
	if err := orch.Sync(ctx, 42); err != nil {
		t.Fatal(err)
	}
	eventually(t, "second expiry", 30*time.Second, func() bool { s, _, _ := c.get(); return s.Status == model.ContractExpired })

	// Deleted: the workflow completes.
	c.mu.Lock()
	c.state = port.ContractLifecycleState{}
	c.mu.Unlock()
	if err := orch.Sync(ctx, 42); err != nil {
		t.Fatal(err)
	}
	eventually(t, "workflow completion", 30*time.Second, func() bool {
		desc, err := client.Client.DescribeWorkflowExecution(ctx, workflows.WorkflowID(42), "")
		return err == nil && desc.WorkflowExecutionInfo.CloseTime != nil
	})
}
