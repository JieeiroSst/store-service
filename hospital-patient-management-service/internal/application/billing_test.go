package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type fakePatients struct{ port.PatientRepository }

func (fakePatients) Get(_ context.Context, id int32) (*model.Patient, error) {
	uid := "user-7"
	return &model.Patient{ID: id, FirstName: "Binh", LastName: "Tran", Email: "b@x.vn", UserID: &uid}, nil
}

type fakeAccounts struct {
	rows  map[int32]model.BillingAccount
	locks int
}

func (f *fakeAccounts) Get(_ context.Context, id int32) (*model.BillingAccount, error) {
	a, ok := f.rows[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return &a, nil
}

func (f *fakeAccounts) Save(_ context.Context, a *model.BillingAccount) error {
	f.rows[a.PatientID] = *a
	return nil
}

func (f *fakeAccounts) WithLock(ctx context.Context, _ int32, fn func(context.Context) error) error {
	f.locks++
	return fn(ctx)
}

type payment struct {
	invoice int32
	amount  float64
	method  string
}

type fakeInvoices struct {
	off      bool
	fail     error
	accounts int
	nextID   int32
	created  map[int32]port.InvoiceState
	payments []payment
}

func (f *fakeInvoices) Enabled() bool { return !f.off }

func (f *fakeInvoices) EnsureAccount(_ context.Context, p *model.Patient, existing *model.BillingAccount) (*model.BillingAccount, error) {
	if f.fail != nil {
		return nil, f.fail
	}
	if existing != nil {
		return existing, nil
	}
	f.accounts++
	return &model.BillingAccount{PatientID: p.ID, CustomerID: 100 + p.ID, SubscriptionID: 200 + p.ID}, nil
}

func (f *fakeInvoices) CreateInvoice(_ context.Context, _ *model.BillingAccount, st port.InvoiceState) (int32, error) {
	if f.fail != nil {
		return 0, f.fail
	}
	f.nextID++
	f.created[f.nextID] = st
	return f.nextID, nil
}

func (f *fakeInvoices) SyncInvoice(_ context.Context, id int32, st port.InvoiceState) error {
	f.created[id] = st
	return nil
}

func (f *fakeInvoices) RecordPayment(_ context.Context, id int32, amount float64, method string, _ time.Time) error {
	f.payments = append(f.payments, payment{id, amount, method})
	return nil
}

type fakeWallet struct {
	off        bool
	fail       error
	refundFail error
	charges    []port.ChargeInput
	refunds    []string
}

func (f *fakeWallet) Enabled() bool { return !f.off }

func (f *fakeWallet) Charge(_ context.Context, in port.ChargeInput) (string, error) {
	if f.fail != nil {
		return "", f.fail
	}
	f.charges = append(f.charges, in)
	return "tr-" + in.Reference, nil
}

func (f *fakeWallet) Refund(_ context.Context, id, _ string) error {
	if f.refundFail != nil {
		return f.refundFail
	}
	f.refunds = append(f.refunds, id)
	return nil
}

type fakeNotifier struct{ sent []port.Notification }

func (f *fakeNotifier) Notify(_ context.Context, n port.Notification) error {
	f.sent = append(f.sent, n)
	return nil
}

type rig struct {
	svc      port.BillingUsecase
	accounts *fakeAccounts
	repo     *memBilling
	invoices *fakeInvoices
	wallet   *fakeWallet
	notes    *fakeNotifier
}

func newRig() *rig {
	r := &rig{
		repo:     &memBilling{},
		invoices: &fakeInvoices{created: map[int32]port.InvoiceState{}},
		wallet:   &fakeWallet{},
		notes:    &fakeNotifier{},
	}
	patients := fakePatients{}
	r.accounts = &fakeAccounts{rows: map[int32]model.BillingAccount{}}
	r.svc = NewBillingService(BillingParams{
		Repo: r.repo, Patients: patients,
		Accounts: r.accounts,
		Invoices: r.invoices, Wallet: r.wallet,
		Notify: NewPatientNotifier(patients, r.notes),
		Clock:  fixedClock(now),
	})
	return r
}

func newBill(amount, insurance float64) *model.Billing {
	return &model.Billing{PatientID: 1, Amount: amount, InsuranceCovered: insurance}
}

func TestCreateIssuesInvoiceForWhatThePatientOwes(t *testing.T) {
	ctx := context.Background()
	r := newRig()

	b, err := r.svc.Create(ctx, newBill(100, 40))
	if err != nil {
		t.Fatal(err)
	}
	if b.InvoiceID == nil {
		t.Fatal("bill has no invoice id")
	}
	if st := r.invoices.created[*b.InvoiceID]; st.Amount != 60 || st.Status != port.InvoiceIssued {
		t.Errorf("invoice = %+v, want amount 60 issued", st)
	}
	if len(r.notes.sent) != 1 || r.notes.sent[0].Email != "b@x.vn" {
		t.Errorf("notifications = %+v", r.notes.sent)
	}

	// The second bill reuses the patient's billing account.
	if _, err := r.svc.Create(ctx, newBill(10, 0)); err != nil {
		t.Fatal(err)
	}
	if r.invoices.accounts != 1 {
		t.Errorf("customer created %d times, want once", r.invoices.accounts)
	}
}

func TestFullyInsuredBillIsNotInvoiced(t *testing.T) {
	r := newRig()
	b, err := r.svc.Create(context.Background(), newBill(100, 100))
	if err != nil || b.InvoiceID != nil || len(r.invoices.created) != 0 || len(r.notes.sent) != 0 {
		t.Errorf("bill=%+v err=%v invoices=%v notes=%v", b, err, r.invoices.created, r.notes.sent)
	}
}

func TestPayingChargesTheWalletOnce(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	b, _ := r.svc.Create(ctx, newBill(100, 40))

	paid, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.wallet.charges) != 1 || r.wallet.charges[0].Amount != 60 || r.wallet.charges[0].PatientID != 1 {
		t.Fatalf("charges = %+v", r.wallet.charges)
	}
	if paid.WalletTransferID == nil || paid.Status != model.BillingPaid {
		t.Errorf("paid = %+v", paid)
	}
	if got := r.invoices.payments; len(got) != 1 || got[0].method != methodWallet || got[0].amount != 60 {
		t.Errorf("payments = %+v", got)
	}
	if st := r.invoices.created[*b.InvoiceID]; st.Status != port.InvoicePaid {
		t.Errorf("invoice status = %s", st.Status)
	}

	// The same request again must not move money again.
	if _, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil); err != nil || len(r.wallet.charges) != 1 {
		t.Errorf("repeat: err=%v charges=%d", err, len(r.wallet.charges))
	}
}

