// Package temporalworker runs the order workflow and its activities. It is an inbound adapter: Temporal drives the
// application through the same use cases the HTTP API does.
package temporalworker

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/JIeeiroSst/ticket-service/internal/adapter/temporalx"
)

const (
	reminderLead = 24 * time.Hour
	pollFloor    = 5 * time.Second
)

// OrderWorkflow is the life of one order:
//
//	pending ── the hold runs out ──▶ expired (the tickets go back on sale)
//	   │   └─ cancelled / refunded ─▶ over
//	   └─ paid ──▶ documents made and kept ──▶ a day before the event, the reminder ──▶ over
//
func OrderWorkflow(ctx workflow.Context, in temporalx.OrderInput) error {
	log := workflow.GetLogger(ctx)
	quick := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{InitialInterval: time.Second, BackoffCoefficient: 2, MaximumInterval: time.Minute,
			NonRetryableErrorTypes: []string{errPermanent}},
	})
	documents := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{InitialInterval: 5 * time.Second, BackoffCoefficient: 2, MaximumInterval: 5 * time.Minute,
			MaximumAttempts: 12, NonRetryableErrorTypes: []string{errPermanent}},
	})
	wake := workflow.GetSignalChannel(ctx, temporalx.SignalOrderChanged)

	advance := func() (temporalx.Progress, error) {
		var p temporalx.Progress
		err := workflow.ExecuteActivity(quick, temporalx.ActivityAdvance, in.OrderID).Get(ctx, &p)
		return p, err
	}

	// 1. the hold
	var p temporalx.Progress
	for {
		var err error
		if p, err = advance(); err != nil {
			return err
		}
		if p.Status != "pending" {
			break
		}
		wait := p.ExpiresAt.Sub(workflow.Now(ctx))
		if p.RetryAfter > 0 {
			wait = p.RetryAfter
		}
		sleepOrWake(ctx, wake, max(wait, pollFloor))
	}
	if p.Status != "paid" {
		log.Info("order ended", "order", in.OrderID, "status", p.Status)
		return nil
	}

	// 2. paid: the documents
	if err := workflow.ExecuteActivity(documents, temporalx.ActivityDocuments, in.OrderID).Get(ctx, nil); err != nil {
		log.Warn("documents not made; the sweeper will retry", "order", in.OrderID, "err", err)
	}

	// 3. the reminder, unless the order or the event ends first
	for {
		if p.EventCancelled {
			return nil
		}
		wait := p.EventStartsAt.Add(-reminderLead).Sub(workflow.Now(ctx))
		if wait <= 0 {
			break
		}
		sleepOrWake(ctx, wake, wait)
		var err error
		if p, err = advance(); err != nil {
			return err
		}
		if p.Status != "paid" {
			log.Info("order ended before its reminder", "order", in.OrderID, "status", p.Status)
			return nil
		}
	}
	return workflow.ExecuteActivity(quick, temporalx.ActivityRemind, in.OrderID).Get(ctx, nil)
}

func sleepOrWake(ctx workflow.Context, wake workflow.ReceiveChannel, d time.Duration) {
	tctx, cancel := workflow.WithCancel(ctx)
	defer cancel()
	timer := workflow.NewTimer(tctx, d)
	sel := workflow.NewSelector(ctx)
	sel.AddFuture(timer, func(workflow.Future) {})
	sel.AddReceive(wake, func(c workflow.ReceiveChannel, _ bool) { c.Receive(ctx, nil) })
	sel.Select(ctx)
	for wake.ReceiveAsync(nil) {
	}
}
