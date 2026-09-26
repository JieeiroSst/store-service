package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

const (
	maxOrderItems = 10
	maxItemQty    = 100
	reminderAhead = 24 * time.Hour
)

type OrderOptions struct {
	HoldTTL    time.Duration
	Log        *slog.Logger
	SoldOut    *SoldOutCache
	Bulkhead   *Bulkhead
	Limiter    *UserLimiter
	MaxPending int
	Notifier   *Notifier
	Waitlist   TicketsBack
	Invoice    InvoiceOptions
	Lifecycle  outbound.OrderLifecycle
}

// OrderService: pending -> paid | expired | cancelled, paid -> refunded.
type OrderService struct {
	orders     outbound.OrderRepository
	events     outbound.EventRepository
	rail       *PaymentRail
	opts       OrderOptions
	flights    *flights
	existsGate *Bulkhead
	now        func() time.Time
}

var _ inbound.OrderUseCase = (*OrderService)(nil)

func NewOrderService(o outbound.OrderRepository, e outbound.EventRepository, rail *PaymentRail, opts OrderOptions) *OrderService {
	opts.Log = defaultLog(opts.Log)
	if opts.HoldTTL <= 0 {
		opts.HoldTTL = 10 * time.Minute
	}
	return &OrderService{orders: o, events: e, rail: rail, opts: opts, flights: newFlights(), existsGate: NewBulkhead(32, 5*time.Millisecond), now: time.Now}
}

func (s *OrderService) SoldOut(ctx context.Context, c inbound.ReserveCommand) bool {
	cache := s.opts.SoldOut
	if cache == nil {
		return false
	}
	if reqID := trim(c.RequestID); reqID != "" {
		if leave, err := s.existsGate.Enter(ctx); err == nil {
			exists, qerr := s.orders.RequestIDExists(ctx, reqID)
			leave()
			if qerr != nil || exists {
				return false
			}
		}
	}
	for _, it := range c.Items {
		if it.Quantity >= 1 && cache.SoldOut(it.TicketTypeID, it.Quantity) {
			return true
		}
	}
	for _, it := range c.Items {
		if it.Quantity < 1 || cache.RecentlyAvailable(it.TicketTypeID, it.Quantity) {
			continue
		}

		left, err := s.flights.Do(it.TicketTypeID, func() (int, error) {
			lctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
			defer cancel()
			m, err := s.orders.Available(lctx, []int64{it.TicketTypeID})
			if err != nil {
				return 0, err
			}
			n, ok := m[it.TicketTypeID]
			if !ok {
				return 0, domain.ErrNotFound
			}
			return n, nil
		})
		if err != nil {
			continue
		}
		if left < it.Quantity {
			cache.MarkSoldOut(it.TicketTypeID, it.Quantity)
			return true
		}
		cache.MarkAvailable(it.TicketTypeID, left)
	}
	return false
}

func (s *OrderService) validate(actor inbound.Principal, c inbound.ReserveCommand) (inbound.ReserveCommand, error) {
	c.BuyerName, c.BuyerEmail, c.BuyerPhone = trim(c.BuyerName), trim(c.BuyerEmail), trim(c.BuyerPhone)
	if c.BuyerEmail == "" {
		c.BuyerEmail = actor.Email
	}
	c.PromoCode = domain.NormalizeCode(c.PromoCode)
	switch {
	case c.EventID <= 0:
		return c, invalid("event_id is required")
	case len(c.Items) == 0 || len(c.Items) > maxOrderItems:
		return c, invalid("an order has between 1 and %d ticket types", maxOrderItems)
	case c.BuyerName == "" || len(c.BuyerName) > 120:
		return c, invalid("buyer_name is required")
	case !validEmail(c.BuyerEmail):
		return c, invalid("buyer_email must be a valid email address")
	case len(c.BuyerPhone) > 32:
		return c, invalid("buyer_phone is too long")
	case len(c.RequestID) > 128:
		return c, invalid("request id is longer than 128 characters")
	}
	seen := map[int64]bool{}
	items := make([]inbound.ReserveItem, len(c.Items))
	for i, it := range c.Items {
		switch {
		case it.TicketTypeID <= 0:
			return c, invalid("ticket_type_id is required")
		case seen[it.TicketTypeID]:
			return c, invalid("ticket type %d is listed twice; use its quantity", it.TicketTypeID)
		case it.Quantity < 1 || it.Quantity > maxItemQty:
			return c, invalid("quantity must be between 1 and %d", maxItemQty)
		case len(it.SeatIDs) != 0 && len(it.SeatIDs) != it.Quantity:
			return c, invalid("pick exactly %d seats for ticket type %d", it.Quantity, it.TicketTypeID)
		}
		seen[it.TicketTypeID] = true
		ids := make(map[int64]bool, len(it.SeatIDs))
		for _, id := range it.SeatIDs {
			if id <= 0 || ids[id] {
				return c, invalid("seat ids must be positive and distinct")
			}
			ids[id] = true
		}
		items[i] = inbound.ReserveItem{TicketTypeID: it.TicketTypeID, Quantity: it.Quantity, SeatIDs: append([]int64(nil), it.SeatIDs...)}
	}
	c.Items = items
	return c, nil
}

