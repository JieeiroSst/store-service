package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type DocumentService struct {
	orders  inbound.OrderUseCase
	repo    outbound.OrderRepository
	events  outbound.EventRepository
	tickets outbound.TicketRepository
	docs    outbound.DocumentRepository
	store   outbound.DocumentStore
	render  outbound.DocumentRenderer
	loc     *time.Location
	log     *slog.Logger
}

var _ inbound.DocumentUseCase = (*DocumentService)(nil)

func NewDocumentService(orders inbound.OrderUseCase, repo outbound.OrderRepository, events outbound.EventRepository, tickets outbound.TicketRepository,
	docs outbound.DocumentRepository, store outbound.DocumentStore, render outbound.DocumentRenderer, loc *time.Location, log *slog.Logger) *DocumentService {
	if loc == nil {
		loc = time.UTC
	}
	return &DocumentService{orders: orders, repo: repo, events: events, tickets: tickets, docs: docs, store: store, render: render, loc: loc, log: defaultLog(log)}
}

var system = inbound.Principal{Admin: true}

func (s *DocumentService) InvoicePDF(ctx context.Context, actor inbound.Principal, orderID int64) ([]byte, error) {
	inv, err := s.orders.Invoice(ctx, actor, orderID)
	if err != nil {
		return nil, err
	}
	return s.render.Invoice(inv, s.loc)
}

func (s *DocumentService) ticketDocument(ctx context.Context, e domain.Event, sessionID int64, buyer string, orderID int64, tickets []domain.Ticket) domain.TicketDocument {
	sess := showtime(e, sessionID)
	doc := domain.TicketDocument{OrderID: orderID, EventTitle: titleOf(e, sess), StartsAt: sess.StartsAt, EndsAt: sess.EndsAt, Venue: e.Venue, Address: e.Address,
		BuyerName: buyer, Location: s.loc}
	for _, t := range tickets {
		doc.Tickets = append(doc.Tickets, domain.TicketPage{TicketID: t.ID, TypeName: t.TypeName, Seat: t.SeatLabel, HolderName: t.HolderName, Code: t.Code, Status: t.Status.Name()})
	}
	return doc
}

func (s *DocumentService) TicketPDF(ctx context.Context, actor inbound.Principal, ticketID int64) ([]byte, error) {
	t, err := s.tickets.Get(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if t.HolderID != actor.UserID && !actor.IsAdmin() {
		return nil, domain.ErrNotFound
	}
	if t.Status != domain.TicketValid {
		return nil, fmt.Errorf("%w: the ticket is %s; only a valid ticket has a printable version", domain.ErrConflict, t.Status.Name())
	}
	e, err := s.events.Get(ctx, t.EventID)
	if err != nil {
		return nil, err
	}
	buyer := t.HolderName
	if buyer == "" {
		buyer = t.HolderEmail
	}
	return s.render.Tickets(s.ticketDocument(ctx, e, t.SessionID, buyer, t.OrderID, []domain.Ticket{t}))
}

func (s *DocumentService) List(ctx context.Context, actor inbound.Principal, orderID int64) ([]domain.OrderDocument, error) {
	if _, err := s.orders.Get(ctx, actor, orderID); err != nil {
		return nil, err
	}
	return s.docs.List(ctx, orderID)
}

func (s *DocumentService) Open(ctx context.Context, actor inbound.Principal, orderID int64, kind string) (inbound.DocumentFile, error) {
	if !domain.ValidDocKind(kind) {
		return inbound.DocumentFile{}, invalid("kind must be %s or %s", domain.DocInvoice, domain.DocTickets)
	}
	if _, err := s.orders.Get(ctx, actor, orderID); err != nil {
		return inbound.DocumentFile{}, err
	}
	d, err := s.docs.Get(ctx, orderID, kind)
	if err != nil {
		return inbound.DocumentFile{}, err
	}
	if s.store == nil {
		return inbound.DocumentFile{}, fmt.Errorf("%w: document storage is not enabled", domain.ErrUpstreamUnavailable)
	}
	body, size, err := s.store.Open(ctx, d.FileID)
	if err != nil {
		return inbound.DocumentFile{}, err
	}
	if size <= 0 {
		size = d.Size
	}
	return inbound.DocumentFile{Name: d.FileName, Size: size, Body: body}, nil
}

func (s *DocumentService) Generate(ctx context.Context, orderID int64) (bool, error) {
	if s.store == nil {
		return false, nil
	}
	x, err := s.repo.Get(ctx, orderID)
	if err != nil {
		return false, err
	}
	if x.Status != domain.OrderPaid {
		return false, nil
	}
	have := map[string]bool{}
	existing, err := s.docs.List(ctx, orderID)
	if err != nil {
		return false, err
	}
	for _, d := range existing {
		have[d.Kind] = true
	}
	made := false
	if !have[domain.DocInvoice] {
		inv, err := s.orders.Invoice(ctx, system, orderID)
		if err != nil {
			return made, err
		}
		pdf, err := s.render.Invoice(inv, s.loc)
		if err != nil {
			return made, err
		}
		if err := s.keep(ctx, x, domain.DocInvoice, fmt.Sprintf("invoice-%s.pdf", inv.Number), pdf); err != nil {
			return made, err
		}
		made = true
	}
	if !have[domain.DocTickets] {
		e, err := s.events.Get(ctx, x.EventID)
		if err != nil {
			return made, err
		}
		var valid []domain.Ticket
		for _, t := range x.Tickets {
			if t.Status == domain.TicketValid {
				valid = append(valid, t)
			}
		}
		if len(valid) > 0 {
			pdf, err := s.render.Tickets(s.ticketDocument(ctx, e, x.SessionID, x.BuyerName, x.ID, valid))
			if err != nil {
				return made, err
			}
			if err := s.keep(ctx, x, domain.DocTickets, fmt.Sprintf("tickets-order-%d.pdf", x.ID), pdf); err != nil {
				return made, err
			}
			made = true
		}
	}
	return made, s.docs.MarkDone(ctx, orderID)
}

func (s *DocumentService) keep(ctx context.Context, x domain.Order, kind, name string, pdf []byte) error {
	id, err := s.store.Put(ctx, x.UserID, name, pdf)
	if err != nil {
		return err
	}
	saved, err := s.docs.Save(ctx, domain.OrderDocument{OrderID: x.ID, Kind: kind, FileID: id, FileName: name, Size: int64(len(pdf))})
	if err != nil || !saved {
		if derr := s.store.Delete(context.WithoutCancel(ctx), id); derr != nil {
			s.log.Warn("stored file not removed", "file", id, "err", derr)
		}
		return err
	}
	return nil
}

func (s *DocumentService) GenerateMissing(ctx context.Context) (int, error) {
	if s.store == nil {
		return 0, nil
	}
	ids, err := s.docs.Pending(ctx, 20)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, id := range ids {
		made, err := s.Generate(ctx, id)
		switch {
		case err == nil && made:
			n++
		case err != nil && !errors.Is(err, domain.ErrNotFound):
			s.log.Warn("order documents not made", "order", id, "err", err)
		}
	}
	return n, nil
}
