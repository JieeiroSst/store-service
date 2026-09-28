package application

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
)

func TestDispatchSendsOnlyAlertsAndMarksEverything(t *testing.T) {
	st := newStore()
	st.events = []domain.Event{
		{ID: "e1", EventType: domain.EventSessionStarted},
		{ID: "e2", EventType: domain.EventInventoryLow},
		{ID: "e3", EventType: domain.EventOrderFailed},
	}
	n := &fakeNotifier{}
	svc := NewAlertService(fakeTx{}, eventRepo{st}, n)

	sent, err := svc.DispatchPending(context.Background())
	if err != nil || sent != 2 {
		t.Fatalf("sent=%d err=%v", sent, err)
	}
	if len(st.dispatched) != 3 {
		t.Fatalf("dispatched %v", st.dispatched)
	}
	if sent, _ := svc.DispatchPending(context.Background()); sent != 0 {
		t.Fatalf("resent %d", sent)
	}
}

func TestDispatchKeepsProgressWhenNotifierFails(t *testing.T) {
	st := newStore()
	st.events = []domain.Event{
		{ID: "e1", EventType: domain.EventInventoryLow},
		{ID: "e2", EventType: domain.EventOrderFailed},
		{ID: "e3", EventType: domain.EventInventoryEmpty},
	}
	n := &fakeNotifier{failOn: domain.EventOrderFailed}
	svc := NewAlertService(fakeTx{}, eventRepo{st}, n)

	sent, err := svc.DispatchPending(context.Background())
	if err == nil || sent != 1 {
		t.Fatalf("sent=%d err=%v", sent, err)
	}
	if !st.dispatched["e1"] || st.dispatched["e2"] || st.dispatched["e3"] {
		t.Fatalf("dispatched %v", st.dispatched)
	}

	n.failOn = ""
	if sent, err := svc.DispatchPending(context.Background()); err != nil || sent != 2 {
		t.Fatalf("retry sent=%d err=%v", sent, err)
	}
	if len(n.sent) != 3 {
		t.Fatalf("notified %v", n.sent)
	}
}