func (s *OrderService) Reserve(ctx context.Context, actor inbound.Principal, c inbound.ReserveCommand) (domain.Order, error) {
	c, err := s.validate(actor, c)
	if err != nil {
		return domain.Order{}, err
	}
	if !s.opts.Limiter.Allow(actor.UserID) {
		return domain.Order{}, fmt.Errorf("%w: too many attempts, wait a moment", domain.ErrBusy)
	}
	if reqID := trim(c.RequestID); reqID != "" {
		if x, found, err := s.orders.FindByRequestID(ctx, actor.UserID, reqID); err != nil {
			return domain.Order{}, err
		} else if found {
			return x, nil
		}
	}
	if s.SoldOut(ctx, c) {
		return domain.Order{}, &domain.SoldOutError{}
	}
	leave, err := s.opts.Bulkhead.Enter(ctx)
	if err != nil {
		return domain.Order{}, err
	}
	defer leave()
	if s.SoldOut(ctx, c) {
		return domain.Order{}, &domain.SoldOutError{}
	}

	tickets := 0
	items := make([]outbound.ReserveItem, len(c.Items))
	for i, it := range c.Items {
		tickets += it.Quantity
		items[i] = outbound.ReserveItem{TicketTypeID: it.TicketTypeID, Quantity: it.Quantity, SeatIDs: it.SeatIDs}
	}
	var promo *domain.Promotion
	if c.PromoCode != "" {
		p, err := s.events.PromotionByCode(ctx, c.EventID, c.PromoCode)
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Order{}, invalid("unknown promo code %q", c.PromoCode)
		}
		if err != nil {
			return domain.Order{}, err
		}
		if why := p.UsableFor(s.now(), tickets); why != "" {
			return domain.Order{}, invalid("%s", why)
		}
		promo = &p
	}
	reqID := trim(c.RequestID)
	if reqID == "" {
		if reqID, err = randomID(); err != nil {
			return domain.Order{}, err
		}
	}
	x, err := s.orders.Reserve(ctx, outbound.ReserveParams{
		UserID: actor.UserID, EventID: c.EventID, Items: items, Promo: promo,
		BuyerName: c.BuyerName, BuyerEmail: c.BuyerEmail, BuyerPhone: c.BuyerPhone,
		RequestID: reqID, ExpiresAt: s.now().Add(s.opts.HoldTTL), MaxPending: s.opts.MaxPending,
	})
	if err == nil {
		s.track(x)
	}
	var so *domain.SoldOutError
	if errors.As(err, &so) {
		for _, it := range c.Items {
			if it.TicketTypeID == so.TicketTypeID {
				s.opts.SoldOut.MarkSoldOut(it.TicketTypeID, it.Quantity) // everyone behind this request is answered from memory
			}
		}
	}
	return x, err
}

func (s *OrderService) load(ctx context.Context, actor inbound.Principal, id int64) (x domain.Order, privileged bool, err error) {
	x, err = s.orders.Get(ctx, id)
	if err != nil {
		return x, false, err
	}
	switch {
	case actor.IsAdmin():
		return x, true, nil
	case x.UserID == actor.UserID:
		return x, false, nil
	}
	e, err := s.events.Get(ctx, x.EventID)
	if err != nil || e.OrganizerID == 0 || e.OrganizerID != actor.UserID {
		return domain.Order{}, false, domain.ErrNotFound
	}
	return x, true, nil
}

