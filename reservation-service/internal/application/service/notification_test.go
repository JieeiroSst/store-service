package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

// memInbox is an in-memory notification repository with the same de-duplication rule as the SQL one.
type memInbox struct {
	outbound.NotificationRepository
	all      []domain.Notification
	pushed   map[int64]bool
	failures map[int64]int
	leased   map[int64]bool
}

func (m *memInbox) Add(_ context.Context, n domain.Notification) error {
	for _, e := range m.all {
		if n.ReservationID != 0 && e.UserID == n.UserID && e.Kind == n.Kind && e.ReservationID == n.ReservationID {
			return nil
		}
	}
	n.ID = int64(len(m.all) + 1)
	n.CreatedAt = time.Now()
	m.all = append(m.all, n)
	return nil
}
func (m *memInbox) forUser(u int64) []domain.Notification {
	var out []domain.Notification
	for _, n := range m.all {
		if n.UserID == u {
			out = append(out, n)
		}
	}
	return out
}
func (m *memInbox) kinds(u int64) []string {
	var out []string
	for _, n := range m.forUser(u) {
		out = append(out, n.Kind)
	}
	return out
}

// ClaimUnpushed models the SQL: a notification is handed out once per lease and each claim counts an attempt.
func (m *memInbox) ClaimUnpushed(_ context.Context, limit int, _ time.Duration) ([]domain.Notification, error) {
	var out []domain.Notification
	for _, n := range m.all {
		if len(out) == limit {
			break
		}
		if m.pushed[n.ID] || m.failures[n.ID] >= 5 || m.leased[n.ID] {
			continue
		}
		if m.failures == nil {
			m.failures = map[int64]int{}
		}
		m.failures[n.ID]++ // an attempt is counted when it is claimed
		m.leased[n.ID] = true
		out = append(out, n)
	}
	return out, nil
}
func (m *memInbox) MarkPushed(_ context.Context, id int64) error {
	if m.pushed == nil {
		m.pushed = map[int64]bool{}
	}
	m.pushed[id] = true
	return nil
}
func (m *memInbox) expireLeases() { m.leased = map[int64]bool{} }

func withInbox(svc *ReservationService) *memInbox {
	inbox := &memInbox{leased: map[int64]bool{}}
	svc.opts.Notifier = NewNotifier(inbox, nil)
	return inbox
}

func TestManagerIsToldOfANewReservation(t *testing.T) {
	svc, _, _ := rushSvc(ReservationOptions{})
	inbox := withInbox(svc)
	if _, err := svc.Reserve(context.Background(), guestP, rushCmd); err != nil {
		t.Fatal(err)
	}
	if got := inbox.kinds(50); len(got) != 1 || got[0] != "reservation.created" {
		t.Fatalf("manager: %v", got)
	}
	if n := inbox.forUser(50)[0]; !strings.Contains(n.Title, "#1") || n.HotelID != 1 || n.ReservationID != 1 {
		t.Fatalf("notification: %+v", n)
	}
	if len(inbox.forUser(7)) != 0 {
		t.Fatal("the guest asked for it; no need to tell them")
	}
}

func TestPaymentNotifiesBothSides(t *testing.T) {
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	svc, _, _ := newResSvc(pending(), w, nil, activeHotel())
	svc.now = time.Now
	inbox := withInbox(svc)
	if _, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodWallet}); err != nil {
		t.Fatal(err)
	}
	if got := inbox.kinds(7); len(got) != 1 || got[0] != "reservation.confirmed" {
		t.Fatalf("guest: %v", got)
	}
	if got := inbox.kinds(50); len(got) != 1 || got[0] != "reservation.paid" {
		t.Fatalf("manager: %v", got)
	}
	// Paying again (a retry) must not notify again.
	svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodWallet})
	if len(inbox.all) != 2 {
		t.Fatalf("a retry notified again: %d", len(inbox.all))
	}
}

func TestCancellationNotifiesTheOtherSideWithTheMoney(t *testing.T) {
	// The guest cancels in the 50% tier: the manager hears about it, the guest gets the refund breakdown.
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	svc, _, _ := newResSvc(paidAt(day(12)), w, nil, tieredHotel())
	inbox := withInbox(svc)
	if _, err := svc.Cancel(context.Background(), guestP, 5); err != nil {
		t.Fatal(err)
	}
	mgr := inbox.forUser(50)
	if len(mgr) != 1 || mgr[0].Kind != "reservation.cancelled" || !strings.Contains(mgr[0].Body, "2500001.0000") {
		t.Fatalf("manager: %+v", mgr)
	}
	guest := inbox.forUser(7)
	if len(guest) != 1 || guest[0].Kind != "reservation.refunded" || !strings.Contains(guest[0].Body, "2500000.0000 VND was paid back") {
		t.Fatalf("guest: %+v", guest)
	}

	// When the hotel cancels, it is the guest who is told.
	svc, _, _ = newResSvc(paidAt(day(12)), &fakeWallets{}, nil, tieredHotel())
	inbox = withInbox(svc)
	svc.Cancel(context.Background(), ownerP, 5)
	if got := inbox.kinds(7); len(got) != 1 || got[0] != "reservation.cancelled_by_hotel" || len(inbox.forUser(50)) != 0 {
		t.Fatalf("hotel-initiated: guest %v manager %v", got, inbox.kinds(50))
	}
}

