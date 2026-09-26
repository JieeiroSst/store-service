package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

const resaleStaleAfter = 2 * time.Minute

type ResaleOptions struct {
	FeePercent       int
	PlatformWalletID string
	MaxTransfers     int
	Notifier         *Notifier
	Log              *slog.Logger
}

type ResaleService struct {
	repo    outbound.ResaleRepository
	events  outbound.EventRepository
	rail    *PaymentRail
	opts    ResaleOptions
	page    func(int) int
	now     func() time.Time
	newCode func() (string, error)
}

var _ inbound.ResaleUseCase = (*ResaleService)(nil)

func NewResaleService(repo outbound.ResaleRepository, events outbound.EventRepository, rail *PaymentRail, opts ResaleOptions) *ResaleService {
	opts.Log = defaultLog(opts.Log)
	return &ResaleService{repo: repo, events: events, rail: rail, opts: opts, page: page, now: time.Now, newCode: newTicketCode}
}

func (s *ResaleService) Browse(ctx context.Context, q inbound.ResaleQuery) (domain.Page[domain.ResaleListing], error) {
	if q.EventID <= 0 {
		return domain.Page[domain.ResaleListing]{}, invalid("event_id is required")
	}
	e, err := s.events.Get(ctx, q.EventID)
	if err != nil {
		return domain.Page[domain.ResaleListing]{}, err
	}
	if e.Status != domain.EventPublished {
		return domain.Page[domain.ResaleListing]{}, domain.ErrNotFound
	}
	limit := s.page(q.Limit)
	rows, err := s.repo.List(ctx, outbound.ResaleFilter{EventID: q.EventID, TypeID: q.TypeID, Status: domain.ResaleOpen,
		CheapestFirst: q.CheapestFirst, AfterID: q.AfterID, Limit: limit + 1})
	if err != nil {
		return domain.Page[domain.ResaleListing]{}, err
	}
	if q.CheapestFirst { // not paged: the cheapest `limit`
		return domain.Page[domain.ResaleListing]{Items: rows[:min(len(rows), limit)]}, nil
	}
	return domain.PageOf(rows, limit, func(l domain.ResaleListing) int64 { return l.ID }), nil
}

func (s *ResaleService) Mine(ctx context.Context, actor inbound.Principal, status domain.ResaleStatus, afterID int64, limit int) (domain.Page[domain.ResaleListing], error) {
	limit = s.page(limit)
	rows, err := s.repo.List(ctx, outbound.ResaleFilter{SellerID: actor.UserID, Status: status, AfterID: afterID, Limit: limit + 1})
	if err != nil {
		return domain.Page[domain.ResaleListing]{}, err
	}
	return domain.PageOf(rows, limit, func(l domain.ResaleListing) int64 { return l.ID }), nil
}

func (s *ResaleService) wallets() error {
	if s.rail.wallets == nil {
		return invalid("resale is paid through wallets, and wallet payment is not enabled")
	}
	return nil
}

func (s *ResaleService) List(ctx context.Context, actor inbound.Principal, ticketID, price int64) (domain.ResaleListing, error) {
	if err := s.wallets(); err != nil {
		return domain.ResaleListing{}, err
	}
	if _, err := s.rail.wallets.GetByUser(ctx, actor.UserID); errors.Is(err, domain.ErrNotFound) {
		return domain.ResaleListing{}, invalid("you have no wallet yet: create one, that is where the sale is paid")
	} else if err != nil {
		return domain.ResaleListing{}, err
	}
	return s.repo.Create(ctx, outbound.CreateListing{TicketID: ticketID, SellerID: actor.UserID, Price: price, MaxTransfers: s.opts.MaxTransfers})
}

func (s *ResaleService) Cancel(ctx context.Context, actor inbound.Principal, id int64) (domain.ResaleListing, error) {
	l, err := s.repo.Cancel(ctx, id, actor.UserID)
	if errors.Is(err, domain.ErrConflict) {
		return domain.ResaleListing{}, fmt.Errorf("%w: the listing is being bought or is already gone", domain.ErrConflict)
	}
	return l, err
}