func TestPaidDateMeansPaidOffline(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	b, _ := r.svc.Create(ctx, newBill(100, 0))

	day := now.AddDate(0, 0, -2)
	paid, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, &day)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.wallet.charges) != 0 || paid.WalletTransferID != nil {
		t.Errorf("offline payment must not touch the wallet: %+v", r.wallet.charges)
	}
	if !paid.PaidDate.Equal(dateOnly(day)) || r.invoices.payments[0].method != methodOffline {
		t.Errorf("paid=%+v payments=%+v", paid, r.invoices.payments)
	}
}

func TestFailedChargeLeavesTheBillUnpaid(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	b, _ := r.svc.Create(ctx, newBill(100, 0))
	r.wallet.fail = model.PaymentFailed("insufficient wallet balance")

	if _, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil); !errors.Is(err, model.ErrPaymentFailed) {
		t.Fatalf("err = %v, want ErrPaymentFailed", err)
	}
	if got, _ := r.svc.Get(ctx, b.ID); got.Status != model.BillingPending || len(r.invoices.payments) != 0 {
		t.Errorf("bill = %+v payments=%v", got, r.invoices.payments)
	}
}

func TestReopeningAPaidBillRefundsTheWallet(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	b, _ := r.svc.Create(ctx, newBill(100, 0))
	paid, _ := r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil)

	got, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingCancelled, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.wallet.refunds) != 1 || r.wallet.refunds[0] != *paid.WalletTransferID {
		t.Errorf("refunds = %v", r.wallet.refunds)
	}
	if got.WalletTransferID != nil || got.PaidDate != nil || got.Status != model.BillingCancelled {
		t.Errorf("bill = %+v", got)
	}
	if st := r.invoices.created[*b.InvoiceID]; st.Status != port.InvoiceCancelled {
		t.Errorf("invoice status = %s", st.Status)
	}
}

func TestPaidBillsAreFrozen(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	b, _ := r.svc.Create(ctx, newBill(100, 0))
	_, _ = r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil)

	edit := *b
	edit.Amount, edit.PatientResponsibility = 500, 500
	if _, err := r.svc.Update(ctx, &edit); !errors.Is(err, model.ErrConflict) {
		t.Errorf("edit paid: got %v, want ErrConflict", err)
	}
	if err := r.svc.Delete(ctx, b.ID); !errors.Is(err, model.ErrConflict) {
		t.Errorf("delete paid: got %v, want ErrConflict", err)
	}
}

