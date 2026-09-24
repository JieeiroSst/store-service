package workflows

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

const day = 24 * time.Hour

var t0 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// fakeCRM is the world the activities see.
type fakeCRM struct {
	mu       sync.Mutex
	state    port.ContractLifecycleState
	expired  []time.Time
	reminded map[time.Duration]time.Time // offset -> when
	reminds  []time.Time
	env      *testsuite.TestWorkflowEnvironment
}

func (f *fakeCRM) set(status string, end *time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = port.ContractLifecycleState{Exists: true, Status: status, EndDate: end}
}

func (f *fakeCRM) gone() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = port.ContractLifecycleState{}
}

func newEnv(t *testing.T, crm *fakeCRM) *testsuite.TestWorkflowEnvironment {
	env := (&testsuite.WorkflowTestSuite{}).NewTestWorkflowEnvironment()
	env.SetStartTime(t0)
	crm.env = env
	crm.reminded = map[time.Duration]time.Time{}

	env.RegisterActivityWithOptions(func(context.Context, uint) (port.ContractLifecycleState, error) {
		crm.mu.Lock()
		defer crm.mu.Unlock()
		return crm.state, nil
	}, activity.RegisterOptions{Name: ActivityLoad})

	env.RegisterActivityWithOptions(func(_ context.Context, in ExpireInput) (bool, error) {
		crm.mu.Lock()
		defer crm.mu.Unlock()
		if !crm.state.Exists || !active(crm.state.Status) || crm.state.EndDate == nil || crm.state.EndDate.After(in.At) {
			return false, nil
		}
		crm.state.Status = model.ContractExpired
		crm.expired = append(crm.expired, env.Now())
		return true, nil
	}, activity.RegisterOptions{Name: ActivityExpire})

	env.RegisterActivityWithOptions(func(_ context.Context, in RemindInput) error {
		crm.mu.Lock()
		defer crm.mu.Unlock()
		crm.reminded[in.Remaining] = env.Now()
		crm.reminds = append(crm.reminds, env.Now())
		return nil
	}, activity.RegisterOptions{Name: ActivityRemind})
	return env
}

func at(d time.Duration) time.Time { return t0.Add(d) }
func ptr(t time.Time) *time.Time   { return &t }

func signal(env *testsuite.TestWorkflowEnvironment) { env.SignalWorkflow(SignalStateChanged, nil) }

func assertTimes(t *testing.T, what string, got []time.Time, want ...time.Time) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %v, want %v", what, got, want)
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Fatalf("%s[%d]: got %v, want %v", what, i, got[i], want[i])
		}
	}
}

func finish(t *testing.T, env *testsuite.TestWorkflowEnvironment) {
	t.Helper()
	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatal(err)
	}
}

var reminders = []time.Duration{30 * day, 7 * day, day}

func TestLifecycle_RemindsThenExpiresThenWaitsForRenewal(t *testing.T) {
	crm := &fakeCRM{}
	env := newEnv(t, crm)
	crm.set(model.ContractApproved, ptr(at(10*day)))

	// Expires at day 10. On day 11 it is renewed until day 13, and on day 20 deleted.
	env.RegisterDelayedCallback(func() {
		crm.set(model.ContractApproved, ptr(at(13*day)))
		signal(env)
	}, 11*day)
	env.RegisterDelayedCallback(func() { crm.gone(); signal(env) }, 20*day)

	env.ExecuteWorkflow(ContractLifecycle, LifecycleInput{ContractID: 5, Reminders: reminders})
	finish(t, env)

	// First cycle: the 30-day reminder is already past when the workflow
	// starts, so it is skipped; 7 and 1 days before day 10 remain. Second
	// cycle (end day 13): only the 1-day reminder is still ahead of day 11.
	assertTimes(t, "reminders", crm.reminds, at(3*day), at(9*day), at(12*day))
	assertTimes(t, "expiries", crm.expired, at(10*day), at(13*day))
}

func TestLifecycle_DoesNotRemindDraftOrPendingButStillExpiresThem(t *testing.T) {
	for _, status := range []string{model.ContractDraft, model.ContractPending} {
		crm := &fakeCRM{}
		env := newEnv(t, crm)
		crm.set(status, ptr(at(10*day)))
		env.RegisterDelayedCallback(func() { crm.gone(); signal(env) }, 15*day)

		env.ExecuteWorkflow(ContractLifecycle, LifecycleInput{ContractID: 5, Reminders: reminders})
		finish(t, env)

		if len(crm.reminds) != 0 {
			t.Fatalf("%s: reminded %v", status, crm.reminds)
		}
		assertTimes(t, status+" expiries", crm.expired, at(10*day))
	}
}

