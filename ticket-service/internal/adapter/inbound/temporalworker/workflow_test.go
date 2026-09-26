package temporalworker_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/JIeeiroSst/ticket-service/internal/adapter/inbound/temporalworker"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/temporalx"
)

// order is a stand-in for an order in the database: its state, and a record of what the workflow asked for.
type order struct {
	mu        sync.Mutex
	env       *testsuite.TestWorkflowEnvironment
	status    string
	expiresAt time.Time
	startsAt  time.Time
	cancelled bool // the event
	inFlight  int  // pending checks answered "a payment is in flight" past the hold

	advances []time.Time
	docs     []time.Time
	reminds  []time.Time
	failNext int // Advance fails this many times, transiently
	docFail  bool
}

func (o *order) set(status string) {
	o.mu.Lock()
	o.status = status
	o.mu.Unlock()
}

func (o *order) AdvanceOrder(_ context.Context, id int64) (temporalx.Progress, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	now := o.env.Now()
	o.advances = append(o.advances, now)
	if o.failNext > 0 {
		o.failNext--
		return temporalx.Progress{}, errors.New("database is busy")
	}
	if o.status == "pending" && !now.Before(o.expiresAt) {
		if o.inFlight > 0 {
			o.inFlight--
			return temporalx.Progress{Status: "pending", ExpiresAt: o.expiresAt, RetryAfter: 30 * time.Second}, nil
		}
		o.status = "expired"
	}
	return temporalx.Progress{Status: o.status, ExpiresAt: o.expiresAt, EventStartsAt: o.startsAt, EventCancelled: o.cancelled}, nil
}

func (o *order) GenerateOrderDocuments(_ context.Context, id int64) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.docs = append(o.docs, o.env.Now())
	if o.docFail {
		return errors.New("upload-service is down")
	}
	return nil
}

func (o *order) RemindOrder(_ context.Context, id int64) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.reminds = append(o.reminds, o.env.Now())
	return nil
}

// start builds a workflow test environment around an order that expires in ten minutes for an event three days away.
func start(t *testing.T) (*testsuite.TestWorkflowEnvironment, *order, time.Time) {
	t.Helper()
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()
	t0 := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	env.SetStartTime(t0)
	o := &order{env: env, status: "pending", expiresAt: t0.Add(10 * time.Minute), startsAt: t0.Add(72 * time.Hour)}
	env.RegisterWorkflowWithOptions(temporalworker.OrderWorkflow, workflowName())
	env.RegisterActivityWithOptions(o.AdvanceOrder, activity.RegisterOptions{Name: temporalx.ActivityAdvance})
	env.RegisterActivityWithOptions(o.GenerateOrderDocuments, activity.RegisterOptions{Name: temporalx.ActivityDocuments})
	env.RegisterActivityWithOptions(o.RemindOrder, activity.RegisterOptions{Name: temporalx.ActivityRemind})
	return env, o, t0
}

func run(t *testing.T, env *testsuite.TestWorkflowEnvironment) {
	t.Helper()
	env.ExecuteWorkflow(temporalx.WorkflowOrder, temporalx.OrderInput{OrderID: 7})
	if !env.IsWorkflowCompleted() {
		t.Fatal("the workflow did not finish")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow: %v", err)
	}
}

func TestUnpaidOrderExpiresWhenItsHoldEnds(t *testing.T) {
	env, o, t0 := start(t)
	run(t, env)
	if o.status != "expired" {
		t.Fatalf("status %s", o.status)
	}
	// looked at the start, then exactly when the hold ended
	if len(o.advances) != 2 || !o.advances[0].Equal(t0) || !o.advances[1].Equal(t0.Add(10*time.Minute)) {
		t.Fatalf("looked at the order at %v", o.advances)
	}
	if len(o.docs) != 0 || len(o.reminds) != 0 {
		t.Fatal("an order that expired gets no documents and no reminder")
	}
}

func TestPaidOrderGetsDocumentsThenAReminderADayBefore(t *testing.T) {
	env, o, t0 := start(t)
	env.RegisterDelayedCallback(func() {
		o.set("paid")
		env.SignalWorkflow(temporalx.SignalOrderChanged, nil)
	}, 3*time.Minute)
	run(t, env)
	if len(o.docs) != 1 || !o.docs[0].Equal(t0.Add(3*time.Minute)) {
		t.Fatalf("documents made at %v, want once, the moment it was paid", o.docs)
	}
	if want := t0.Add(48 * time.Hour); len(o.reminds) != 1 || !o.reminds[0].Equal(want) {
		t.Fatalf("reminded at %v, want once at %v (a day before the event)", o.reminds, want)
	}
	// it did not wait for the hold to end once it was paid
	for _, a := range o.advances {
		if a.Equal(t0.Add(10 * time.Minute)) {
			t.Fatal("the workflow kept waiting for a hold that no longer mattered")
		}
	}
}