func (s *OrderService) Get(ctx context.Context, actor inbound.Principal, id int64) (domain.Order, error) {
	x, _, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Order{}, err
	}
	return s.syncGateway(ctx, x), nil
}

func (s *OrderService) List(ctx context.Context, actor inbound.Principal, q inbound.OrderListQuery) (domain.Page[domain.Order], error) {
	f := outbound.OrderFilter{Status: q.Status, AfterID: q.AfterID}
	if q.EventID != 0 {
		e, err := s.events.Get(ctx, q.EventID)
		if err != nil {
			return domain.Page[domain.Order]{}, err
		}
		if !canManage(actor, e) {
			return domain.Page[domain.Order]{}, domain.ErrNotFound
		}
		f.EventID = q.EventID
	} else {
		f.UserID = actor.UserID
	}
	limit := page(q.Limit)
	f.Limit = limit + 1
	rows, err := s.orders.List(ctx, f)
	if err != nil {
		return domain.Page[domain.Order]{}, err
	}
	return domain.PageOf(rows, limit, func(x domain.Order) int64 { return x.ID }), nil
}

// ---- payment

func (s *OrderService) Pay(ctx context.Context, actor inbound.Principal, id int64, cmd inbound.PayCommand) (domain.Order, error) {
	x, err := s.orders.Get(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	if x.UserID != actor.UserID {
		return domain.Order{}, domain.ErrNotFound
	}
	switch x.Status {
	case domain.OrderPaid:
		return x, nil
	case domain.OrderPending:
	default:
		return domain.Order{}, fmt.Errorf("%w: the order is %s", domain.ErrConflict, x.Status.Name())
	}
	if s.now().After(x.ExpiresAt) {
		return domain.Order{}, fmt.Errorf("%w: the hold on the tickets expired", domain.ErrConflict)
	}
	if x.Total == 0 {
		return s.confirm(ctx, x, domain.MethodFree, "free")
	}
	switch cmd.Method {
	case domain.MethodWallet:
		return s.payWithWallet(ctx, actor, x)
	case domain.MethodGateway:
		return s.payWithGateway(ctx, actor, x, cmd.Provider)
	default:
		return domain.Order{}, invalid("method must be %q or %q", domain.MethodWallet, domain.MethodGateway)
	}
}

func (s *OrderService) confirm(ctx context.Context, x domain.Order, m domain.PaymentMethod, ref string) (domain.Order, error) {
	codes := make([]string, x.Quantity())
	for i := range codes {
		c, err := newTicketCode()
		if err != nil {
			return domain.Order{}, err
		}
		codes[i] = c
	}
	paid, err := s.orders.MarkPaid(ctx, x.ID, m, ref, codes)
	if err == nil {
		s.notifyPaid(ctx, paid)
		s.nudge(paid.ID)
	}
	return paid, err
}

func (s *OrderService) payWithWallet(ctx context.Context, actor inbound.Principal, x domain.Order) (domain.Order, error) {
	e, err := s.events.Get(ctx, x.EventID)
	if err != nil {
		return domain.Order{}, err
	}
	transferID, err := s.rail.chargeWallet(ctx, actor.UserID, e.WalletID, x.Currency, x.Total,
		fmt.Sprintf("order-%d", x.ID), fmt.Sprintf("Tickets for %s (order #%d)", e.Title, x.ID))
	if err != nil {
		return domain.Order{}, err
	}
	paid, err := s.confirm(ctx, x, domain.MethodWallet, transferID)
	if errors.Is(err, domain.ErrConflict) {
		// The order ended (expired, cancelled) while the money moved: give the money back.
		s.rail.undoWallet(ctx, transferID, fmt.Sprintf("order #%d no longer payable", x.ID))
		return domain.Order{}, fmt.Errorf("%w: the order is no longer payable, payment refunded", domain.ErrConflict)
	}
	return paid, err
}

func (s *OrderService) payWithGateway(ctx context.Context, actor inbound.Principal, x domain.Order, provider string) (domain.Order, error) {
	existing := ""
	if x.PaymentMethod == domain.MethodGateway {
		existing = x.PaymentRef
	}
	p, fresh, err := s.rail.gatewayPay(ctx, existing, outbound.CreateGatewayPayment{
		Provider: strings.ToLower(trim(provider)), Amount: x.Total, Currency: strings.ToUpper(x.Currency),
		PayerEmail: actor.Email, Description: fmt.Sprintf("Tickets, order #%d", x.ID),
		IdempotencyKey: fmt.Sprintf("order-%d-%d", x.ID, x.Version),
	})
	if err != nil {
		return domain.Order{}, err
	}
	if !fresh {
		if p.Status == domain.GatewayCaptured {
			return s.confirm(ctx, x, domain.MethodGateway, existing)
		}
		return x, nil
	}
	ref := strconv.FormatInt(p.ID, 10)
	if x, err = s.orders.SetPaymentAttempt(ctx, x.ID, domain.MethodGateway, ref); err != nil {
		return domain.Order{}, err
	}
	switch p.Status {
	case domain.GatewayCaptured:
		return s.confirm(ctx, x, domain.MethodGateway, ref)
	case domain.GatewayFailed:
		return domain.Order{}, domain.ErrPaymentFailed
	}
	return x, nil
}

func (s *OrderService) syncGateway(ctx context.Context, x domain.Order) domain.Order {
	if x.Status != domain.OrderPending || x.PaymentMethod != domain.MethodGateway || x.PaymentRef == "" {
		return x
	}
	p, ok := s.rail.gatewayStatus(ctx, x.PaymentRef)
	if !ok || p.Status != domain.GatewayCaptured {
		return x
	}
	if paid, err := s.confirm(ctx, x, domain.MethodGateway, x.PaymentRef); err == nil {
		return paid
	}
	return x
}

// ---- cancellation and refund

func (s *OrderService) Cancel(ctx context.Context, actor inbound.Principal, id int64) (domain.Order, error) {
	x, privileged, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Order{}, err
	}
	x = s.syncGateway(ctx, x)
	note := "cancelled by the buyer"
	if privileged {
		note = "cancelled by the organizer"
	}
	switch x.Status {
	case domain.OrderPending:
		if err := s.rail.refund(ctx, x.PaymentMethod, x.PaymentRef, fmt.Sprintf("order #%d cancelled", x.ID), true); err != nil {
			return domain.Order{}, err
		}
		return s.end(ctx, outbound.TransitionParams{ID: x.ID, From: []domain.OrderStatus{domain.OrderPending}, To: domain.OrderCancelled, Note: note})

	case domain.OrderPaid:
		if !privileged {
			e, err := s.events.Get(ctx, x.EventID)
			if err != nil {
				return domain.Order{}, err
			}
			if !e.RefundableAt(showtime(e, x.SessionID).StartsAt, s.now()) {
				return domain.Order{}, fmt.Errorf("%w: tickets for this event cannot be refunded now", domain.ErrConflict)
			}
		}
		for _, t := range x.Tickets {
			if !privileged && (t.HolderID != x.UserID || t.Status == domain.TicketTransferring || t.Status == domain.TicketListed) {
				return domain.Order{}, fmt.Errorf("%w: some tickets of this order were passed on to someone else", domain.ErrConflict)
			}
			if t.CheckedInAt != nil {
				return domain.Order{}, fmt.Errorf("%w: a ticket of this order was already used", domain.ErrConflict)
			}
		}
		if err := s.rail.refund(ctx, x.PaymentMethod, x.PaymentRef, fmt.Sprintf("order #%d refunded", x.ID), false); err != nil {
			return domain.Order{}, err
		}
		done, err := s.end(ctx, outbound.TransitionParams{ID: x.ID, From: []domain.OrderStatus{domain.OrderPaid}, To: domain.OrderRefunded, Note: note, Refund: x.Total})
		if err == nil {
			s.notifyRefunded(ctx, done, "Your tickets were refunded")
		}
		return done, err
	}
	return x, nil
}

