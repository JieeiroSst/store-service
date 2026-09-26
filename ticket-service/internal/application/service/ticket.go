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

const (
	expiryGrace   = 6 * time.Hour
	expireBatches = 10
	expireSize    = 500
)

type TicketOptions struct {
	TransferTTL  time.Duration
	MaxTransfers int
	Notifier     *Notifier
	Log          *slog.Logger
}

type TicketService struct {
	tickets outbound.TicketRepository
	events  outbound.EventRepository
	staff   outbound.StaffRepository
	dir     outbound.UserDirectory
	opts    TicketOptions
	now     func() time.Time
}

var _ inbound.TicketUseCase = (*TicketService)(nil)

func NewTicketService(t outbound.TicketRepository, e outbound.EventRepository, st outbound.StaffRepository, opts TicketOptions) *TicketService {
	opts.Log = defaultLog(opts.Log)
	if opts.TransferTTL <= 0 {
		opts.TransferTTL = 48 * time.Hour
	}
	return &TicketService{tickets: t, events: e, staff: st, opts: opts, now: time.Now}
}

func (s *TicketService) WithDirectory(d outbound.UserDirectory) *TicketService {
	s.dir = d
	return s
}

func canScan(ctx context.Context, staff outbound.StaffRepository, a inbound.Principal, e domain.Event) bool {
	if canManage(a, e) {
		return true
	}
	if staff == nil || a.UserID == 0 {
		return false
	}
	ok, err := staff.Has(ctx, e.ID, a.UserID)
	return err == nil && ok
}

func (s *TicketService) load(ctx context.Context, actor inbound.Principal, id int64) (t domain.Ticket, holder, privileged bool, err error) {
	t, err = s.tickets.Get(ctx, id)
	if err != nil {
		return t, false, false, err
	}
	if actor.UserID != 0 && t.HolderID == actor.UserID {
		return t, true, false, nil
	}
	e, err := s.events.Get(ctx, t.EventID)
	if err != nil || !canScan(ctx, s.staff, actor, e) {
		return domain.Ticket{}, false, false, domain.ErrNotFound // do not reveal other people's tickets
	}
	return t, false, true, nil
}

func (s *TicketService) List(ctx context.Context, actor inbound.Principal, q inbound.TicketListQuery) (domain.Page[domain.Ticket], error) {
	limit := page(q.Limit)
	rows, err := s.tickets.List(ctx, outbound.TicketFilter{HolderID: actor.UserID, EventID: q.EventID, Status: q.Status, AfterID: q.AfterID, Limit: limit + 1})
	if err != nil {
		return domain.Page[domain.Ticket]{}, err
	}
	return domain.PageOf(rows, limit, func(t domain.Ticket) int64 { return t.ID }), nil
}

func (s *TicketService) Get(ctx context.Context, actor inbound.Principal, id int64) (domain.Ticket, error) {
	t, _, _, err := s.load(ctx, actor, id)
	return t, err
}

func (s *TicketService) History(ctx context.Context, actor inbound.Principal, id int64) ([]domain.TicketEvent, error) {
	if _, _, _, err := s.load(ctx, actor, id); err != nil {
		return nil, err
	}
	return s.tickets.History(ctx, id)
}

func (s *TicketService) SetHolder(ctx context.Context, actor inbound.Principal, id int64, in inbound.HolderInput) (domain.Ticket, error) {
	t, holder, _, err := s.load(ctx, actor, id)
	if err != nil {
		return t, err
	}
	if !holder {
		return domain.Ticket{}, domain.ErrForbidden
	}
	in.Name, in.Email = trim(in.Name), trim(in.Email)
	switch {
	case len(in.Name) > 120:
		return domain.Ticket{}, invalid("name is too long")
	case in.Email != "" && !validEmail(in.Email):
		return domain.Ticket{}, invalid("email must be a valid email address")
	}
	out, err := s.tickets.SetHolder(ctx, id, in.Name, in.Email, actor.UserID)
	return out, s.conflict(err, "only a valid ticket can be edited")
}

func (s *TicketService) Reissue(ctx context.Context, actor inbound.Principal, id int64) (domain.Ticket, error) {
	t, holder, _, err := s.load(ctx, actor, id)
	if err != nil {
		return t, err
	}
	if !holder {
		return domain.Ticket{}, domain.ErrForbidden
	}
	code, err := newTicketCode()
	if err != nil {
		return domain.Ticket{}, err
	}
	out, err := s.tickets.Reissue(ctx, id, code, actor.UserID)
	return out, s.conflict(err, "only a valid ticket can get a new code")
}

func (s *TicketService) conflict(err error, what string) error {
	if errors.Is(err, domain.ErrConflict) {
		return fmt.Errorf("%w: %s", domain.ErrConflict, what)
	}
	return err
}

// ---- transfers