func TestCancelledOrderEndsTheWorkflowAtOnce(t *testing.T) {
	env, o, t0 := start(t)
	env.RegisterDelayedCallback(func() {
		o.set("cancelled")
		env.SignalWorkflow(temporalx.SignalOrderChanged, nil)
	}, 2*time.Minute)
	run(t, env)
	if len(o.advances) != 2 || !o.advances[1].Equal(t0.Add(2*time.Minute)) || len(o.docs)+len(o.reminds) != 0 {
		t.Fatalf("advances %v docs %v reminds %v", o.advances, o.docs, o.reminds)
	}
}

func TestRefundBeforeTheReminderCancelsIt(t *testing.T) {
	env, o, _ := start(t)
	env.RegisterDelayedCallback(func() { o.set("paid"); env.SignalWorkflow(temporalx.SignalOrderChanged, nil) }, time.Minute)
	env.RegisterDelayedCallback(func() { o.set("refunded"); env.SignalWorkflow(temporalx.SignalOrderChanged, nil) }, 10*time.Hour)
	run(t, env)
	if len(o.docs) != 1 || len(o.reminds) != 0 {
		t.Fatalf("docs %v reminds %v: a refunded order must not be reminded", o.docs, o.reminds)
	}
}

func TestNoReminderForACancelledEvent(t *testing.T) {
	env, o, _ := start(t)
	env.RegisterDelayedCallback(func() { o.set("paid"); env.SignalWorkflow(temporalx.SignalOrderChanged, nil) }, time.Minute)
	env.RegisterDelayedCallback(func() {
		o.mu.Lock()
		o.cancelled = true
		o.mu.Unlock()
		env.SignalWorkflow(temporalx.SignalOrderChanged, nil)
	}, time.Hour)
	run(t, env)
	if len(o.reminds) != 0 {
		t.Fatalf("reminded for a cancelled event: %v", o.reminds)
	}
}

func TestAPaymentInFlightPastTheHoldIsWaitedFor(t *testing.T) {
	env, o, t0 := start(t)
	o.inFlight = 2
	run(t, env)
	// at the end of the hold, then 30 seconds later, then 30 more: only then expired
	want := []time.Time{t0, t0.Add(10 * time.Minute), t0.Add(10*time.Minute + 30*time.Second), t0.Add(11 * time.Minute)}
	if len(o.advances) != len(want) {
		t.Fatalf("looked at %v", o.advances)
	}
	for i, w := range want {
		if !o.advances[i].Equal(w) {
			t.Fatalf("look %d at %v, want %v", i, o.advances[i], w)
		}
	}
	if o.status != "expired" {
		t.Fatal(o.status)
	}
}

func TestTransientFailuresAreRetried(t *testing.T) {
	env, o, _ := start(t)
	o.failNext = 3
	run(t, env)
	if o.status != "expired" || len(o.advances) < 5 {
		t.Fatalf("status %s after %d looks", o.status, len(o.advances))
	}
}

func TestDocumentsThatKeepFailingDoNotStopTheReminder(t *testing.T) {
	env, o, _ := start(t)
	o.docFail = true
	env.RegisterDelayedCallback(func() { o.set("paid"); env.SignalWorkflow(temporalx.SignalOrderChanged, nil) }, time.Minute)
	run(t, env)
	if len(o.docs) != 12 { // the retry policy's limit
		t.Fatalf("%d attempts to make the documents", len(o.docs))
	}
	if len(o.reminds) != 1 {
		t.Fatalf("the reminder is not held up by the documents: %v", o.reminds)
	}
}

func TestAnOrderThatIsGoneEndsTheWorkflow(t *testing.T) {
	env, o, _ := start(t)
	o.set("gone")
	run(t, env)
	if len(o.advances) != 1 {
		t.Fatalf("looked %d times", len(o.advances))
	}
}

func TestAPermanentFailureFailsTheWorkflowWithoutRetrying(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()
	calls := 0
	env.RegisterWorkflowWithOptions(temporalworker.OrderWorkflow, workflowName())
	env.RegisterActivityWithOptions(func(context.Context, int64) (temporalx.Progress, error) {
		calls++
		return temporalx.Progress{}, temporal.NewNonRetryableApplicationError("bad order", "permanent", nil)
	}, activity.RegisterOptions{Name: temporalx.ActivityAdvance})
	env.ExecuteWorkflow(temporalx.WorkflowOrder, temporalx.OrderInput{OrderID: 1})
	if env.GetWorkflowError() == nil || calls != 1 {
		t.Fatalf("error %v after %d calls", env.GetWorkflowError(), calls)
	}
}

func workflowName() workflow.RegisterOptions {
	return workflow.RegisterOptions{Name: temporalx.WorkflowOrder}
}
