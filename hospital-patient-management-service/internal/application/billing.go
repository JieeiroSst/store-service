package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

const (
	methodWallet  = "wallet"
	methodOffline = "offline"
)

type BillingParams struct {
	fx.In

	Repo     port.BillingRepository
	Patients port.PatientRepository
	Accounts port.BillingAccountRepository
	Invoices port.InvoiceGateway
	Wallet   port.WalletGateway
	Notify   *PatientNotifier
	Clock    port.Clock
}

type BillingService struct {
	repo     port.BillingRepository
	patients port.PatientRepository
	accounts port.BillingAccountRepository
	invoices port.InvoiceGateway
	wallet   port.WalletGateway
	notify   *PatientNotifier
	clock    port.Clock
}

func NewBillingService(p BillingParams) port.BillingUsecase {
	return &BillingService{
		repo: p.Repo, patients: p.Patients, accounts: p.Accounts,
		invoices: p.Invoices, wallet: p.Wallet, notify: p.Notify, clock: p.Clock,
	}
}

func validateBilling(b *model.Billing) error {
	if b.AppointmentID != nil && *b.AppointmentID <= 0 {
		b.AppointmentID = nil
	}
	if b.Status == "" {
		b.Status = model.BillingPending
	}
	if !b.Status.Valid() {
		return model.Invalid("status is not valid")
	}
	if err := positive("patient_id", b.PatientID); err != nil {
		return err
	}
	b.Amount = money(b.Amount)
	b.InsuranceCovered = money(b.InsuranceCovered)
	b.PatientResponsibility = money(b.PatientResponsibility)
	if b.Amount <= 0 {
		return model.Invalid("amount must be greater than zero")
	}
	if b.InsuranceCovered < 0 || b.PatientResponsibility < 0 {
		return model.Invalid("insurance_covered and patient_responsibility cannot be negative")
	}
	if b.InsuranceCovered > b.Amount {
		return model.Invalid("insurance_covered exceeds amount")
	}
	// The patient owes whatever insurance does not cover.
	if b.PatientResponsibility == 0 {
		b.PatientResponsibility = money(b.Amount - b.InsuranceCovered)
	}
	if math.Abs(b.InsuranceCovered+b.PatientResponsibility-b.Amount) > 0.005 {
		return model.Invalid("insurance_covered + patient_responsibility must equal amount")
	}
	return nil
}

func (s *BillingService) Get(ctx context.Context, id int32) (*model.Billing, error) {
	if id <= 0 {
		return nil, model.Invalid("id is required")
	}
	return s.repo.Get(ctx, id)
}

func (s *BillingService) Create(ctx context.Context, b *model.Billing) (*model.Billing, error) {
	if err := validateBilling(b); err != nil {
		return nil, err
	}
	b.PaidDate, b.InvoiceID, b.WalletTransferID = nil, nil, nil
	if b.Status == model.BillingPaid {
		now := dateOnly(s.clock.Now())
		b.PaidDate = &now
	}
	out, err := create[model.Billing](ctx, s.repo, b)
	if err != nil {
		return nil, err
	}

	s.issueInvoice(ctx, out)
	switch out.Status {
	case model.BillingPaid:
		s.recordPayment(ctx, out, methodOffline)
	case model.BillingPending:
		if out.PatientResponsibility > 0 {
			s.notify.Notify(ctx, out.PatientID, "New bill",
				fmt.Sprintf("A bill of %.2f was issued. You owe %.2f.", out.Amount, out.PatientResponsibility))
		}
	}
	return out, nil
}