func (s *OrderService) end(ctx context.Context, p outbound.TransitionParams) (domain.Order, error) {
	x, err := s.orders.Transition(ctx, p)
	if errors.Is(err, domain.ErrConflict) {
		return domain.Order{}, fmt.Errorf("%w: the order changed while you were acting on it; look at it again", domain.ErrConflict)
	}
	if err == nil {
		s.freed(ctx, x)
		s.nudge(x.ID)
	}
	return x, err
}

func (s *OrderService) freed(ctx context.Context, x domain.Order) {
	for _, it := range x.Items {
		s.opts.SoldOut.Release(it.TicketTypeID)
		if s.opts.Waitlist != nil {
			s.opts.Waitlist.TicketsBack(ctx, it.TicketTypeID, it.Quantity)
		}
	}
}

// ---- background work

type release int

const (
	released release = iota
	keptHeld
	nowPaid
	nothingToDo
)

func (s *OrderService) releaseOne(ctx context.Context, x domain.Order, now time.Time) (release, error) {
	if x.PaymentMethod == domain.MethodGateway && x.PaymentRef != "" {
		switch s.rail.judgeStale(ctx, x.PaymentRef, x.ExpiresAt, s.opts.HoldTTL, now) {
		case staleKeep:
			return keptHeld, nil
		case stalePaid:
			if _, err := s.confirm(ctx, x, domain.MethodGateway, x.PaymentRef); err != nil {
				return nothingToDo, fmt.Errorf("confirm captured payment: %w", err)
			}
			return nowPaid, nil
		}
	}
	done, err := s.orders.Transition(ctx, outbound.TransitionParams{ID: x.ID, From: []domain.OrderStatus{domain.OrderPending}, To: domain.OrderExpired, Note: "hold expired"})
	switch {
	case err == nil:
		s.freed(ctx, done)
		return released, nil
	case errors.Is(err, domain.ErrConflict):
		return nothingToDo, nil
	}
	return nothingToDo, err
}