func TestLifecycle_ReactsToAChangedEndDate(t *testing.T) {
	crm := &fakeCRM{}
	env := newEnv(t, crm)
	crm.set(model.ContractApproved, ptr(at(5*day)))

	// Extended on day 1 to day 8: the old timer must not fire on day 5.
	env.RegisterDelayedCallback(func() {
		crm.set(model.ContractApproved, ptr(at(8*day)))
		signal(env)
	}, day)
	env.RegisterDelayedCallback(func() { crm.gone(); signal(env) }, 12*day)

	env.ExecuteWorkflow(ContractLifecycle, LifecycleInput{ContractID: 5, Reminders: []time.Duration{2 * day}})
	finish(t, env)

	assertTimes(t, "expiries", crm.expired, at(8*day))
	assertTimes(t, "reminders", crm.reminds, at(6*day)) // 2 days before the new end date
}

func TestLifecycle_ManyRapidSignalsCauseOneReload(t *testing.T) {
	crm := &fakeCRM{}
	env := newEnv(t, crm)
	crm.set(model.ContractApproved, ptr(at(5*day)))
	env.RegisterDelayedCallback(func() {
		for i := 0; i < 50; i++ {
			signal(env)
		}
		crm.gone()
		signal(env)
	}, day)

	env.ExecuteWorkflow(ContractLifecycle, LifecycleInput{ContractID: 5})
	finish(t, env)
}

func TestLifecycle_FinishesForRejectedAndMissingContracts(t *testing.T) {
	crm := &fakeCRM{}
	env := newEnv(t, crm)
	crm.set(model.ContractRejected, ptr(at(5*day)))
	env.ExecuteWorkflow(ContractLifecycle, LifecycleInput{ContractID: 5})
	finish(t, env)

	crm = &fakeCRM{}
	env = newEnv(t, crm)
	env.ExecuteWorkflow(ContractLifecycle, LifecycleInput{ContractID: 5})
	finish(t, env)
}

func TestLifecycle_ContractWithoutEndDateWaitsForAChange(t *testing.T) {
	crm := &fakeCRM{}
	env := newEnv(t, crm)
	crm.set(model.ContractApproved, nil)
	// Given an end date on day 3, it then behaves as usual.
	env.RegisterDelayedCallback(func() {
		crm.set(model.ContractApproved, ptr(at(6*day)))
		signal(env)
	}, 3*day)
	env.RegisterDelayedCallback(func() { crm.gone(); signal(env) }, 9*day)

	env.ExecuteWorkflow(ContractLifecycle, LifecycleInput{ContractID: 5})
	finish(t, env)
	assertTimes(t, "expiries", crm.expired, at(6*day))
}

func TestLifecycle_ExpireThatFindsNothingToDoJustLooksAgain(t *testing.T) {
	crm := &fakeCRM{}
	env := newEnv(t, crm)
	crm.set(model.ContractApproved, ptr(at(5*day)))

	// Extended in the database without a signal (a missed sync) just before
	// the timer fires: the expire activity refuses, the workflow reloads and
	// sleeps to the new date.
	env.RegisterDelayedCallback(func() { crm.set(model.ContractApproved, ptr(at(9*day))) }, 5*day-time.Second)
	env.RegisterDelayedCallback(func() { crm.gone(); signal(env) }, 12*day)

	env.ExecuteWorkflow(ContractLifecycle, LifecycleInput{ContractID: 5})
	finish(t, env)
	// The refused expire at day 5 costs one backoff, then it sleeps to day 9.
	assertTimes(t, "expiries", crm.expired, at(9*day))
}

func TestLifecycle_DoesNotSpinWhenExpireKeepsRefusing(t *testing.T) {
	crm := &fakeCRM{}
	env := newEnv(t, crm)
	crm.set(model.ContractApproved, ptr(at(-time.Hour))) // already due
	refusals := 0
	env.RegisterActivityWithOptions(func(context.Context, ExpireInput) (bool, error) {
		refusals++ // an application clock that is behind: it never agrees the contract is due
		return false, nil
	}, activity.RegisterOptions{Name: ActivityExpire})
	env.RegisterDelayedCallback(func() { crm.gone(); signal(env) }, 10*time.Minute)

	env.ExecuteWorkflow(ContractLifecycle, LifecycleInput{ContractID: 5})
	finish(t, env)

	// One attempt per backoff (30s) over 10 minutes, not a hot loop.
	if refusals < 10 {
		t.Fatalf("the refusing activity was used only %d times", refusals)
	}
	if refusals > 25 {
		t.Fatalf("expire was attempted %d times in 10 minutes", refusals)
	}
}

func TestWorkflowID(t *testing.T) {
	if got := WorkflowID(42); got != "contract-42" {
		t.Fatalf("got %q", got)
	}
}