func (s *ResaleService) Buy(ctx context.Context, actor inbound.Principal, id int64) (domain.Ticket, error) {
	if err := s.wallets(); err != nil {
		return domain.Ticket{}, err
	}
	l, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Ticket{}, err
	}
	switch {
	case l.SellerID == actor.UserID:
		return domain.Ticket{}, invalid("you cannot buy your own ticket")
	case l.Status != domain.ResaleOpen:
		return domain.Ticket{}, fmt.Errorf("%w: the ticket is no longer for sale", domain.ErrConflict)
	}
	e, err := s.events.Get(ctx, l.EventID)
	if err != nil {
		return domain.Ticket{}, err
	}
	if e.Status != domain.EventPublished || !s.now().Before(l.StartsAt) {
		return domain.Ticket{}, fmt.Errorf("%w: the event is not selling tickets any more", domain.ErrConflict)
	}

	// 1. the listing is mine
	if l, err = s.repo.Claim(ctx, id, actor.UserID); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return domain.Ticket{}, fmt.Errorf("%w: somebody else got there first", domain.ErrConflict)
		}
		return domain.Ticket{}, err
	}
	giveUp := func(undo ...string) {
		for _, ref := range undo {
			if ref != "" {
				s.rail.undoWallet(context.WithoutCancel(ctx), ref, fmt.Sprintf("resale #%d did not complete", id))
			}
		}
		if err := s.repo.Release(context.WithoutCancel(ctx), id); err != nil {
			s.opts.Log.Error("resale listing not released", "listing", id, "err", err)
		}
	}

	// 2. the money
	seller, err := s.rail.wallets.GetByUser(ctx, l.SellerID)
	if err != nil {
		giveUp()
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Ticket{}, fmt.Errorf("%w: the seller can no longer be paid", domain.ErrConflict)
		}
		return domain.Ticket{}, err
	}
	fee := int64(0)
	if s.opts.PlatformWalletID != "" {
		fee = domain.ResaleFee(l.Price, s.opts.FeePercent)
	}
	desc := fmt.Sprintf("Resale of a ticket to %s (listing #%d)", e.Title, id)
	ref1, err := s.rail.chargeWallet(ctx, actor.UserID, seller.ID, l.Currency, l.Price-fee, fmt.Sprintf("resale-%d", id), desc)
	if err != nil {
		giveUp()
		return domain.Ticket{}, err
	}
	ref2 := ""
	if fee > 0 {
		if ref2, err = s.rail.chargeWallet(ctx, actor.UserID, s.opts.PlatformWalletID, l.Currency, fee, fmt.Sprintf("resale-%d-fee", id), desc+" - fee"); err != nil {
			giveUp(ref1)
			return domain.Ticket{}, err
		}
	}
	if err := s.repo.SetPayment(ctx, id, ref1+"|"+ref2, fee); err != nil {
		giveUp(ref1, ref2)
		return domain.Ticket{}, fmt.Errorf("%w: the listing changed while you paid; you were not charged", domain.ErrConflict)
	}

	// 3. the ticket
	code, err := s.newCode()
	if err != nil {
		return domain.Ticket{}, err
	}
	t, err := s.repo.Complete(ctx, id, actor.UserID, actor.Email, code)
	if errors.Is(err, domain.ErrConflict) {
		giveUp(ref1, ref2)
		return domain.Ticket{}, fmt.Errorf("%w: the ticket can no longer be sold; you were not charged", domain.ErrConflict)
	}
	if err != nil {
		return domain.Ticket{}, err
	}
	s.notifySold(ctx, l, t, e, fee)
	return t, nil
}

func (s *ResaleService) notifySold(ctx context.Context, l domain.ResaleListing, t domain.Ticket, e domain.Event, fee int64) {
	n := s.opts.Notifier
	body := fmt.Sprintf("Your ticket to %s was bought. You received %s %s.", e.Title, formatMoney(l.Price-fee), l.Currency)
	n.SendMail(ctx, l.SellerID, "resale_sold", "Your ticket was sold", body, 0, e.ID, nil)
	n.SendMail(ctx, l.BuyerID, "resale_bought", "Your ticket is ready",
		fmt.Sprintf("You bought a ticket to %s. It is in your tickets.", e.Title), 0, e.ID, nil)
}

func (s *ResaleService) Recover(ctx context.Context) (int, error) {
	done := 0
	stale, err := s.repo.Stale(ctx, s.now().Add(-resaleStaleAfter), 100)
	if err != nil {
		return 0, err
	}
	for _, l := range stale {
		if l.PaymentRef == "" {
			if err := s.repo.Release(ctx, l.ID); err != nil {
				s.opts.Log.Error("release stale resale", "listing", l.ID, "err", err)
				continue
			}
			done++
			continue
		}
		code, err := s.newCode()
		if err != nil {
			continue
		}
		_, err = s.repo.Complete(ctx, l.ID, l.BuyerID, "", code)
		switch {
		case err == nil:
			done++
		case errors.Is(err, domain.ErrConflict): 
			for _, ref := range strings.Split(l.PaymentRef, "|") {
				if ref != "" {
					s.rail.undoWallet(ctx, ref, fmt.Sprintf("resale #%d could not be completed", l.ID))
				}
			}
			if err := s.repo.Release(ctx, l.ID); err != nil {
				s.opts.Log.Error("release refunded resale", "listing", l.ID, "err", err)
			}
			done++
		default:
			s.opts.Log.Error("complete stale resale", "listing", l.ID, "err", err)
		}
	}
	n, err := s.repo.ExpireStarted(ctx, s.now(), 500)
	return done + n, err
}
