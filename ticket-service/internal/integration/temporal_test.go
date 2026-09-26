package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.temporal.io/sdk/testsuite"

	"github.com/JIeeiroSst/ticket-service/internal/adapter/inbound/temporalworker"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/temporalx"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

// workflowEnv runs the real order workflow against the real service and database, with virtual time: waiting ten
// minutes for a hold to end takes no time, and the callbacks change the world (real database, real payment) at the
// virtual moments they name.
func workflowEnv(t *testing.T, p *pod) *testsuite.TestWorkflowEnvironment {
	t.Helper()
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()
	env.SetStartTime(time.Now())
	temporalworker.Register(env, &temporalworker.Activities{Orders: p.orders, Documents: p.docs})
	return env
}

func runOrder(t *testing.T, env *testsuite.TestWorkflowEnvironment, orderID int64) {
	t.Helper()
	env.ExecuteWorkflow(temporalx.WorkflowOrder, temporalx.OrderInput{OrderID: orderID})
	if !env.IsWorkflowCompleted() || env.GetWorkflowError() != nil {
		t.Fatalf("workflow: completed=%v err=%v", env.IsWorkflowCompleted(), env.GetWorkflowError())
	}
}

func orderStatus(t *testing.T, p *pod, id int64) domain.OrderStatus {
	t.Helper()
	x, err := p.repo.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return x.Status
}

// The hold of an order that is never paid ends, and its tickets go back on sale, by the workflow and not the sweeper.
func TestWorkflowExpiresAnUnpaidOrder(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 5, nil)
	o, err := p.orders.Reserve(ctx, user(5), cmd(e.ID, tt.ID, 3, "wf-expire"))
	if err != nil {
		t.Fatal(err)
	}
	if c := typeCounters(t, p, tt.ID); c.held != 3 {
		t.Fatalf("held: %+v", c)
	}
	env := workflowEnv(t, p)
	// the hold really runs out (the database clock is the real one): ten virtual minutes in, its time is past
	env.RegisterDelayedCallback(func() {
		if _, err := p.pool.Exec(ctx, `update ticket_order set expires_at = now() - interval '1 second' where id = $1`, o.ID); err != nil {
			t.Error(err)
		}
	}, 5*time.Minute)
	runOrder(t, env, o.ID)

	if orderStatus(t, p, o.ID) != domain.OrderExpired {
		t.Fatalf("the order is %s", orderStatus(t, p, o.ID).Name())
	}
	if c := typeCounters(t, p, tt.ID); c.held != 0 || c.available != 5 {
		t.Fatalf("the tickets did not come back: %+v", c)
	}
	checkLedger(t, p, tt.ID)
}

// Paid: documents are made straight away and kept for the buyer; a day before the event the buyer is reminded.
func TestWorkflowFollowsAPaidOrder(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 5, nil)
	o, err := p.orders.Reserve(ctx, user(5), cmd(e.ID, tt.ID, 2, "wf-paid"))
	if err != nil {
		t.Fatal(err)
	}
	env := workflowEnv(t, p)
	env.RegisterDelayedCallback(func() {
		if _, err := p.orders.Pay(ctx, user(5), o.ID, inbound.PayCommand{Method: domain.MethodWallet}); err != nil {
			t.Error(err)
		}
		env.SignalWorkflow(temporalx.SignalOrderChanged, nil)
	}, 2*time.Minute)
	runOrder(t, env, o.ID)

	if n := len(p.store.of(5)); n != 2 {
		t.Fatalf("%d files kept for the buyer, want the invoice and the tickets", n)
	}
	var reminders int
	if err := p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 5 and kind = 'event_reminder'`).Scan(&reminders); err != nil || reminders != 1 {
		t.Fatalf("reminders: %d %v", reminders, err)
	}
	// the sweeper finds nothing left to do: the workflow did all of it, once
	if n, _ := p.docs.GenerateMissing(ctx); n != 0 {
		t.Fatalf("documents made twice: %d", n)
	}
	if n, _ := p.orders.SendReminders(ctx); n != 0 {
		t.Fatalf("reminded twice: %d", n)
	}
}

// A buyer who gives the order up before paying ends its workflow; a refund ends it before the reminder.
func TestWorkflowStopsWhenTheOrderEnds(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 5, nil)

	o1, _ := p.orders.Reserve(ctx, user(5), cmd(e.ID, tt.ID, 1, "wf-cancel"))
	env := workflowEnv(t, p)
	env.RegisterDelayedCallback(func() {
		if _, err := p.orders.Cancel(ctx, user(5), o1.ID); err != nil {
			t.Error(err)
		}
		env.SignalWorkflow(temporalx.SignalOrderChanged, nil)
	}, time.Minute)
	runOrder(t, env, o1.ID)
	if orderStatus(t, p, o1.ID) != domain.OrderCancelled || len(p.store.of(5)) != 0 {
		t.Fatalf("cancelled order: %s, %d files", orderStatus(t, p, o1.ID).Name(), len(p.store.of(5)))
	}

	o2 := paidOrder(t, p, 6, e, tt, 1)
	env2 := workflowEnv(t, p)
	env2.RegisterDelayedCallback(func() {
		if _, err := p.orders.Cancel(ctx, organizer, o2.ID); err != nil { // the organizer: this event does not refund buyers
			t.Error(err)
		}
		env2.SignalWorkflow(temporalx.SignalOrderChanged, nil)
	}, 10*time.Hour)
	runOrder(t, env2, o2.ID)
	var reminders int
	_ = p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 6 and kind = 'event_reminder'`).Scan(&reminders)
	if orderStatus(t, p, o2.ID) != domain.OrderRefunded || reminders != 0 {
		t.Fatalf("refunded order: %s, %d reminders", orderStatus(t, p, o2.ID).Name(), reminders)
	}
	checkLedger(t, p, tt.ID)
}