func (s *TicketService) Offer(ctx context.Context, actor inbound.Principal, ticketID int64, c inbound.TransferCommand) (domain.Transfer, error) {
	t, holder, _, err := s.load(ctx, actor, ticketID)
	if err != nil {
		return domain.Transfer{}, err
	}
	if !holder {
		return domain.Transfer{}, domain.ErrForbidden
	}
	e, err := s.events.Get(ctx, t.EventID)
	if err != nil {
		return domain.Transfer{}, err
	}
	now := s.now()
	sess := showtime(e, t.SessionID)
	c.ToEmail, c.Message = trim(c.ToEmail), trim(c.Message)
	switch {
	case !e.Transferable:
		return domain.Transfer{}, fmt.Errorf("%w: the organizer does not allow tickets of this event to be transferred", domain.ErrConflict)
	case e.Status != domain.EventPublished || sess.Status != domain.SessionScheduled || !now.Before(sess.StartsAt):
		return domain.Transfer{}, fmt.Errorf("%w: tickets can only be transferred before a published event starts", domain.ErrConflict)
	case c.ToUserID <= 0 && c.ToEmail == "":
		return domain.Transfer{}, invalid("name the recipient by to_user_id or to_email")
	case c.ToEmail != "" && !validEmail(c.ToEmail):
		return domain.Transfer{}, invalid("to_email must be a valid email address")
	case c.ToUserID == actor.UserID || (c.ToEmail != "" && actor.Email != "" && equalFold(c.ToEmail, actor.Email)):
		return domain.Transfer{}, invalid("you cannot transfer a ticket to yourself")
	case len(c.Message) > 500:
		return domain.Transfer{}, invalid("message is too long")
	}
	var recipientEmail string 
	if c.ToUserID > 0 && s.dir != nil {
		id, err := s.dir.Lookup(ctx, c.ToUserID)
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Transfer{}, invalid("there is no user %d to send the ticket to", c.ToUserID)
		} else if err != nil {
			return domain.Transfer{}, err
		}
		recipientEmail = id.Email
	}
	x, err := s.tickets.Offer(ctx, outbound.OfferParams{
		TicketID: ticketID, FromUser: actor.UserID, ToUser: max(c.ToUserID, 0), ToEmail: c.ToEmail, Message: c.Message,
		ExpiresAt: minTime(now.Add(s.opts.TransferTTL), sess.StartsAt), MaxTransfers: s.opts.MaxTransfers,
	})
	if err != nil {
		return domain.Transfer{}, err
	}
	when := s.opts.Notifier.Format(x.ExpiresAt)
	body := fmt.Sprintf("Someone sent you a ticket to %s. Accept it before %s.", e.Title, when)
	data := map[string]string{"Name": "bạn", "EventTitle": titleOf(e, sess), "EventDate": s.opts.Notifier.Format(sess.StartsAt), "Venue": e.Venue, "Message": c.Message, "ExpiresAt": when}
	var mail *domain.Email
	if recipientEmail != "" {
		mail = &domain.Email{To: recipientEmail, Template: "ticket_transfer_offer", Data: data}
	}
	s.opts.Notifier.SendMail(ctx, x.ToUser, "ticket_offered", "A ticket was sent to you", body, 0, e.ID, mail)
	if c.ToEmail != "" && !equalFold(c.ToEmail, recipientEmail) { // an address: there may be no account to notify
		s.opts.Notifier.MailOnly(ctx, c.ToEmail, "ticket_transfer_offer", "A ticket was sent to you", e.ID, data)
	}
	return x, nil
}

func equalFold(a, b string) bool { return len(a) == len(b) && (a == b || lower(a) == lower(b)) }

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if 'A' <= c && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func minTime(a, b time.Time) time.Time {
	if b.Before(a) {
		return b
	}
	return a
}

func (s *TicketService) Transfers(ctx context.Context, actor inbound.Principal, incoming bool, afterID int64, limit int) (domain.Page[domain.Transfer], error) {
	limit = page(limit)
	var rows []domain.Transfer
	var err error
	if incoming {
		rows, err = s.tickets.Incoming(ctx, actor.UserID, actor.Email, afterID, limit+1)
	} else {
		rows, err = s.tickets.Outgoing(ctx, actor.UserID, afterID, limit+1)
	}
	if err != nil {
		return domain.Page[domain.Transfer]{}, err
	}
	return domain.PageOf(rows, limit, func(t domain.Transfer) int64 { return t.ID }), nil
}

func (s *TicketService) offer(ctx context.Context, id int64) (domain.Transfer, error) {
	x, err := s.tickets.GetTransfer(ctx, id)
	if err != nil {
		return x, err
	}
	return x, nil
}

