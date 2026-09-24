package workflows

import (
	"fmt"
	"sort"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	SignalStateChanged       = "state-changed"
	ActivityLoad             = "crm.contract.load"
	ActivityExpire           = "crm.contract.expire"
	ActivityRemind           = "crm.contract.remind"
	loopsBeforeContinueAsNew = 200
	expireBackoff            = 30 * time.Second
)

func WorkflowID(contractID uint) string { return fmt.Sprintf("contract-%d", contractID) }

type LifecycleInput struct {
	ContractID uint            `json:"contract_id"`
	Reminders  []time.Duration `json:"reminders"`
	SentFor    *time.Time      `json:"sent_for,omitempty"`
	Sent       []time.Duration `json:"sent,omitempty"`
}

type ExpireInput struct {
	ContractID uint      `json:"contract_id"`
	At         time.Time `json:"at"`
}

type RemindInput struct {
	ContractID uint          `json:"contract_id"`
	Remaining  time.Duration `json:"remaining"`
}

func active(status string) bool {
	return status == model.ContractDraft || status == model.ContractPending || status == model.ContractApproved
}

func ContractLifecycle(ctx workflow.Context, in LifecycleInput) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    10,
		},
	})
	signals := workflow.GetSignalChannel(ctx, SignalStateChanged)

	reminders := append([]time.Duration(nil), in.Reminders...)
	sort.Slice(reminders, func(i, j int) bool { return reminders[i] > reminders[j] }) // furthest first

	sentFor := in.SentFor
	sent := map[time.Duration]bool{}
	for _, r := range in.Sent {
		sent[r] = true
	}

	wait := func(d time.Duration) bool {
		timerCtx, cancel := workflow.WithCancel(ctx)
		defer cancel()
		timer := workflow.NewTimer(timerCtx, d)
		fired := false
		sel := workflow.NewSelector(ctx)
		sel.AddFuture(timer, func(f workflow.Future) { fired = f.Get(ctx, nil) == nil })
		sel.AddReceive(signals, func(c workflow.ReceiveChannel, _ bool) { c.Receive(ctx, nil) })
		sel.Select(ctx)
		return fired
	}
	expire := func() error {
		var expired bool
		in := ExpireInput{ContractID: in.ContractID, At: workflow.Now(ctx)}
		if err := workflow.ExecuteActivity(ctx, ActivityExpire, in).Get(ctx, &expired); err != nil {
			return err
		}
		if !expired {
			wait(expireBackoff)
		}
		return nil
	}

	for loop := 0; loop < loopsBeforeContinueAsNew; loop++ {
		for signals.ReceiveAsync(nil) {
		}

		var st port.ContractLifecycleState
		if err := workflow.ExecuteActivity(ctx, ActivityLoad, in.ContractID).Get(ctx, &st); err != nil {
			return err
		}
		if !st.Exists || st.Status == model.ContractRejected {
			return nil
		}
		if !active(st.Status) || st.EndDate == nil {
			signals.Receive(ctx, nil)
			continue
		}

		end, now := *st.EndDate, workflow.Now(ctx)
		if sentFor == nil || !sentFor.Equal(end) {
			sentFor = &end
			sent = map[time.Duration]bool{}
			for _, r := range reminders {
				if !end.Add(-r).After(now) {
					sent[r] = true
				}
			}
		}
		if !end.After(now) {
			if err := expire(); err != nil {
				return err
			}
			continue
		}

		wake, remind := end, time.Duration(0)
		if st.Status == model.ContractApproved {
			for _, r := range reminders {
				if t := end.Add(-r); !sent[r] && t.Before(wake) {
					wake, remind = t, r
				}
			}
		}

		if !wait(wake.Sub(now)) {
			continue
		}

		if remind > 0 {
			err := workflow.ExecuteActivity(ctx, ActivityRemind, RemindInput{ContractID: in.ContractID, Remaining: remind}).Get(ctx, nil)
			if err != nil {
				return err
			}
			sent[remind] = true
			continue
		}
		if err := expire(); err != nil {
			return err
		}
	}

	next := in
	next.SentFor = sentFor
	next.Sent = next.Sent[:0:0]
	for r := range sent {
		next.Sent = append(next.Sent, r)
	}
	return workflow.NewContinueAsNewError(ctx, ContractLifecycle, next)
}
