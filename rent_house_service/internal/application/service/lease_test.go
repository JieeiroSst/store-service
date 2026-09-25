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

type fakeHomestayRepo struct {
	outbound.HomestayRepository
	h domain.Homestay
}

func (f fakeHomestayRepo) Get(context.Context, int64) (domain.Homestay, error) { return f.h, nil }

type memLeases struct {
	outbound.LeaseRepository
	l       domain.Lease
	created *outbound.CreateLeaseParams
}

func (m *memLeases) Create(_ context.Context, p outbound.CreateLeaseParams) (domain.Lease, error) {
	m.created = &p
	m.l = domain.Lease{ID: 9, UserID: p.UserID, Model: p.Model, Status: domain.LeasePending, Currency: p.Currency, Rent: p.Rent, ExpiresAt: &p.ExpiresAt}
	for _, s := range p.Schedule {
		m.l.Invoices = append(m.l.Invoices, domain.Invoice{ID: int64(100 + s.PeriodNo), LeaseID: 9, PeriodNo: s.PeriodNo, Amount: p.Rent, Status: domain.InvoiceUnpaid, DueDate: s.DueDate})
	}
	return m.l, nil
}
func (m *memLeases) Get(context.Context, int64) (domain.Lease, error) { return m.l, nil }
func (m *memLeases) MarkInvoicePaid(_ context.Context, id int64, method domain.PaymentMethod, ref string) (domain.Invoice, error) {
	for i := range m.l.Invoices {
		if m.l.Invoices[i].ID == id {
			m.l.Invoices[i].Status, m.l.Invoices[i].PaymentMethod, m.l.Invoices[i].PaymentRef = domain.InvoicePaid, method, ref
			m.l.Status = domain.LeaseActive
			return m.l.Invoices[i], nil
		}
	}
	return domain.Invoice{}, domain.ErrConflict
}

func homestayWithRates(rates ...domain.Rate) fakeHomestayRepo {
	return fakeHomestayRepo{h: domain.Homestay{ID: 1, Status: domain.HomestayActive, WalletID: "host-wallet", HostID: 50, Rates: rates}}
}

func newLeaseSvc(h fakeHomestayRepo, w *fakeWallets) (*LeaseService, *memLeases) {
	repo := &memLeases{}
	var wg outbound.WalletGateway
	if w != nil {
		wg = w
	}
	s := NewLeaseService(repo, h, NewPaymentRail(wg, nil, nil, nil), LeaseOptions{HoldTTL: time.Minute})
	s.now = func() time.Time { return time.Date(2027, 1, 10, 9, 0, 0, 0, time.UTC) }
	return s, repo
}

var monthRate = domain.Rate{HomestayID: 1, Model: domain.ModelMonth, Price: "5000000.00", Currency: "VND", MinPeriods: 3, Active: true}

func TestLeaseStart(t *testing.T) {
	start := time.Date(2027, 2, 15, 0, 0, 0, 0, time.UTC)
	svc, repo := newLeaseSvc(homestayWithRates(monthRate), nil)

	l, err := svc.Start(context.Background(), guestP, inbound.LeaseCommand{HomestayID: 1, Model: domain.ModelMonth, Start: start, Periods: 6, BillingDay: 5})
	if err != nil {
		t.Fatal(err)
	}
	p := repo.created
	if p.Rent != "5000000.00" || p.Currency != "VND" || len(p.Schedule) != 6 || !p.End.Equal(time.Date(2027, 8, 15, 0, 0, 0, 0, time.UTC)) || p.RequestID == "" {
		t.Fatalf("bad params: %+v", p)
	}
	// Rent is collected on the 5th from the second month on; the first invoice is due at signing.
	if !p.Schedule[0].DueDate.Equal(start) || !p.Schedule[1].DueDate.Equal(time.Date(2027, 3, 5, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("due dates: %v %v", p.Schedule[0].DueDate, p.Schedule[1].DueDate)
	}
	if l.Status != domain.LeasePending {
		t.Fatal("new lease must wait for its first payment")
	}
}

func TestLeaseStartRejects(t *testing.T) {
	start := time.Date(2027, 2, 15, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		who   inbound.Principal
		cmd   inbound.LeaseCommand
		homes fakeHomestayRepo
		want  error
	}{
		{"owner", inbound.Principal{UserID: 50}, inbound.LeaseCommand{Model: domain.ModelMonth, Start: start, Periods: 3}, homestayWithRates(monthRate), domain.ErrForbidden},
		{"day model", guestP, inbound.LeaseCommand{Model: domain.ModelDay, Start: start, Periods: 3}, homestayWithRates(monthRate), domain.ErrInvalid},
		{"model not offered", guestP, inbound.LeaseCommand{Model: domain.ModelYear, Start: start, Periods: 1}, homestayWithRates(monthRate), domain.ErrInvalid},
		{"below minimum", guestP, inbound.LeaseCommand{Model: domain.ModelMonth, Start: start, Periods: 2}, homestayWithRates(monthRate), domain.ErrInvalid},
		{"past start", guestP, inbound.LeaseCommand{Model: domain.ModelMonth, Start: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), Periods: 3}, homestayWithRates(monthRate), domain.ErrInvalid},
		{"bad billing day", guestP, inbound.LeaseCommand{Model: domain.ModelMonth, Start: start, Periods: 3, BillingDay: 31}, homestayWithRates(monthRate), domain.ErrInvalid},
		{"inactive rate", guestP, inbound.LeaseCommand{Model: domain.ModelMonth, Start: start, Periods: 3}, homestayWithRates(domain.Rate{Model: domain.ModelMonth, Price: "1", Currency: "VND", MinPeriods: 1}), domain.ErrInvalid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newLeaseSvc(tc.homes, nil)
			tc.cmd.HomestayID = 1
			if _, err := svc.Start(context.Background(), tc.who, tc.cmd); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if repo.created != nil {
				t.Fatal("nothing must be stored")
			}
		})
	}
}