func (s *TicketService) AcceptTransfer(ctx context.Context, actor inbound.Principal, id int64) (domain.Ticket, error) {
	x, err := s.offer(ctx, id)
	if err != nil {
		return domain.Ticket{}, err
	}
	if !x.AddressedTo(actor.UserID, actor.Email) || x.FromUser == actor.UserID {
		return domain.Ticket{}, domain.ErrNotFound 
	}
	code, err := newTicketCode() 
	if err != nil {
		return domain.Ticket{}, err
	}
	t, err := s.tickets.AcceptTransfer(ctx, id, actor.UserID, "", actor.Email, code)
	if err != nil {
		return domain.Ticket{}, s.conflict(err, "the offer is no longer open (it was answered, called off or has expired)")
	}
	s.opts.Notifier.Send(ctx, x.FromUser, "ticket_accepted", "Your ticket was accepted",
		fmt.Sprintf("Your ticket to %s now belongs to its new holder.", t.EventTitle), 0, t.EventID)
	return t, nil
}

func (s *TicketService) DeclineTransfer(ctx context.Context, actor inbound.Principal, id int64) (domain.Transfer, error) {
	x, err := s.offer(ctx, id)
	if err != nil {
		return x, err
	}
	if !x.AddressedTo(actor.UserID, actor.Email) {
		return domain.Transfer{}, domain.ErrNotFound
	}
	out, err := s.tickets.CloseTransfer(ctx, id, domain.TransferDeclined, actor.UserID)
	return out, s.conflict(err, "the offer is no longer open")
}

func (s *TicketService) CancelTransfer(ctx context.Context, actor inbound.Principal, id int64) (domain.Transfer, error) {
	x, err := s.offer(ctx, id)
	if err != nil {
		return x, err
	}
	if x.FromUser != actor.UserID {
		return domain.Transfer{}, domain.ErrNotFound
	}
	out, err := s.tickets.CloseTransfer(ctx, id, domain.TransferCancelled, actor.UserID)
	return out, s.conflict(err, "the offer is no longer open")
}

// ---- the organizer's hand on a ticket

func (s *TicketService) manage(ctx context.Context, actor inbound.Principal, id int64) (domain.Ticket, error) {
	t, err := s.tickets.Get(ctx, id)
	if err != nil {
		return t, err
	}
	e, err := s.events.Get(ctx, t.EventID)
	if err != nil {
		return t, err
	}
	if !canManage(actor, e) {
		if t.HolderID == actor.UserID {
			return t, domain.ErrForbidden
		}
		return t, notYours(e)
	}
	return t, nil
}

func (s *TicketService) Void(ctx context.Context, actor inbound.Principal, id int64, reason string) (domain.Ticket, error) {
	t, err := s.manage(ctx, actor, id)
	if err != nil {
		return t, err
	}
	if reason = trim(reason); reason == "" {
		return domain.Ticket{}, invalid("give a reason for revoking the ticket")
	}
	out, err := s.tickets.Void(ctx, id, actor.UserID, reason)
	if err != nil {
		return out, s.conflict(err, "only a valid or transferring ticket can be revoked")
	}
	s.opts.Notifier.Send(ctx, t.HolderID, "ticket_voided", "A ticket was revoked",
		fmt.Sprintf("Your ticket to %s was revoked by the organizer: %s", t.EventTitle, reason), 0, t.EventID)
	return out, nil
}

func (s *TicketService) RevertCheckIn(ctx context.Context, actor inbound.Principal, id int64) (domain.Ticket, error) {
	if _, err := s.manage(ctx, actor, id); err != nil {
		return domain.Ticket{}, err
	}
	out, err := s.tickets.RevertCheckIn(ctx, id, actor.UserID)
	return out, s.conflict(err, "the ticket has not been checked in")
}

func (s *TicketService) Lookup(ctx context.Context, actor inbound.Principal, eventID int64, code string) (domain.Ticket, error) {
	code = normalizeCode(code)
	if code == "" {
		return domain.Ticket{}, invalid("code is required")
	}
	if eventID == 0 {
		if err := requireAdmin(actor); err != nil {
			return domain.Ticket{}, err
		}
		return s.tickets.ByCode(ctx, 0, code)
	}
	e, err := s.events.Get(ctx, eventID)
	if err != nil {
		return domain.Ticket{}, err
	}
	if !canScan(ctx, s.staff, actor, e) {
		return domain.Ticket{}, notYours(e)
	}
	return s.tickets.ByCode(ctx, eventID, code)
}

// ---- background work

func (s *TicketService) ExpireTickets(ctx context.Context) (int, error) {
	total := 0
	for i := 0; i < expireBatches; i++ {
		n, err := s.tickets.ExpireTickets(ctx, s.now().Add(-expiryGrace), expireSize)
		total += n
		if err != nil || n < expireSize {
			return total, err
		}
	}
	return total, nil
}

func (s *TicketService) ExpireTransfers(ctx context.Context) (int, error) {
	total := 0
	for i := 0; i < expireBatches; i++ {
		n, err := s.tickets.ExpireTransfers(ctx, s.now(), expireSize)
		total += n
		if err != nil || n < expireSize {
			return total, err
		}
	}
	return total, nil
}
