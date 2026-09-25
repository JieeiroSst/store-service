package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type memBookings struct {
	outbound.BookingRepository
	b        domain.Booking
	cancels  int
	failPaid error
}

func (m *memBookings) Get(context.Context, int64) (domain.Booking, error) { return m.b, nil }
func (m *memBookings) SetPaymentAttempt(_ context.Context, _ int64, method domain.PaymentMethod, ref string) (domain.Booking, error) {
	m.b.PaymentMethod, m.b.PaymentRef, m.b.Version = method, ref, m.b.Version+1
	return m.b, nil
}
func (m *memBookings) MarkPaid(_ context.Context, _ int64, method domain.PaymentMethod, ref string) (domain.Booking, error) {
	if m.failPaid != nil {
		err := m.failPaid
		m.failPaid = nil
		return domain.Booking{}, err
	}
	if m.b.Status != domain.BookingPending {
		if m.b.Status == domain.BookingConfirmed && m.b.PaymentRef == ref {
			return m.b, nil
		}
		return domain.Booking{}, domain.ErrConflict
	}
	m.b.Status, m.b.PaymentMethod, m.b.PaymentRef = domain.BookingConfirmed, method, ref
	return m.b, nil
}
func (m *memBookings) Cancel(context.Context, int64, int64) (domain.Booking, error) {
	m.cancels++
	m.b.Status = domain.BookingCancelled
	return m.b, nil
}
func (m *memBookings) Expire(context.Context, int64) (bool, error) {
	if m.b.Status != domain.BookingPending {
		return false, nil
	}
	m.b.Status = domain.BookingCancelled
	return true, nil
}
func (m *memBookings) ListExpired(context.Context, time.Time, int) ([]domain.Booking, error) {
	if m.b.Status == domain.BookingPending && m.b.ExpiresAt != nil && m.b.ExpiresAt.Before(time.Now()) {
		return []domain.Booking{m.b}, nil
	}
	return nil, nil
}

type fakeWallets struct {
	outbound.WalletGateway
	wallet      domain.Wallet
	noWallet    bool
	transfers   map[string]string // reference -> transfer id
	reversed    []string
	transferErr error
	missing     map[string]bool
	status      map[string]string
	ownerID     int64    // user the wallet returned by Get belongs to
	to          []string // destination of each transfer
}

func (f *fakeWallets) Get(_ context.Context, id string) (domain.Wallet, error) {
	if f.missing[id] {
		return domain.Wallet{}, domain.ErrNotFound
	}
	if st, ok := f.status[id]; ok {
		return domain.Wallet{ID: id, Status: st}, nil
	}
	return domain.Wallet{ID: id, Status: "ACTIVE", UserID: f.ownerID}, nil
}
func (f *fakeWallets) GetByUser(context.Context, int64) (domain.Wallet, error) {
	if f.noWallet {
		return domain.Wallet{}, domain.ErrNotFound
	}
	return f.wallet, nil
}
func (f *fakeWallets) Transfer(_ context.Context, p outbound.TransferParams) (string, error) {
	if f.transferErr != nil {
		return "", f.transferErr
	}
	f.to = append(f.to, p.ToWalletID)
	if f.transfers == nil {
		f.transfers = map[string]string{}
	}
	if id, ok := f.transfers[p.ReferenceID]; ok {
		return id, nil // idempotent
	}
	id := "t-" + p.ReferenceID
	f.transfers[p.ReferenceID] = id
	return id, nil
}
func (f *fakeWallets) ReverseTransfer(_ context.Context, id, _ string) error {
	f.reversed = append(f.reversed, id)
	return nil
}

type fakeGateway struct {
	outbound.PaymentGateway
	status   domain.GatewayStatus
	created  []outbound.CreateGatewayPayment
	refunded []int64
}

func (f *fakeGateway) Create(_ context.Context, in outbound.CreateGatewayPayment) (domain.GatewayPayment, error) {
	f.created = append(f.created, in)
	return domain.GatewayPayment{ID: 77, Status: f.status, Amount: in.Amount}, nil
}
func (f *fakeGateway) Get(context.Context, int64) (domain.GatewayPayment, error) {
	return domain.GatewayPayment{ID: 77, Status: f.status}, nil
}
func (f *fakeGateway) Refund(_ context.Context, id int64) error {
	f.refunded = append(f.refunded, id)
	return nil
}