func TestEditingKeepsLinksAndResyncsTheInvoice(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	b, _ := r.svc.Create(ctx, newBill(100, 0))

	edit := model.Billing{ID: b.ID, PatientID: 1, Amount: 80}
	got, err := r.svc.Update(ctx, &edit)
	if err != nil {
		t.Fatal(err)
	}
	if got.InvoiceID == nil || *got.InvoiceID != *b.InvoiceID {
		t.Errorf("invoice link lost: %+v", got)
	}
	if st := r.invoices.created[*b.InvoiceID]; st.Amount != 80 {
		t.Errorf("invoice amount = %v, want 80", st.Amount)
	}
	edit2 := model.Billing{ID: b.ID, PatientID: 2, Amount: 80}
	if _, err := r.svc.Update(ctx, &edit2); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("moving an invoiced bill to another patient: got %v, want ErrInvalid", err)
	}
}

func TestDeletingCancelsTheInvoice(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	b, _ := r.svc.Create(ctx, newBill(100, 0))
	if err := r.svc.Delete(ctx, b.ID); err != nil {
		t.Fatal(err)
	}
	if st := r.invoices.created[*b.InvoiceID]; st.Status != port.InvoiceCancelled {
		t.Errorf("invoice status = %s", st.Status)
	}
}

func TestBillingServiceDownDoesNotBlockBillsAndIsRetried(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	r.invoices.fail = model.ErrUpstream

	b, err := r.svc.Create(ctx, newBill(100, 0))
	if err != nil || b.InvoiceID != nil {
		t.Fatalf("bill=%+v err=%v", b, err)
	}

	r.invoices.fail = nil // billing-service recovers
	got, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingOverdue, nil)
	if err != nil || got.InvoiceID == nil {
		t.Fatalf("retry: bill=%+v err=%v", got, err)
	}
	if st := r.invoices.created[*got.InvoiceID]; st.Status != port.InvoiceOverdue {
		t.Errorf("invoice status = %s", st.Status)
	}
}

func TestIntegrationsCanBeSwitchedOff(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	r.invoices.off, r.wallet.off = true, true

	b, err := r.svc.Create(ctx, newBill(100, 0))
	if err != nil || b.InvoiceID != nil {
		t.Fatalf("bill=%+v err=%v", b, err)
	}
	paid, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil)
	if err != nil || paid.Status != model.BillingPaid || len(r.wallet.charges) != 0 || len(r.invoices.created) != 0 {
		t.Errorf("paid=%+v err=%v charges=%v invoices=%v", paid, err, r.wallet.charges, r.invoices.created)
	}
}

func TestAccountCreationHappensUnderALock(t *testing.T) {
	r := newRig()
	if _, err := r.svc.Create(context.Background(), newBill(100, 0)); err != nil {
		t.Fatal(err)
	}
	if r.accounts.locks != 1 {
		t.Errorf("lock taken %d times, want 1", r.accounts.locks)
	}
}

func TestChargeUsesTheLinkedUserAccount(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	b, _ := r.svc.Create(ctx, newBill(100, 0))
	if _, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil); err != nil {
		t.Fatal(err)
	}
	if got := r.wallet.charges[0].UserID; got != "user-7" {
		t.Errorf("charged user %q, want user-7", got)
	}
}

// A refund that fails after the status was saved must not strand the bill:
// the wallet transfer stays recorded and the next call finishes the refund.
func TestInterruptedRefundIsFinishedOnRetry(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	b, _ := r.svc.Create(ctx, newBill(100, 0))
	paid, _ := r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil)

	r.wallet.refundFail = model.ErrUpstream
	if _, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingCancelled, nil); !errors.Is(err, model.ErrUpstream) {
		t.Fatalf("err = %v, want ErrUpstream", err)
	}
	stuck, _ := r.svc.Get(ctx, b.ID)
	if stuck.Status != model.BillingCancelled || stuck.WalletTransferID == nil {
		t.Fatalf("want Cancelled with the refund still owed, got %+v", stuck)
	}

	r.wallet.refundFail = nil // wallet-service recovers; the same request retried
	got, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingCancelled, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.wallet.refunds) != 1 || r.wallet.refunds[0] != *paid.WalletTransferID || got.WalletTransferID != nil {
		t.Errorf("refunds=%v bill=%+v", r.wallet.refunds, got)
	}
}

func TestOwedRefundIsSettledBeforePayingAgain(t *testing.T) {
	ctx := context.Background()
	r := newRig()
	b, _ := r.svc.Create(ctx, newBill(100, 0))
	_, _ = r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil)
	r.wallet.refundFail = model.ErrUpstream
	_, _ = r.svc.UpdateStatus(ctx, b.ID, model.BillingCancelled, nil)

	r.wallet.refundFail = nil
	got, err := r.svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.wallet.refunds) != 1 || len(r.wallet.charges) != 2 || got.WalletTransferID == nil {
		t.Errorf("refunds=%v charges=%d bill=%+v", r.wallet.refunds, len(r.wallet.charges), got)
	}
	if r.wallet.charges[0].Reference == r.wallet.charges[1].Reference {
		t.Error("a payment after a refund must use a fresh wallet reference")
	}
}