func (s *OrderService) ReleaseExpired(ctx context.Context) (int, error) {
	now := s.now()
	expired, err := s.orders.ListExpired(ctx, now, expireBatch)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, x := range expired {
		res, err := s.releaseOne(ctx, x, now)
		switch {
		case err != nil:
			s.opts.Log.Error("expire order", "order", x.ID, "err", err)
		case res == released:
			n++
		}
	}
	return n, nil
}

func (s *OrderService) Advance(ctx context.Context, id int64) (inbound.OrderProgress, error) {
	x, err := s.orders.Get(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		return inbound.OrderProgress{Status: "gone"}, nil
	}
	if err != nil {
		return inbound.OrderProgress{}, err
	}
	now := s.now()
	if x.Status == domain.OrderPending {
		x = s.syncGateway(ctx, x)
		if x.Status == domain.OrderPending && now.After(x.ExpiresAt) {
			res, err := s.releaseOne(ctx, x, now)
			if err != nil {
				return inbound.OrderProgress{}, err
			}
			if res == keptHeld {
				return inbound.OrderProgress{Status: x.Status.Name(), ExpiresAt: x.ExpiresAt, RetryAfter: 30 * time.Second}, nil
			}
			if x, err = s.orders.Get(ctx, id); err != nil {
				return inbound.OrderProgress{}, err
			}
		}
	}
	p := inbound.OrderProgress{Status: x.Status.Name(), ExpiresAt: x.ExpiresAt}
	if x.Status == domain.OrderPaid {
		e, err := s.events.Get(ctx, x.EventID)
		if err != nil {
			return inbound.OrderProgress{}, err
		}
		sess := showtime(e, x.SessionID)
		p.EventStartsAt, p.EventCancelled = sess.StartsAt, e.Status == domain.EventCancelled || sess.Status == domain.SessionCancelled
	}
	return p, nil
}