func TestPayInvoiceInOrderAndActivates(t *testing.T) {
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	svc, repo := newLeaseSvc(homestayWithRates(monthRate), w)
	if _, err := svc.Start(context.Background(), guestP, inbound.LeaseCommand{HomestayID: 1, Model: domain.ModelMonth, Start: time.Date(2027, 2, 15, 0, 0, 0, 0, time.UTC), Periods: 3}); err != nil {
		t.Fatal(err)
	}
	pay := inbound.PayCommand{Method: domain.MethodWallet}

	// Period 2 cannot be paid before period 1.
	if _, err := svc.PayInvoice(context.Background(), guestP, 9, 1, pay); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("out of order: %v", err)
	}
	l, err := svc.PayInvoice(context.Background(), guestP, 9, 0, pay)
	if err != nil || l.Status != domain.LeaseActive || l.Invoices[0].Status != domain.InvoicePaid {
		t.Fatalf("first payment: %+v %v", l, err)
	}
	if _, ok := w.transfers["rent-house-invoice-100"]; !ok {
		t.Fatalf("wrong wallet reference: %v", w.transfers)
	}
	// Paying the same invoice again is a no-op, and the next one is now allowed.
	if _, err := svc.PayInvoice(context.Background(), guestP, 9, 0, pay); err != nil || len(w.transfers) != 1 {
		t.Fatalf("repeat: %v transfers=%d", err, len(w.transfers))
	}
	if l, err = svc.PayInvoice(context.Background(), guestP, 9, 1, pay); err != nil || l.Invoices[1].Status != domain.InvoicePaid {
		t.Fatalf("second period: %v", err)
	}
	// Someone else's lease is invisible.
	if _, err := svc.PayInvoice(context.Background(), inbound.Principal{UserID: 99}, 9, 2, pay); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("other tenant: %v", err)
	}
	_ = repo
}

func TestPayInvoiceRefundsWhenLeaseGone(t *testing.T) {
	w := &fakeWallets{wallet: domain.Wallet{ID: "w1", Currency: "VND"}}
	svc, repo := newLeaseSvc(homestayWithRates(monthRate), w)
	svc.Start(context.Background(), guestP, inbound.LeaseCommand{HomestayID: 1, Model: domain.ModelMonth, Start: time.Date(2027, 2, 15, 0, 0, 0, 0, time.UTC), Periods: 3})
	repo.l.Invoices[0].ID = 555 // MarkInvoicePaid will not find it: the invoice was voided meanwhile
	repo.l.Invoices[0].PeriodNo = 0

	repoGone := &memLeasesGone{memLeases: repo}
	svc.leases = repoGone
	_, err := svc.PayInvoice(context.Background(), guestP, 9, 0, inbound.PayCommand{Method: domain.MethodWallet})
	if !errors.Is(err, domain.ErrConflict) || len(w.reversed) != 1 {
		t.Fatalf("want conflict + refund, got %v reversed=%v", err, w.reversed)
	}
}

type memLeasesGone struct{ *memLeases }

func (memLeasesGone) MarkInvoicePaid(context.Context, int64, domain.PaymentMethod, string) (domain.Invoice, error) {
	return domain.Invoice{}, domain.ErrConflict
}

// terminatingLeases ends the lease the way the SQL does: unpaid future periods are voided,
// paid future periods are marked refunded, everything else is left alone.
type terminatingLeases struct {
	*memLeases
	from time.Time
	ends int
}

func (t *terminatingLeases) Terminate(_ context.Context, _ int64, from time.Time) (domain.Lease, error) {
	t.ends++
	t.l.Status = domain.LeaseEnded
	for i := range t.l.Invoices {
		inv := &t.l.Invoices[i]
		if inv.PeriodStart.Before(from) {
			continue
		}
		switch inv.Status {
		case domain.InvoiceUnpaid:
			inv.Status = domain.InvoiceVoid
		case domain.InvoicePaid:
			inv.Status = domain.InvoiceRefunded
		}
	}
	return t.l, nil
}