func (s *BillingService) Update(ctx context.Context, b *model.Billing) (*model.Billing, error) {
	current, err := s.Get(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	if current.Status == model.BillingPaid {
		return nil, model.Conflict("a paid bill cannot be edited; reopen it first")
	}
	b.Status, b.PaidDate = current.Status, current.PaidDate
	b.InvoiceID, b.WalletTransferID = current.InvoiceID, current.WalletTransferID
	if err := validateBilling(b); err != nil {
		return nil, err
	}
	if current.InvoiceID != nil && b.PatientID != current.PatientID {
		return nil, model.Invalid("patient_id cannot change once the bill is invoiced")
	}
	out, err := update[model.Billing](ctx, s.repo, b.ID, b)
	if err != nil {
		return nil, err
	}
	if out.InvoiceID == nil {
		s.issueInvoice(ctx, out)
	} else {
		s.syncInvoice(ctx, out)
	}
	return out, nil
}

func (s *BillingService) UpdateStatus(ctx context.Context, id int32, status model.BillingStatus, paidDate *time.Time) (*model.Billing, error) {
	b, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !status.Valid() {
		return nil, model.Invalid("status is not valid")
	}
	if b.Status != model.BillingPaid && b.WalletTransferID != nil {
		if err := s.settleRefund(ctx, b); err != nil {
			return nil, err
		}
		if b, err = s.Get(ctx, id); err != nil {
			return nil, err
		}
	}
	prev := b.Status
	switch {
	case prev == status:
		return b, nil
	case status == model.BillingPaid:
		return s.markPaid(ctx, b, paidDate)
	case prev == model.BillingPaid:
		return s.reopen(ctx, b, status)
	}

	b.Status = status
	out, err := update[model.Billing](ctx, s.repo, id, b)
	if err != nil {
		return nil, err
	}
	if out.InvoiceID == nil {
		s.issueInvoice(ctx, out)
	} else {
		s.syncInvoice(ctx, out)
	}
	return out, nil
}

func (s *BillingService) markPaid(ctx context.Context, b *model.Billing, paidDate *time.Time) (*model.Billing, error) {
	method := methodOffline
	if paidDate == nil && b.PatientResponsibility > 0 && s.walletEnabled() {
		patient, err := s.patients.Get(ctx, b.PatientID)
		if err != nil {
			return nil, err
		}
		transferID, err := s.wallet.Charge(ctx, port.ChargeInput{
			PatientID:   b.PatientID,
			UserID:      deref(patient.UserID),
			Amount:      b.PatientResponsibility,
			Reference:   fmt.Sprintf("hospital-bill-%d-%d", b.ID, b.UpdatedAt.UnixNano()),
			Description: fmt.Sprintf("Hospital bill #%d", b.ID),
		})
		if err != nil {
			return nil, err
		}
		b.WalletTransferID = &transferID
		method = methodWallet
	}

	day := dateOnly(s.clock.Now())
	if paidDate != nil {
		day = dateOnly(*paidDate)
	}
	b.Status, b.PaidDate = model.BillingPaid, &day
	out, err := update[model.Billing](ctx, s.repo, b.ID, b)
	if err != nil {
		return nil, err
	}

	if out.InvoiceID == nil {
		s.issueInvoice(ctx, out)
	} else {
		s.syncInvoice(ctx, out)
	}
	s.recordPayment(ctx, out, method)
	s.notify.Notify(ctx, out.PatientID, "Payment received",
		fmt.Sprintf("We received your payment of %.2f for bill #%d. Thank you.", out.PatientResponsibility, out.ID))
	return out, nil
}

func (s *BillingService) reopen(ctx context.Context, b *model.Billing, status model.BillingStatus) (*model.Billing, error) {
	b.Status, b.PaidDate = status, nil
	out, err := update[model.Billing](ctx, s.repo, b.ID, b)
	if err != nil {
		return nil, err
	}
	if out.WalletTransferID != nil {
		if err := s.settleRefund(ctx, out); err != nil {
			return nil, err
		}
	}
	s.syncInvoice(ctx, out)
	return out, nil
}

func (s *BillingService) settleRefund(ctx context.Context, b *model.Billing) error {
	if s.wallet == nil {
		return fmt.Errorf("%w: wallet is not configured, cannot refund transfer %s", model.ErrUpstream, *b.WalletTransferID)
	}
	reason := fmt.Sprintf("hospital bill #%d changed from Paid to %s", b.ID, b.Status)
	if err := s.wallet.Refund(ctx, *b.WalletTransferID, reason); err != nil {
		return err
	}
	b.WalletTransferID = nil
	return s.repo.Update(ctx, b)
}

func (s *BillingService) Delete(ctx context.Context, id int32) error {
	b, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if b.Status == model.BillingPaid {
		return model.Conflict("a paid bill cannot be deleted; reopen or cancel it first")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if b.InvoiceID != nil {
		b.Status = model.BillingCancelled
		s.syncInvoice(ctx, b)
	}
	return nil
}

func (s *BillingService) List(ctx context.Context, patientID int32, status model.BillingStatus, page model.Page) ([]model.Billing, int64, error) {
	if err := positive("patient_id", patientID); err != nil {
		return nil, 0, err
	}
	if status != "" && !status.Valid() {
		return nil, 0, model.Invalid("status is not valid")
	}
	return s.repo.List(ctx, patientID, status, page)
}

func (s *BillingService) Stats(ctx context.Context, from, to *time.Time) (*model.BillingStats, error) {
	if from != nil && to != nil && to.Before(*from) {
		return nil, model.Invalid("date_to is before date_from")
	}
	return s.repo.Stats(ctx, from, to)
}

// ---- integrations (best-effort) ----

func (s *BillingService) walletEnabled() bool { return s.wallet != nil && s.wallet.Enabled() }

func (s *BillingService) invoiceEnabled() bool { return s.invoices != nil && s.invoices.Enabled() }

func invoiceState(b *model.Billing) port.InvoiceState {
	st := port.InvoiceIssued
	switch b.Status {
	case model.BillingPaid:
		st = port.InvoicePaid
	case model.BillingOverdue:
		st = port.InvoiceOverdue
	case model.BillingCancelled:
		st = port.InvoiceCancelled
	}
	return port.InvoiceState{Amount: b.PatientResponsibility, DueDate: b.DueDate, Status: st}
}

// issueInvoice creates the bill's invoice unless there is nothing to invoice
// or it already has one.
func (s *BillingService) issueInvoice(ctx context.Context, b *model.Billing) {
	if !s.invoiceEnabled() || b.InvoiceID != nil || b.PatientResponsibility <= 0 {
		return
	}
	if err := s.tryIssue(ctx, b); err != nil {
		logrus.WithError(err).WithField("bill_id", b.ID).Warn("invoice not issued; will retry on the next change to this bill")
	}
}

func (s *BillingService) tryIssue(ctx context.Context, b *model.Billing) error {
	patient, err := s.patients.Get(ctx, b.PatientID)
	if err != nil {
		return err
	}

	var acct *model.BillingAccount
	err = s.accounts.WithLock(ctx, b.PatientID, func(ctx context.Context) error {
		existing, err := s.accounts.Get(ctx, b.PatientID)
		if errors.Is(err, model.ErrNotFound) {
			existing, err = nil, nil
		}
		if err != nil {
			return err
		}
		acct, err = s.invoices.EnsureAccount(ctx, patient, existing)
		if err != nil {
			return err
		}
		if existing == nil {
			return s.accounts.Save(ctx, acct)
		}
		return nil
	})
	if err != nil {
		return err
	}

	id, err := s.invoices.CreateInvoice(ctx, acct, invoiceState(b))
	if err != nil {
		return err
	}
	b.InvoiceID = &id
	return s.repo.Update(ctx, b)
}

func (s *BillingService) syncInvoice(ctx context.Context, b *model.Billing) {
	if !s.invoiceEnabled() || b.InvoiceID == nil {
		return
	}
	if err := s.invoices.SyncInvoice(ctx, *b.InvoiceID, invoiceState(b)); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"bill_id": b.ID, "invoice_id": *b.InvoiceID}).Warn("invoice not synced")
	}
}

func (s *BillingService) recordPayment(ctx context.Context, b *model.Billing, method string) {
	if !s.invoiceEnabled() || b.InvoiceID == nil {
		return
	}
	at := s.clock.Now()
	if b.PaidDate != nil {
		at = *b.PaidDate
	}
	if err := s.invoices.RecordPayment(ctx, *b.InvoiceID, b.PatientResponsibility, method, at); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"bill_id": b.ID, "invoice_id": *b.InvoiceID}).Warn("payment not recorded in billing-service")
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