func TestRejectionAndExpiryNotifyTheGuest(t *testing.T) {
	svc, _, _ := newResSvc(pending(), &fakeWallets{}, nil, activeHotel())
	inbox := withInbox(svc)
	if _, err := svc.Reject(context.Background(), ownerP, 5, "renovation"); err != nil {
		t.Fatal(err)
	}
	if n := inbox.forUser(7); len(n) != 1 || n[0].Kind != "reservation.rejected" || !strings.Contains(n[0].Body, "renovation") {
		t.Fatalf("rejection: %+v", n)
	}

	x := pending()
	past := time.Now().Add(-2 * time.Minute)
	x.ExpiresAt = &past
	svc, _, _ = newResSvc(x, &fakeWallets{}, nil, activeHotel())
	svc.now = time.Now
	inbox = withInbox(svc)
	svc.ReleaseExpired(context.Background())
	if got := inbox.kinds(7); len(got) != 1 || got[0] != "reservation.expired" {
		t.Fatalf("expiry: %v", got)
	}
}

func TestRemindersAreSentOncePerStay(t *testing.T) {
	x := paidAt(day(11)) // "today" in newResSvc is Jan 10: the stay starts tomorrow
	svc, repo, _ := newResSvc(x, &fakeWallets{}, nil, activeHotel())
	inbox := withInbox(svc)
	repo.due = []domain.Reservation{x}

	if n, err := svc.SendReminders(context.Background()); err != nil || n != 1 {
		t.Fatalf("first sweep: %d %v", n, err)
	}
	n := inbox.forUser(7)
	if len(n) != 1 || n[0].Kind != "reservation.reminder" || !strings.Contains(n[0].Title, "tomorrow") || !strings.Contains(n[0].Body, "14:00") {
		t.Fatalf("reminder: %+v", n)
	}
	// A second replica (or the next sweep) finds it already marked and sends nothing.
	if n, _ := svc.SendReminders(context.Background()); n != 0 || len(inbox.forUser(7)) != 1 {
		t.Fatalf("second sweep sent again: %d", n)
	}
	// No notifier configured: nothing happens, and nothing is marked.
	svc, repo, _ = newResSvc(x, &fakeWallets{}, nil, activeHotel())
	repo.due = []domain.Reservation{x}
	if n, _ := svc.SendReminders(context.Background()); n != 0 || len(repo.reminded) != 0 {
		t.Fatal("without a notifier there is nothing to send")
	}
}

type flakyPush struct {
	fails int
	got   []int64
}

func (f *flakyPush) Push(_ context.Context, n domain.Notification) error {
	if f.fails > 0 {
		f.fails--
		return domain.ErrUpstreamUnavailable
	}
	f.got = append(f.got, n.ID)
	return nil
}

func TestPushRetriesUntilNotificationServiceAccepts(t *testing.T) {
	inbox := &memInbox{leased: map[int64]bool{}}
	inbox.Add(context.Background(), domain.Notification{UserID: 7, Kind: "k", ReservationID: 1})
	push := &flakyPush{fails: 2}
	svc := NewNotificationService(inbox, push, nil)

	// Each round is a later sweep, after the lease of the previous claim ran out.
	for round, want := range []int{0, 0, 1, 0} {
		inbox.expireLeases()
		if n, err := svc.PushPending(context.Background()); err != nil || n != want {
			t.Fatalf("round %d: sent %d (want %d) err %v", round, n, want, err)
		}
	}
	if len(push.got) != 1 || inbox.failures[1] != 3 {
		t.Fatalf("delivered %v attempts %v", push.got, inbox.failures)
	}
	// It gives up after 5 attempts instead of hammering an endpoint that never accepts it.
	inbox2 := &memInbox{leased: map[int64]bool{}}
	inbox2.Add(context.Background(), domain.Notification{UserID: 7, Kind: "k", ReservationID: 1})
	svc = NewNotificationService(inbox2, &flakyPush{fails: 100}, nil)
	for i := 0; i < 8; i++ {
		inbox2.expireLeases()
		svc.PushPending(context.Background())
	}
	if inbox2.failures[1] != 5 {
		t.Fatalf("attempts: %d", inbox2.failures[1])
	}
	// With no notification-service configured the inbox is all there is.
	if n, err := NewNotificationService(inbox, nil, nil).PushPending(context.Background()); n != 0 || err != nil {
		t.Fatal("no gateway")
	}
	_ = errors.New
}

func TestTwoReplicasNeverPushTheSameNotification(t *testing.T) {
	inbox := &memInbox{leased: map[int64]bool{}}
	for i := 1; i <= 10; i++ {
		inbox.Add(context.Background(), domain.Notification{UserID: 7, Kind: "k", ReservationID: int64(i)})
	}
	pushA, pushB := &flakyPush{}, &flakyPush{}
	a, b := NewNotificationService(inbox, pushA, nil), NewNotificationService(inbox, pushB, nil)
	// Both sweep "at the same time": the second claim finds the first one's leases in place.
	if n, _ := a.PushPending(context.Background()); n != 10 {
		t.Fatalf("a: %d", n)
	}
	if n, _ := b.PushPending(context.Background()); n != 0 || len(pushB.got) != 0 {
		t.Fatalf("b pushed %v again", pushB.got)
	}
	if len(pushA.got) != 10 {
		t.Fatalf("a: %v", pushA.got)
	}
}
