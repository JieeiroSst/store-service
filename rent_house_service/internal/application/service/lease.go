package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type LeaseOptions struct {
	HoldTTL time.Duration
	Log     *slog.Logger
}

type LeaseService struct {
	leases    outbound.LeaseRepository
	homestays outbound.HomestayRepository
	rail      *PaymentRail
	opts      LeaseOptions
	now       func() time.Time
}

var _ inbound.LeaseUseCase = (*LeaseService)(nil)

func NewLeaseService(l outbound.LeaseRepository, h outbound.HomestayRepository, rail *PaymentRail, opts LeaseOptions) *LeaseService {
	if opts.Log == nil {
		opts.Log = slog.Default()
	}
	if opts.HoldTTL <= 0 {
		opts.HoldTTL = 15 * time.Minute
	}
	return &LeaseService{leases: l, homestays: h, rail: rail, opts: opts, now: time.Now}
}

func (s *LeaseService) today() time.Time { return s.now().UTC().Truncate(24 * time.Hour) }

func (s *LeaseService) Start(ctx context.Context, actor inbound.Principal, c inbound.LeaseCommand) (domain.Lease, error) {
	if !c.Model.IsLease() {
		return domain.Lease{}, fmt.Errorf("%w: model must be week, month or year (use bookings for nightly stays)", domain.ErrInvalid)
	}
	if c.Start.Before(s.today()) {
		return domain.Lease{}, fmt.Errorf("%w: start date is in the past", domain.ErrInvalid)
	}
	h, err := s.homestays.Get(ctx, c.HomestayID)
	if err != nil {
		return domain.Lease{}, err
	}
	if h.Status != domain.HomestayActive {
		return domain.Lease{}, domain.ErrNotFound
	}
	if h.HostID != 0 && h.HostID == actor.UserID {
		return domain.Lease{}, fmt.Errorf("%w: you cannot rent your own homestay", domain.ErrForbidden)
	}
	var rate *domain.Rate
	for i := range h.Rates {
		if h.Rates[i].Model == c.Model && h.Rates[i].Active {
			rate = &h.Rates[i]
		}
	}
	if rate == nil {
		return domain.Lease{}, fmt.Errorf("%w: this homestay is not offered by the %s", domain.ErrInvalid, c.Model)
	}
	if c.Periods < rate.MinPeriods {
		return domain.Lease{}, fmt.Errorf("%w: minimum rental is %d %s(s)", domain.ErrInvalid, rate.MinPeriods, c.Model)
	}
	billingDay := c.BillingDay
	if billingDay == 0 {
		billingDay = domain.DefaultBillingDay(c.Model, c.Start)
	}
	plans, end, err := domain.BuildSchedule(c.Model, c.Start, c.Periods, billingDay)
	if err != nil {
		return domain.Lease{}, err
	}
	reqID := strings.TrimSpace(c.RequestID)
	if reqID == "" {
		if reqID, err = randomID(); err != nil {
			return domain.Lease{}, err
		}
	}
	return s.leases.Create(ctx, outbound.CreateLeaseParams{
		UserID: actor.UserID, HomestayID: c.HomestayID, Model: c.Model, Start: c.Start, End: end,
		Periods: c.Periods, BillingDay: billingDay, Currency: rate.Currency, Rent: rate.Price,
		Note: c.Note, RequestID: reqID, ExpiresAt: s.now().Add(s.opts.HoldTTL), Schedule: plans,
	})
}

func (s *LeaseService) load(ctx context.Context, actor inbound.Principal, id int64) (l domain.Lease, privileged bool, err error) {
	l, err = s.leases.Get(ctx, id)
	if err != nil {
		return l, false, err
	}
	allowed, privileged := stayAccess(ctx, s.homestays, actor, l.UserID, l.HomestayID)
	if !allowed {
		return domain.Lease{}, false, domain.ErrNotFound // do not reveal other tenants' leases
	}
	return l, privileged, nil
}

func (s *LeaseService) Get(ctx context.Context, actor inbound.Principal, id int64) (domain.Lease, error) {
	l, _, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Lease{}, err
	}
	return s.syncGateway(ctx, l), nil
}

func (s *LeaseService) List(ctx context.Context, actor inbound.Principal, userID int64, asHost bool, limit, offset int) ([]domain.Lease, error) {
	var hostID int64
	switch {
	case asHost:
		userID, hostID = 0, actor.UserID // leases at the homestays I own
	case !actor.IsAdmin():
		userID = actor.UserID
	}
	limit, offset = page(limit, offset)
	return s.leases.List(ctx, userID, hostID, limit, offset)
}

func (s *LeaseService) Invoices(ctx context.Context, actor inbound.Principal, q inbound.InvoiceQuery) ([]domain.Invoice, error) {
	var hostID int64
	if !actor.IsAdmin() {
		hostID = actor.UserID
	}
	limit, offset := page(q.Limit, q.Offset)
	return s.leases.ListInvoices(ctx, outbound.InvoiceFilter{Status: q.Status, DueBefore: q.DueBefore, HostID: hostID, Limit: limit, Offset: offset})
}