var guestP = inbound.Principal{UserID: 7, Email: "g@x.com"}

func pendingBooking() domain.Booking {
	exp := time.Now().Add(10 * time.Minute)
	return domain.Booking{ID: 5, UserID: 7, HomestayID: 1, Status: domain.BookingPending, Currency: "VND",
		TotalAmount: "300001.500000", CheckIn: time.Now().AddDate(0, 0, 10), CheckOut: time.Now().AddDate(0, 0, 12), ExpiresAt: &exp}
}

func newPaySvc(b domain.Booking, w *fakeWallets, g *fakeGateway) (*BookingService, *memBookings) {
	repo := &memBookings{b: b}
	var wg outbound.WalletGateway
	var pg outbound.PaymentGateway
	if w != nil {
		wg = w
	}
	if g != nil {
		pg = g
	}
	rail := NewPaymentRail(wg, pg, nil, nil)
	return NewBookingService(repo, fakeHomestays{h: domain.Homestay{ID: 1, WalletID: "host-wallet", HostID: 50}}, rail, BookingOptions{HoldTTL: time.Minute}), repo
}

func TestPayWalletIsIdempotentAndRoundsAmount(t *testing.T) {
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "vnd"}}
	svc, repo := newPaySvc(pendingBooking(), w, nil)
	cmd := inbound.PayCommand{Method: domain.MethodWallet}

	b, err := svc.Pay(context.Background(), guestP, 5, cmd)
	if err != nil || b.Status != domain.BookingConfirmed || b.PaymentRef != "t-rent-house-booking-5" {
		t.Fatalf("first pay: %+v, %v", b, err)
	}
	if len(w.to) != 1 || w.to[0] != "host-wallet" {
		t.Fatalf("money must go to the homestay's own wallet, went to %v", w.to)
	}
	if _, err := svc.Pay(context.Background(), guestP, 5, cmd); err != nil || len(w.transfers) != 1 {
		t.Fatalf("second pay must not charge again: transfers=%d err=%v", len(w.transfers), err)
	}
	if repo.b.Status != domain.BookingConfirmed {
		t.Fatal("not confirmed")
	}
}

func TestPayWalletErrors(t *testing.T) {
	tests := []struct {
		name string
		w    *fakeWallets
		who  inbound.Principal
		want error
	}{
		{"no wallet", &fakeWallets{noWallet: true}, guestP, domain.ErrInvalid},
		{"currency mismatch", &fakeWallets{wallet: domain.Wallet{ID: "w", Currency: "USD"}}, guestP, domain.ErrInvalid},
		{"insufficient", &fakeWallets{wallet: domain.Wallet{ID: "w", Currency: "VND"}, transferErr: domain.ErrInsufficientFunds}, guestP, domain.ErrInsufficientFunds},
		{"someone else's booking", &fakeWallets{wallet: domain.Wallet{ID: "w", Currency: "VND"}}, inbound.Principal{UserID: 99}, domain.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newPaySvc(pendingBooking(), tc.w, nil)
			_, err := svc.Pay(context.Background(), tc.who, 5, inbound.PayCommand{Method: domain.MethodWallet})
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if repo.b.Status != domain.BookingPending {
				t.Fatal("booking must stay pending")
			}
		})
	}
}

func TestPayWalletRefundsWhenBookingNoLongerPayable(t *testing.T) {
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	svc, repo := newPaySvc(pendingBooking(), w, nil)
	repo.failPaid = domain.ErrConflict // hold expired while the transfer was in flight

	_, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodWallet})
	if !errors.Is(err, domain.ErrConflict) || len(w.reversed) != 1 {
		t.Fatalf("want conflict + refund, got err=%v reversed=%v", err, w.reversed)
	}
}