// An order with no workflow (it was made before Temporal was on, or the start was lost) is still expired by the sweeper.
func TestSweeperIsTheSafetyNet(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 5, nil)
	o, _ := p.orders.Reserve(ctx, user(5), cmd(e.ID, tt.ID, 1, "no-wf"))
	if _, err := p.pool.Exec(ctx, `update ticket_order set expires_at = now() - interval '1 minute' where id = $1`, o.ID); err != nil {
		t.Fatal(err)
	}
	if n, err := p.orders.ReleaseExpired(ctx); err != nil || n != 1 {
		t.Fatalf("sweep: %d %v", n, err)
	}
	// and Advance, run by a workflow that comes late, agrees the order is over
	if pr, err := p.orders.Advance(ctx, o.ID); err != nil || pr.Status != "expired" {
		t.Fatalf("advance: %+v %v", pr, err)
	}
	if pr, err := p.orders.Advance(ctx, 999_999); err != nil || pr.Status != "gone" {
		t.Fatalf("advance of nothing: %+v %v", pr, err)
	}
}

type fakeLifecycle struct {
	mu      sync.Mutex
	tracked []int64
	nudged  []int64
}

func (f *fakeLifecycle) Track(_ context.Context, id int64, _ time.Time) error {
	f.mu.Lock()
	f.tracked = append(f.tracked, id)
	f.mu.Unlock()
	return nil
}

func (f *fakeLifecycle) Nudge(_ context.Context, id int64) error {
	f.mu.Lock()
	f.nudged = append(f.nudged, id)
	f.mu.Unlock()
	return nil
}

func (f *fakeLifecycle) counts() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.tracked), len(f.nudged)
}

func eventually(t *testing.T, what string, ok func() bool) {
	t.Helper()
	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if ok() {
			return
		}
	}
	t.Fatalf("timed out waiting for %s", what)
}

// The service starts a workflow when tickets are held, and wakes it when the order is paid or given up.
func TestServiceDrivesTheLifecycle(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	lc := &fakeLifecycle{}
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000, lifecycle: lc})
	e, tt := seedEvent(t, p, 5, nil)

	o, err := p.orders.Reserve(ctx, user(5), cmd(e.ID, tt.ID, 1, "lc-1"))
	if err != nil {
		t.Fatal(err)
	}
	eventually(t, "the workflow to be started", func() bool { tr, _ := lc.counts(); return tr == 1 })
	if lc.tracked[0] != o.ID {
		t.Fatalf("tracked %v", lc.tracked)
	}
	if _, err := p.orders.Pay(ctx, user(5), o.ID, inbound.PayCommand{Method: domain.MethodWallet}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the workflow to be told of the payment", func() bool { _, n := lc.counts(); return n == 1 })
	if _, err := p.orders.Cancel(ctx, organizer, o.ID); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the workflow to be told of the refund", func() bool { _, n := lc.counts(); return n == 2 })

	// a request that is refused starts nothing
	before, _ := lc.counts()
	if _, err := p.orders.Reserve(ctx, user(6), cmd(e.ID, tt.ID, 99, "lc-2")); err == nil {
		t.Fatal("more tickets than exist")
	}
	time.Sleep(100 * time.Millisecond)
	if after, _ := lc.counts(); after != before {
		t.Fatalf("a refused reservation started a workflow (%d -> %d)", before, after)
	}
}