func (s *LeaseService) PayInvoice(ctx context.Context, actor inbound.Principal, leaseID int64, periodNo int, cmd inbound.PayCommand) (domain.Lease, error) {
	l, err := s.leases.Get(ctx, leaseID)
	if err != nil {
		return domain.Lease{}, err
	}
	if l.UserID != actor.UserID {
		return domain.Lease{}, domain.ErrNotFound
	}
	if l.Status != domain.LeasePending && l.Status != domain.LeaseActive {
		return domain.Lease{}, fmt.Errorf("%w: lease is %s", domain.ErrConflict, leaseStatusName(l.Status))
	}
	if l.Status == domain.LeasePending && l.ExpiresAt != nil && s.now().After(*l.ExpiresAt) {
		return domain.Lease{}, fmt.Errorf("%w: lease hold expired", domain.ErrConflict)
	}
	var inv *domain.Invoice
	for i := range l.Invoices {
		if l.Invoices[i].PeriodNo == periodNo {
			inv = &l.Invoices[i]
		}
	}
	if inv == nil {
		return domain.Lease{}, domain.ErrNotFound
	}
	switch inv.Status {
	case domain.InvoicePaid:
		return l, nil
	case domain.InvoiceVoid:
		return domain.Lease{}, fmt.Errorf("%w: invoice is void", domain.ErrConflict)
	}
	for _, other := range l.Invoices {
		if other.PeriodNo < periodNo && other.Status == domain.InvoiceUnpaid {
			return domain.Lease{}, fmt.Errorf("%w: pay period %d first", domain.ErrConflict, other.PeriodNo+1)
		}
	}
	amount, err := domain.MinorUnits(inv.Amount, l.Currency)
	if err != nil {
		return domain.Lease{}, err
	}
	if amount == 0 {
		err = s.settle(ctx, l, *inv, amount, domain.MethodWallet, "free")
	} else {
		switch cmd.Method {
		case domain.MethodWallet:
			err = s.payWithWallet(ctx, actor, l, *inv, amount)
		case domain.MethodGateway:
			err = s.payWithGateway(ctx, actor, l, *inv, amount, cmd.Provider)
		default:
			err = fmt.Errorf("%w: method must be %q or %q", domain.ErrInvalid, domain.MethodWallet, domain.MethodGateway)
		}
	}
	if err != nil {
		return domain.Lease{}, err
	}
	return s.leases.Get(ctx, leaseID)
}

func (s *LeaseService) settle(ctx context.Context, l domain.Lease, inv domain.Invoice, amount int64, m domain.PaymentMethod, ref string) error {
	if _, err := s.leases.MarkInvoicePaid(ctx, inv.ID, m, ref); err != nil {
		return err
	}
	s.rail.earn(ctx, l.UserID, fmt.Sprintf("rent-house-invoice-%d", inv.ID), amount)
	return nil
}

func (s *LeaseService) payWithWallet(ctx context.Context, actor inbound.Principal, l domain.Lease, inv domain.Invoice, amount int64) error {
	h, err := s.homestays.Get(ctx, l.HomestayID)
	if err != nil {
		return err
	}
	transferID, err := s.rail.chargeWallet(ctx, actor.UserID, h.WalletID, l.Currency, amount,
		fmt.Sprintf("rent-house-invoice-%d", inv.ID), fmt.Sprintf("Lease #%d rent, period %d", l.ID, inv.PeriodNo+1))
	if err != nil {
		return err
	}
	err = s.settle(ctx, l, inv, amount, domain.MethodWallet, transferID)
	if errors.Is(err, domain.ErrConflict) {
		s.rail.undoWallet(ctx, transferID, fmt.Sprintf("lease #%d invoice no longer payable", l.ID))
		return fmt.Errorf("%w: invoice is no longer payable, payment refunded", domain.ErrConflict)
	}
	return err
}

func (s *LeaseService) payWithGateway(ctx context.Context, actor inbound.Principal, l domain.Lease, inv domain.Invoice, amount int64, provider string) error {
	existing := ""
	if inv.PaymentMethod == domain.MethodGateway {
		existing = inv.PaymentRef
	}
	p, fresh, err := s.rail.gatewayPay(ctx, existing, outbound.CreateGatewayPayment{
		Provider: strings.ToLower(strings.TrimSpace(provider)), Amount: amount, Currency: strings.ToUpper(l.Currency),
		PayerEmail: actor.Email, Description: fmt.Sprintf("Lease #%d rent, period %d", l.ID, inv.PeriodNo+1),
		IdempotencyKey: fmt.Sprintf("rent-house-invoice-%d-%d", inv.ID, inv.Version),
	})
	if err != nil {
		return err
	}
	if !fresh {
		if p.Status == domain.GatewayCaptured {
			return s.settle(ctx, l, inv, amount, domain.MethodGateway, existing)
		}
		return nil // in flight: poll GET /leases/{id}
	}
	ref := strconv.FormatInt(p.ID, 10)
	if _, err := s.leases.SetInvoiceAttempt(ctx, inv.ID, domain.MethodGateway, ref); err != nil {
		return err
	}
	switch p.Status {
	case domain.GatewayCaptured:
		return s.settle(ctx, l, inv, amount, domain.MethodGateway, ref)
	case domain.GatewayFailed:
		return domain.ErrPaymentFailed
	}
	return nil
}