func (s *OrderService) SettleCancelledEvents(ctx context.Context) (int, error) {
	open, err := s.orders.ListOpenOfCancelledEvents(ctx, expireBatch)
	if err != nil {
		return 0, err
	}
	settled := 0
	for _, x := range open {
		reason := fmt.Sprintf("order #%d: the event was cancelled", x.ID)
		var p outbound.TransitionParams
		switch x.Status {
		case domain.OrderPending:
			if err := s.rail.refund(ctx, x.PaymentMethod, x.PaymentRef, reason, true); err != nil {
				s.opts.Log.Error("refund pending order of cancelled event", "order", x.ID, "err", err)
				continue
			}
			p = outbound.TransitionParams{ID: x.ID, From: []domain.OrderStatus{domain.OrderPending}, To: domain.OrderCancelled, Note: "event cancelled"}
		case domain.OrderPaid:
			if err := s.rail.refund(ctx, x.PaymentMethod, x.PaymentRef, reason, false); err != nil {
				s.opts.Log.Error("refund paid order of cancelled event", "order", x.ID, "err", err)
				continue
			}
			p = outbound.TransitionParams{ID: x.ID, From: []domain.OrderStatus{domain.OrderPaid}, To: domain.OrderRefunded, Note: "event cancelled", Refund: x.Total, Force: true}
		default:
			continue
		}
		done, err := s.orders.Transition(ctx, p)
		switch {
		case err == nil:
			s.freed(ctx, done)
			settled++
			var mail *domain.Email
			if n := s.opts.Notifier; n.EmailOn() {
				if e, err := s.events.Get(ctx, x.EventID); err == nil {
					sess := showtime(e, x.SessionID)
					reason := e.ReviewNote
					if sess.Status == domain.SessionCancelled && e.Status != domain.EventCancelled {
						reason = sess.CancelReason
					}
					mail = &domain.Email{To: x.BuyerEmail, Template: "event_cancelled", Data: map[string]string{"Name": x.BuyerName,
						"EventTitle": titleOf(e, sess), "Reason": reason, "OrderID": strconv.FormatInt(x.ID, 10)}}
				}
			}
			s.opts.Notifier.SendMail(ctx, x.UserID, "event_cancelled", "Event cancelled",
				fmt.Sprintf("The event of order #%d was cancelled. Your payment has been refunded.", x.ID), x.ID, x.EventID, mail)
		case errors.Is(err, domain.ErrConflict):
		default:
			s.opts.Log.Error("settle order of cancelled event", "order", x.ID, "err", err)
		}
	}
	return settled, nil
}

func (s *OrderService) remind(ctx context.Context, x domain.Order, e domain.Event) bool {
	claimed, err := s.orders.MarkReminded(ctx, x.ID)
	if err != nil || !claimed {
		return false
	}
	sess := showtime(e, x.SessionID)
	var mail *domain.Email
	if s.opts.Notifier.EmailOn() {
		mail = &domain.Email{To: x.BuyerEmail, Template: "event_reminder", Data: map[string]string{"Name": x.BuyerName, "EventTitle": titleOf(e, sess),
			"EventDate": s.opts.Notifier.Format(sess.StartsAt), "Venue": e.Venue}}
	}
	s.opts.Notifier.SendMail(ctx, x.UserID, "event_reminder", "Your event starts soon",
		fmt.Sprintf("%s starts %s at %s. Have your tickets ready.", titleOf(e, sess), s.opts.Notifier.Format(sess.StartsAt), e.Venue), x.ID, x.EventID, mail)
	return true
}

func (s *OrderService) Remind(ctx context.Context, id int64) (bool, error) {
	x, err := s.orders.Get(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		return false, nil
	}
	if err != nil || x.Status != domain.OrderPaid {
		return false, err
	}
	e, err := s.events.Get(ctx, x.EventID)
	if err != nil {
		return false, err
	}
	if sess := showtime(e, x.SessionID); e.Status != domain.EventPublished || sess.Status != domain.SessionScheduled || !s.now().Before(sess.StartsAt) {
		return false, nil
	}
	return s.remind(ctx, x, e), nil
}

func (s *OrderService) SendReminders(ctx context.Context) (int, error) {
	now := s.now()
	due, err := s.orders.DueForReminder(ctx, now, now.Add(reminderAhead), expireBatch)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, x := range due {
		e, err := s.events.Get(ctx, x.EventID)
		if err != nil {
			continue
		}
		if s.remind(ctx, x, e) {
			sent++
		}
	}
	return sent, nil
}

// ---- following each order in a workflow engine

func (s *OrderService) track(x domain.Order) {
	lc := s.opts.Lifecycle
	if lc == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := lc.Track(ctx, x.ID, x.ExpiresAt); err != nil {
			s.opts.Log.Warn("order workflow not started", "order", x.ID, "err", err)
		}
	}()
}

func (s *OrderService) nudge(id int64) {
	lc := s.opts.Lifecycle
	if lc == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := lc.Nudge(ctx, id); err != nil {
			s.opts.Log.Warn("order workflow not told", "order", id, "err", err)
		}
	}()
}

// ---- notifications