func activeLeaseWithPayments() *memLeases {
	d := func(m time.Month, day int) time.Time { return time.Date(2027, m, day, 0, 0, 0, 0, time.UTC) }
	return &memLeases{l: domain.Lease{ID: 9, UserID: 7, Status: domain.LeaseActive, Currency: "VND", Invoices: []domain.Invoice{
		{ID: 100, PeriodNo: 0, PeriodStart: d(1, 1), Status: domain.InvoicePaid, PaymentMethod: domain.MethodWallet, PaymentRef: "t-0"}, // in the past
		{ID: 101, PeriodNo: 1, PeriodStart: d(2, 1), Status: domain.InvoicePaid, PaymentMethod: domain.MethodWallet, PaymentRef: "t-1"}, // running (today = Feb 10)
		{ID: 102, PeriodNo: 2, PeriodStart: d(3, 1), Status: domain.InvoicePaid, PaymentMethod: domain.MethodWallet, PaymentRef: "t-2"}, // paid ahead: refund
		{ID: 103, PeriodNo: 3, PeriodStart: d(4, 1), Status: domain.InvoiceUnpaid},                                                      // owed no more
	}}}
}

func TestTerminateRefundsPeriodsNotYetStarted(t *testing.T) {
	w := &fakeWallets{}
	svc, _ := newLeaseSvc(homestayWithRates(monthRate), w)
	repo := &terminatingLeases{memLeases: activeLeaseWithPayments()}
	svc.leases = repo
	svc.now = func() time.Time { return time.Date(2027, 2, 10, 9, 0, 0, 0, time.UTC) }

	l, err := svc.Terminate(context.Background(), adminP, 9)
	if err != nil || l.Status != domain.LeaseEnded {
		t.Fatalf("terminate: %+v %v", l, err)
	}
	if len(w.reversed) != 1 || w.reversed[0] != "t-2" {
		t.Fatalf("only the paid period that had not begun is refunded, reversed = %v", w.reversed)
	}
	st := func(i int) domain.InvoiceStatus { return l.Invoices[i].Status }
	if st(0) != domain.InvoicePaid || st(1) != domain.InvoicePaid || st(2) != domain.InvoiceRefunded || st(3) != domain.InvoiceVoid {
		t.Fatalf("statuses: %v %v %v %v", st(0), st(1), st(2), st(3))
	}
}

func TestTerminateRetryOnlyRepeatsRefunds(t *testing.T) {
	w := &failingReverse{fakeWallets: &fakeWallets{}, failFirst: true}
	svc, _ := newLeaseSvc(homestayWithRates(monthRate), nil)
	svc.rail = NewPaymentRail(w, nil, nil, nil)
	repo := &terminatingLeases{memLeases: activeLeaseWithPayments()}
	svc.leases = repo
	svc.now = func() time.Time { return time.Date(2027, 2, 10, 9, 0, 0, 0, time.UTC) }

	if _, err := svc.Terminate(context.Background(), adminP, 9); err == nil {
		t.Fatal("a failed refund must be reported")
	}
	if repo.l.Status != domain.LeaseEnded {
		t.Fatal("the lease is already ended even though the refund failed")
	}
	if _, err := svc.Terminate(context.Background(), adminP, 9); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if repo.ends != 1 {
		t.Fatalf("retry must not end the lease again, ends = %d", repo.ends)
	}
	if len(w.reversed) != 1 || w.reversed[0] != "t-2" {
		t.Fatalf("reversed = %v", w.reversed)
	}
}

type failingReverse struct {
	*fakeWallets
	failFirst bool
}

func (f *failingReverse) ReverseTransfer(ctx context.Context, id, reason string) error {
	if f.failFirst {
		f.failFirst = false
		return domain.ErrUpstreamUnavailable
	}
	return f.fakeWallets.ReverseTransfer(ctx, id, reason)
}

func TestTerminateOnlyByAdminAndOnlyActive(t *testing.T) {
	svc, _ := newLeaseSvc(homestayWithRates(monthRate), &fakeWallets{})
	svc.leases = &terminatingLeases{memLeases: activeLeaseWithPayments()}
	if _, err := svc.Terminate(context.Background(), guestP, 9); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("guest: %v", err)
	}
	pending := activeLeaseWithPayments()
	pending.l.Status = domain.LeasePending
	svc.leases = &terminatingLeases{memLeases: pending}
	if _, err := svc.Terminate(context.Background(), adminP, 9); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("pending: %v", err)
	}
}

func TestOwnerTerminatesTenantCannotOthersSeeNothing(t *testing.T) {
	svc, _ := newLeaseSvc(homestayWithRates(monthRate), &fakeWallets{})
	svc.leases = &terminatingLeases{memLeases: activeLeaseWithPayments()}
	svc.now = func() time.Time { return time.Date(2027, 2, 10, 9, 0, 0, 0, time.UTC) }

	if _, err := svc.Terminate(context.Background(), guestP, 9); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("tenant ending their own lease: %v", err)
	}
	if _, err := svc.Terminate(context.Background(), otherP, 9); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("stranger: %v", err)
	}
	if _, err := svc.Terminate(context.Background(), ownerP, 9); err != nil {
		t.Fatalf("the homestay's owner: %v", err)
	}
}