func (s *LeaseService) syncGateway(ctx context.Context, l domain.Lease) domain.Lease {
	changed := false
	for _, inv := range l.Invoices {
		if inv.Status != domain.InvoiceUnpaid || inv.PaymentMethod != domain.MethodGateway || inv.PaymentRef == "" {
			continue
		}
		if p, ok := s.rail.gatewayStatus(ctx, inv.PaymentRef); ok && p.Status == domain.GatewayCaptured {
			amount, _ := domain.MinorUnits(inv.Amount, l.Currency)
			if s.settle(ctx, l, inv, amount, domain.MethodGateway, inv.PaymentRef) == nil {
				changed = true
			}
		}
	}
	if changed {
		if fresh, err := s.leases.Get(ctx, l.ID); err == nil {
			return fresh
		}
	}
	return l
}

func (s *LeaseService) Cancel(ctx context.Context, actor inbound.Principal, id int64) (domain.Lease, error) {
	l, err := s.Get(ctx, actor, id)
	if err != nil {
		return domain.Lease{}, err
	}
	if l.Status == domain.LeaseCancelled {
		return l, nil
	}
	if l.Status != domain.LeasePending {
		return domain.Lease{}, fmt.Errorf("%w: only a lease that has not been paid yet can be cancelled; ask the manager to end an active one", domain.ErrConflict)
	}
	for _, inv := range l.Invoices {
		if err := s.rail.refund(ctx, inv.PaymentMethod, inv.PaymentRef, fmt.Sprintf("lease #%d cancelled", l.ID), true); err != nil {
			return domain.Lease{}, err
		}
	}
	if _, err := s.leases.Expire(ctx, id); err != nil {
		return domain.Lease{}, err
	}
	return s.leases.Get(ctx, id)
}

func (s *LeaseService) Terminate(ctx context.Context, actor inbound.Principal, id int64) (domain.Lease, error) {
	l, privileged, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Lease{}, err
	}

	if !privileged {
		return domain.Lease{}, domain.ErrForbidden
	}
	switch l.Status {
	case domain.LeaseActive:
		if l, err = s.leases.Terminate(ctx, id, s.today()); err != nil {
			return domain.Lease{}, err
		}
	case domain.LeaseEnded:
	default:
		return domain.Lease{}, fmt.Errorf("%w: lease is %s", domain.ErrConflict, leaseStatusName(l.Status))
	}
	var failed error
	for _, inv := range l.Invoices {
		if inv.Status != domain.InvoiceRefunded {
			continue
		}
		reason := fmt.Sprintf("lease #%d ended, period %d not started", l.ID, inv.PeriodNo+1)
		if err := s.rail.refund(ctx, inv.PaymentMethod, inv.PaymentRef, reason, false); err != nil {
			s.opts.Log.Error("refund on lease termination", "lease", l.ID, "invoice", inv.ID, "err", err)
			failed = err
		}
	}
	if failed != nil {
		return domain.Lease{}, fmt.Errorf("lease ended but a refund failed, call terminate again to retry: %w", failed)
	}
	return l, nil
}

func (s *LeaseService) ReleaseExpired(ctx context.Context) (int, error) {
	now := s.now()
	expired, err := s.leases.ListExpired(ctx, now, expireBatch)
	if err != nil {
		return 0, err
	}
	released := 0
	for _, e := range expired {
		l, err := s.leases.Get(ctx, e.ID)
		if err != nil {
			continue
		}
		if first := firstInvoice(l); first != nil && first.PaymentMethod == domain.MethodGateway && first.PaymentRef != "" {
			switch s.rail.judgeStale(ctx, first.PaymentRef, l.ExpiresAt, s.opts.HoldTTL, now) {
			case staleKeep:
				continue
			case stalePaid:
				amount, _ := domain.MinorUnits(first.Amount, l.Currency)
				if err := s.settle(ctx, l, *first, amount, domain.MethodGateway, first.PaymentRef); err != nil {
					s.opts.Log.Error("settle captured payment", "lease", l.ID, "err", err)
				}
				continue
			}
		}
		ok, err := s.leases.Expire(ctx, l.ID)
		if err != nil {
			s.opts.Log.Error("expire lease", "lease", l.ID, "err", err)
			continue
		}
		if ok {
			released++
		}
	}
	return released, nil
}

func firstInvoice(l domain.Lease) *domain.Invoice {
	for i := range l.Invoices {
		if l.Invoices[i].PeriodNo == 0 {
			return &l.Invoices[i]
		}
	}
	return nil
}

func leaseStatusName(s domain.LeaseStatus) string {
	switch s {
	case domain.LeasePending:
		return "pending"
	case domain.LeaseActive:
		return "active"
	case domain.LeaseEnded:
		return "ended"
	}
	return "cancelled"
}