func TestPayWalletTransientErrorDoesNotRefund(t *testing.T) {
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	svc, repo := newPaySvc(pendingBooking(), w, nil)
	repo.failPaid = domain.ErrUpstreamUnavailable // e.g. database blip: money may be owed, a retry settles it

	if _, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodWallet}); err == nil {
		t.Fatal("want error")
	}
	if len(w.reversed) != 0 {
		t.Fatal("must not refund on a transient error")
	}
	if b, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodWallet}); err != nil || b.Status != domain.BookingConfirmed {
		t.Fatalf("retry should confirm: %+v %v", b, err)
	}
}

func TestPayGateway(t *testing.T) {
	g := &fakeGateway{status: domain.GatewayPending}
	svc, repo := newPaySvc(pendingBooking(), nil, g)

	b, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodGateway, Provider: "Stripe"})
	if err != nil || b.Status != domain.BookingPending || b.PaymentRef != "77" {
		t.Fatalf("pending payment: %+v %v", b, err)
	}
	in := g.created[0]
	if in.Provider != "stripe" || in.Amount != 300002 || in.PayerEmail != "g@x.com" {
		t.Fatalf("bad gateway request: %+v", in)
	}

	// Retrying while in flight must not create a second payment.
	if _, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodGateway, Provider: "stripe"}); err != nil || len(g.created) != 1 {
		t.Fatalf("duplicate charge: created=%d err=%v", len(g.created), err)
	}

	// Polling the booking confirms it once the provider captured the payment.
	g.status = domain.GatewayCaptured
	got, err := svc.Get(context.Background(), guestP, 5)
	if err != nil || got.Status != domain.BookingConfirmed || repo.b.Status != domain.BookingConfirmed {
		t.Fatalf("sync: %+v %v", got, err)
	}
}

func TestPayGatewayFailedThenRetryUsesNewKey(t *testing.T) {
	g := &fakeGateway{status: domain.GatewayFailed}
	svc, _ := newPaySvc(pendingBooking(), nil, g)
	cmd := inbound.PayCommand{Method: domain.MethodGateway, Provider: "stripe"}

	if _, err := svc.Pay(context.Background(), guestP, 5, cmd); !errors.Is(err, domain.ErrPaymentFailed) {
		t.Fatalf("err = %v", err)
	}
	if _, err := svc.Pay(context.Background(), guestP, 5, cmd); !errors.Is(err, domain.ErrPaymentFailed) {
		t.Fatalf("err = %v", err)
	}
	if len(g.created) != 2 || g.created[0].IdempotencyKey == g.created[1].IdempotencyKey {
		t.Fatalf("second attempt must use a new idempotency key: %+v", g.created)
	}
}

func TestPayDisabledMethodAndExpiredHold(t *testing.T) {
	svc, _ := newPaySvc(pendingBooking(), nil, nil)
	for _, m := range []domain.PaymentMethod{domain.MethodWallet, domain.MethodGateway, "cash"} {
		if _, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: m, Provider: "stripe"}); !errors.Is(err, domain.ErrInvalid) {
			t.Fatalf("%s: err = %v", m, err)
		}
	}
	b := pendingBooking()
	past := time.Now().Add(-time.Minute)
	b.ExpiresAt = &past
	svc, _ = newPaySvc(b, &fakeWallets{}, nil)
	if _, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodWallet}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expired: err = %v", err)
	}
}

func TestCancelRefunds(t *testing.T) {
	paid := pendingBooking()
	paid.Status, paid.PaymentMethod, paid.PaymentRef = domain.BookingConfirmed, domain.MethodWallet, "t-1"
	w := &fakeWallets{}
	svc, repo := newPaySvc(paid, w, nil)
	if _, err := svc.Cancel(context.Background(), guestP, 5); err != nil || len(w.reversed) != 1 || w.reversed[0] != "t-1" || repo.cancels != 1 {
		t.Fatalf("wallet refund: err=%v reversed=%v cancels=%d", err, w.reversed, repo.cancels)
	}

	paid.PaymentMethod, paid.PaymentRef = domain.MethodGateway, "77"
	g := &fakeGateway{}
	svc, _ = newPaySvc(paid, nil, g)
	if _, err := svc.Cancel(context.Background(), guestP, 5); err != nil || len(g.refunded) != 1 || g.refunded[0] != 77 {
		t.Fatalf("gateway refund: err=%v refunded=%v", err, g.refunded)
	}

	// Nothing paid, nothing to refund.
	svc, _ = newPaySvc(pendingBooking(), w, nil)
	before := len(w.reversed)
	if _, err := svc.Cancel(context.Background(), guestP, 5); err != nil || len(w.reversed) != before {
		t.Fatalf("unpaid cancel: err=%v", err)
	}
}