func (s *OrderService) notifyPaid(ctx context.Context, x domain.Order) {
	title := "Your tickets are ready"
	body := fmt.Sprintf("Order #%d is paid: %d ticket(s). Show the QR code at the gate.", x.ID, x.Quantity())
	var mail *domain.Email
	if n := s.opts.Notifier; n.EmailOn() && !strings.HasPrefix(x.PaymentRef, "invitation:") {
		if e, err := s.events.Get(ctx, x.EventID); err == nil {
			mail = &domain.Email{To: x.BuyerEmail, Template: "ticket_paid", Data: map[string]string{
				"Name": x.BuyerName, "EventTitle": titleOf(e, showtime(e, x.SessionID)), "EventDate": n.Format(showtime(e, x.SessionID).StartsAt), "Venue": e.Venue,
				"OrderID": strconv.FormatInt(x.ID, 10), "TicketCount": strconv.Itoa(x.Quantity()), "Total": formatMoney(x.Total), "Currency": x.Currency}}
		}
	}
	s.opts.Notifier.SendMail(ctx, x.UserID, "order_paid", title, body, x.ID, x.EventID, mail)
}

func (s *OrderService) notifyRefunded(ctx context.Context, x domain.Order, title string) {
	s.opts.Notifier.Send(ctx, x.UserID, "order_refunded", title, fmt.Sprintf("Order #%d was refunded.", x.ID), x.ID, x.EventID)
}

// ---- invitations

func (s *OrderService) Invite(ctx context.Context, actor inbound.Principal, eventID int64, c inbound.InviteCommand) (domain.Order, error) {
	e, err := s.events.Get(ctx, eventID)
	if err != nil {
		return domain.Order{}, err
	}
	if !canManage(actor, e) {
		return domain.Order{}, notYours(e)
	}
	c.Name, c.Email, c.Note = trim(c.Name), trim(c.Email), trim(c.Note)
	switch {
	case !e.OnSale(s.now()):
		return domain.Order{}, fmt.Errorf("%w: the event is not on sale", domain.ErrConflict)
	case c.UserID <= 0:
		return domain.Order{}, invalid("user_id of the invited person is required")
	case c.TicketTypeID <= 0:
		return domain.Order{}, invalid("ticket_type_id is required")
	case c.Quantity < 1 || c.Quantity > maxItemQty:
		return domain.Order{}, invalid("quantity must be between 1 and %d", maxItemQty)
	case len(c.SeatIDs) != 0 && len(c.SeatIDs) != c.Quantity:
		return domain.Order{}, invalid("pick exactly %d seats", c.Quantity)
	case c.Name == "" || len(c.Name) > 120:
		return domain.Order{}, invalid("name of the invited person is required")
	case !validEmail(c.Email):
		return domain.Order{}, invalid("email must be a valid email address")
	}
	tt, err := s.events.TicketType(ctx, c.TicketTypeID)
	if err != nil || tt.EventID != eventID {
		return domain.Order{}, invalid("ticket type %d is not a ticket type of this event", c.TicketTypeID)
	}
	sess := showtime(e, tt.SessionID)
	reqID, err := randomID()
	if err != nil {
		return domain.Order{}, err
	}
	x, err := s.orders.Reserve(ctx, outbound.ReserveParams{
		UserID: c.UserID, EventID: eventID, Comp: true, InvitedBy: actor.UserID,
		Items:     []outbound.ReserveItem{{TicketTypeID: c.TicketTypeID, Quantity: c.Quantity, SeatIDs: c.SeatIDs}},
		BuyerName: c.Name, BuyerEmail: c.Email, RequestID: reqID, ExpiresAt: s.now().Add(s.opts.HoldTTL),
	})
	if err != nil {
		return domain.Order{}, err
	}
	paid, err := s.confirm(ctx, x, domain.MethodFree, fmt.Sprintf("invitation:%d", actor.UserID))
	if err != nil {
		return domain.Order{}, err
	}
	body := fmt.Sprintf("You are invited to %s: %d ticket(s) are waiting in your account.", e.Title, c.Quantity)
	if c.Note != "" {
		body += " " + c.Note
	}
	var mail *domain.Email
	if n := s.opts.Notifier; n.EmailOn() {
		mail = &domain.Email{To: c.Email, Template: "ticket_invited", Data: map[string]string{"Name": c.Name, "EventTitle": titleOf(e, sess),
			"EventDate": n.Format(sess.StartsAt), "Venue": e.Venue, "TicketCount": strconv.Itoa(c.Quantity), "Note": c.Note}}
	}
	s.opts.Notifier.SendMail(ctx, c.UserID, "ticket_invited", "You have been invited", body, 0, e.ID, mail)
	return paid, nil
}