func TestReleaseExpired(t *testing.T) {
	b := pendingBooking()
	past := time.Now().Add(-2 * time.Minute)
	b.ExpiresAt = &past

	svc, repo := newPaySvc(b, &fakeWallets{}, nil)
	if n, err := svc.ReleaseExpired(context.Background()); err != nil || n != 1 || repo.b.Status != domain.BookingCancelled {
		t.Fatalf("plain expiry: n=%d err=%v status=%v", n, err, repo.b.Status)
	}

	// A gateway payment captured after the hold ended still confirms the booking.
	b.PaymentMethod, b.PaymentRef = domain.MethodGateway, "77"
	svc, repo = newPaySvc(b, nil, &fakeGateway{status: domain.GatewayCaptured})
	if n, _ := svc.ReleaseExpired(context.Background()); n != 0 || repo.b.Status != domain.BookingConfirmed {
		t.Fatalf("captured: n=%d status=%v", n, repo.b.Status)
	}

	// A payment still in flight gets one more hold period (1 minute in these tests).
	recent := time.Now().Add(-30 * time.Second)
	b.ExpiresAt = &recent
	svc, repo = newPaySvc(b, nil, &fakeGateway{status: domain.GatewayPending})
	if n, _ := svc.ReleaseExpired(context.Background()); n != 0 || repo.b.Status != domain.BookingPending {
		t.Fatalf("in flight: n=%d status=%v", n, repo.b.Status)
	}

	// Once that grace passes, the hold is released.
	longAgo := time.Now().Add(-10 * time.Minute)
	b.ExpiresAt = &longAgo
	svc, repo = newPaySvc(b, nil, &fakeGateway{status: domain.GatewayPending})
	if n, _ := svc.ReleaseExpired(context.Background()); n != 1 || repo.b.Status != domain.BookingCancelled {
		t.Fatalf("stale: n=%d status=%v", n, repo.b.Status)
	}
}

func TestWalletPaymentNeedsHomestayWallet(t *testing.T) {
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	repo := &memBookings{b: pendingBooking()}
	svc := NewBookingService(repo, fakeHomestays{h: domain.Homestay{ID: 1}}, NewPaymentRail(w, nil, nil, nil), BookingOptions{})
	_, err := svc.Pay(context.Background(), guestP, 5, inbound.PayCommand{Method: domain.MethodWallet})
	if !errors.Is(err, domain.ErrInvalid) || len(w.to) != 0 || repo.b.Status != domain.BookingPending {
		t.Fatalf("a homestay without a wallet must refuse wallet payment, err=%v transfers=%v", err, w.to)
	}
}

func TestHostSeesAndCancelsBookingsAtTheirHomestay(t *testing.T) {
	started := pendingBooking()
	started.CheckIn = time.Now().AddDate(0, 0, -1) // the stay has begun
	svc, repo := newPaySvc(started, &fakeWallets{}, nil)

	if _, err := svc.Get(context.Background(), otherP, 5); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger must not see the booking: %v", err)
	}
	if _, err := svc.Get(context.Background(), ownerP, 5); err != nil {
		t.Fatalf("the owner of the homestay sees it: %v", err)
	}
	// The guest is locked out once the stay started; the owner can still cancel.
	if _, err := svc.Cancel(context.Background(), guestP, 5); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("guest after start: %v", err)
	}
	if _, err := svc.Cancel(context.Background(), ownerP, 5); err != nil || repo.cancels != 1 {
		t.Fatalf("owner after start: %v cancels=%d", err, repo.cancels)
	}
}
