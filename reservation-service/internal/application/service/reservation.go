package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

const (
	maxRooms      = 10 // rooms in one reservation
	maxStayNights = 30
	maxAddonQty   = 20
)

type ReservationOptions struct {
	HoldTTL    time.Duration
	Log        *slog.Logger
	SoldOut    *SoldOutCache
	Bulkhead   *Bulkhead
	Limiter    *UserLimiter
	MaxPending int
	Notifier   *Notifier
	Promoter   Promoter
}

type Promoter interface {
	Promote(ctx context.Context, roomTypeID int64) (int, error)
}

// pending -> paid | canceled | rejected, paid -> refunded.
type ReservationService struct {
	reservations outbound.ReservationRepository
	hotels       outbound.HotelRepository
	rail         *PaymentRail
	opts         ReservationOptions
	flights      *flights
	existsGate   *Bulkhead
	now          func() time.Time
}

var _ inbound.ReservationUseCase = (*ReservationService)(nil)

func NewReservationService(r outbound.ReservationRepository, h outbound.HotelRepository, rail *PaymentRail, opts ReservationOptions) *ReservationService {
	opts.Log = defaultLog(opts.Log)
	if opts.HoldTTL <= 0 {
		opts.HoldTTL = 15 * time.Minute
	}
	return &ReservationService{reservations: r, hotels: h, rail: rail, opts: opts, flights: newFlights(), existsGate: NewBulkhead(32, 5*time.Millisecond), now: time.Now}
}

func (s *ReservationService) today() time.Time { return s.now().UTC().Truncate(24 * time.Hour) }

func (s *ReservationService) SoldOut(ctx context.Context, c inbound.ReserveCommand) bool {
	cache := s.opts.SoldOut
	if cache == nil {
		return false
	}

	if reqID := trim(c.RequestID); reqID != "" {
		if leave, err := s.existsGate.Enter(ctx); err == nil {
			exists, qerr := s.reservations.RequestIDExists(ctx, reqID)
			leave()
			if qerr != nil || exists {
				return false
			}
		}
	}
	if cache.SoldOut(c.RoomTypeID, c.Start, c.End, c.Rooms) {
		return true
	}
	if cache.RecentlyAvailable(c.RoomTypeID, c.Start, c.End, c.Rooms) {
		return false
	}
	key := soldOutKey{c.RoomTypeID, c.Start, c.End}
	left, err := s.flights.Do(key, func() (int, error) {
		return s.reservations.RoomsLeft(ctx, c.HotelID, c.RoomTypeID, c.Start, c.End)
	})
	if err != nil {
		return false
	}
	if left < c.Rooms {
		cache.MarkSoldOut(c.RoomTypeID, c.Start, c.End, c.Rooms)
		return true
	}
	cache.MarkAvailable(c.RoomTypeID, c.Start, c.End, left)
	return false
}

func (s *ReservationService) Reserve(ctx context.Context, actor inbound.Principal, c inbound.ReserveCommand) (domain.Reservation, error) {
	invalid := func(format string, a ...any) (domain.Reservation, error) {
		return domain.Reservation{}, fmt.Errorf("%w: "+format, append([]any{domain.ErrInvalid}, a...)...)
	}
	if c.Start.Before(s.today()) {
		return invalid("check-in is in the past")
	}
	if !c.End.After(c.Start) {
		return invalid("check-out must be after check-in")
	}
	nights := int(c.End.Sub(c.Start).Hours() / 24)
	if nights > maxStayNights {
		return invalid("a stay cannot exceed %d nights", maxStayNights)
	}
	if c.Rooms < 1 || c.Rooms > maxRooms {
		return invalid("rooms must be between 1 and %d", maxRooms)
	}
	if c.Adults < 1 || c.Children < 0 {
		return invalid("at least one adult is required")
	}
	if len(c.RequestID) > 128 {
		return invalid("request id is longer than 128 characters")
	}

	if !s.opts.Limiter.Allow(actor.UserID) {
		return domain.Reservation{}, fmt.Errorf("%w: too many attempts, wait a moment", domain.ErrBusy)
	}
	if reqID := trim(c.RequestID); reqID != "" {
		if x, found, err := s.reservations.FindByRequestID(ctx, actor.UserID, reqID); err != nil {
			return domain.Reservation{}, err
		} else if found {
			return x, nil
		}
	}
	if s.SoldOut(ctx, c) {
		return domain.Reservation{}, fmt.Errorf("%w: sold out for those nights", domain.ErrDatesUnavailable)
	}
	leave, err := s.opts.Bulkhead.Enter(ctx)
	if err != nil {
		return domain.Reservation{}, err
	}
	defer leave()
	if s.SoldOut(ctx, c) {
		return domain.Reservation{}, fmt.Errorf("%w: sold out for those nights", domain.ErrDatesUnavailable)
	}

	h, err := s.hotels.Get(ctx, c.HotelID)
	if err != nil {
		return domain.Reservation{}, err
	}
	if h.Status != domain.HotelActive {
		return domain.Reservation{}, domain.ErrNotFound
	}
	if h.OwnerID != 0 && h.OwnerID == actor.UserID {
		return domain.Reservation{}, fmt.Errorf("%w: you cannot reserve a room at your own hotel", domain.ErrForbidden)
	}
	var rt *domain.RoomType
	for i := range h.RoomTypes {
		if h.RoomTypes[i].ID == c.RoomTypeID && h.RoomTypes[i].Active {
			rt = &h.RoomTypes[i]
		}
	}
	if rt == nil {
		return domain.Reservation{}, fmt.Errorf("%w: room type is not offered by this hotel", domain.ErrNotFound)
	}
	if guests := c.Adults + c.Children; guests > c.Rooms*rt.Capacity {
		return invalid("%d room(s) of %s sleep at most %d guests", c.Rooms, rt.Name, c.Rooms*rt.Capacity)
	}

	addons, addonsTotal, err := s.priceAddons(ctx, h.ID, c.Addons, nights)
	if err != nil {
		return domain.Reservation{}, err
	}
	var promo *domain.Promotion
	if code := trim(c.PromoCode); code != "" {
		p, err := s.hotels.PromotionByCode(ctx, h.ID, code)
		if errors.Is(err, domain.ErrNotFound) {
			return invalid("unknown promo code %q", code)
		}
		if err != nil {
			return domain.Reservation{}, err
		}
		if why := p.UsableFor(c.Start, nights); why != "" {
			return invalid("%s", why)
		}
		promo = &p
	}
	reqID := trim(c.RequestID)
	if reqID == "" {
		if reqID, err = randomID(); err != nil {
			return domain.Reservation{}, err
		}
	}
	x, err := s.reservations.Reserve(ctx, outbound.ReserveParams{
		GuestID: actor.UserID, HotelID: h.ID, RoomTypeID: rt.ID, Start: c.Start, End: c.End,
		Rooms: c.Rooms, Adults: c.Adults, Children: c.Children, Currency: h.Currency,
		SpecialRequests: trim(c.SpecialRequests), RequestID: reqID, ExpiresAt: s.now().Add(s.opts.HoldTTL),
		Addons: addons, AddonsTotal: addonsTotal, MaxPending: s.opts.MaxPending, Promo: promo, Actor: actor.UserID,
	})
	if errors.Is(err, domain.ErrDatesUnavailable) {
		s.opts.SoldOut.MarkSoldOut(rt.ID, c.Start, c.End, c.Rooms) // everyone behind this request is answered from memory
	}
	if err == nil {
		s.notifyCreated(ctx, x, h)
	}
	return x, err
}

func (s *ReservationService) priceAddons(ctx context.Context, hotelID int64, req []inbound.AddonRequest, nights int) ([]domain.ReservationAddon, string, error) {
	if len(req) == 0 {
		return nil, "0.0000", nil
	}
	catalogue, err := s.hotels.Services(ctx, hotelID, true)
	if err != nil {
		return nil, "", err
	}
	byID := make(map[int64]domain.HotelService, len(catalogue))
	for _, c := range catalogue {
		byID[c.ID] = c
	}
	seen := map[int64]bool{}
	out := make([]domain.ReservationAddon, 0, len(req))
	amounts := make([]string, 0, len(req))
	for _, r := range req {
		svc, ok := byID[r.ServiceID]
		if !ok {
			return nil, "", fmt.Errorf("%w: service %d is not offered by this hotel", domain.ErrInvalid, r.ServiceID)
		}
		if seen[r.ServiceID] {
			return nil, "", fmt.Errorf("%w: service %d is listed twice; use its quantity", domain.ErrInvalid, r.ServiceID)
		}
		seen[r.ServiceID] = true
		if r.Quantity < 1 || r.Quantity > maxAddonQty {
			return nil, "", fmt.Errorf("%w: quantity must be between 1 and %d", domain.ErrInvalid, maxAddonQty)
		}
		amount, err := domain.AddonAmount(svc.Price, svc.Unit, r.Quantity, nights)
		if err != nil {
			return nil, "", err
		}
		out = append(out, domain.ReservationAddon{ServiceID: svc.ID, Name: svc.Name, Quantity: r.Quantity, UnitPrice: svc.Price, Amount: amount})
		amounts = append(amounts, amount)
	}
	total, err := domain.SumAmounts(amounts...)
	return out, total, err
}

func (s *ReservationService) load(ctx context.Context, actor inbound.Principal, id int64) (x domain.Reservation, privileged bool, err error) {
	x, err = s.reservations.Get(ctx, id)
	if err != nil {
		return x, false, err
	}
	allowed, privileged := reservationAccess(ctx, s.hotels, actor, x.GuestID, x.HotelID)
	if !allowed {
		return domain.Reservation{}, false, domain.ErrNotFound // do not reveal other guests' reservations
	}
	return x, privileged, nil
}

func (s *ReservationService) Get(ctx context.Context, actor inbound.Principal, id int64) (domain.Reservation, error) {
	x, _, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	return s.syncGateway(ctx, x), nil
}

func (s *ReservationService) List(ctx context.Context, actor inbound.Principal, q inbound.ReservationListQuery) (domain.Page[domain.Reservation], error) {
	f := outbound.ReservationFilter{HotelID: q.HotelID, Status: q.Status, AfterID: q.AfterID}
	switch {
	case q.AsManager:
		f.OwnerID = actor.UserID
	case !actor.IsAdmin():
		f.GuestID = actor.UserID
	}
	limit, _ := page(q.Limit, 0)
	f.Limit = limit + 1
	rows, err := s.reservations.List(ctx, f)
	if err != nil {
		return domain.Page[domain.Reservation]{}, err
	}
	return domain.PageOf(rows, limit, func(x domain.Reservation) int64 { return x.ID }), nil
}

// ---- payment

func (s *ReservationService) Pay(ctx context.Context, actor inbound.Principal, id int64, cmd inbound.PayCommand) (domain.Reservation, error) {
	x, err := s.reservations.Get(ctx, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	if x.GuestID != actor.UserID {
		return domain.Reservation{}, domain.ErrNotFound
	}
	switch x.Status {
	case domain.ReservationPaid:
		return x, nil
	case domain.ReservationPending:
	default:
		return domain.Reservation{}, fmt.Errorf("%w: reservation is %s", domain.ErrConflict, x.Status.Name())
	}
	if x.ExpiresAt != nil && s.now().After(*x.ExpiresAt) {
		return domain.Reservation{}, fmt.Errorf("%w: the hold on the rooms expired", domain.ErrConflict)
	}
	amount, err := domain.MinorUnits(x.TotalAmount, x.Currency)
	if err != nil {
		return domain.Reservation{}, err
	}
	if amount == 0 {
		return s.confirm(ctx, x, amount, domain.MethodWallet, "free")
	}
	switch cmd.Method {
	case domain.MethodWallet:
		return s.payWithWallet(ctx, actor, x, amount)
	case domain.MethodGateway:
		return s.payWithGateway(ctx, actor, x, amount, cmd.Provider)
	default:
		return domain.Reservation{}, fmt.Errorf("%w: method must be %q or %q", domain.ErrInvalid, domain.MethodWallet, domain.MethodGateway)
	}
}

func (s *ReservationService) confirm(ctx context.Context, x domain.Reservation, amount int64, m domain.PaymentMethod, ref string) (domain.Reservation, error) {
	paid, err := s.reservations.MarkPaid(ctx, x.ID, m, ref)
	if err == nil {
		s.rail.earn(ctx, x.GuestID, fmt.Sprintf("reservation-%d", x.ID), amount)
		s.notifyPaid(ctx, paid)
	}
	return paid, err
}

func (s *ReservationService) payWithWallet(ctx context.Context, actor inbound.Principal, x domain.Reservation, amount int64) (domain.Reservation, error) {
	h, err := s.hotels.Get(ctx, x.HotelID)
	if err != nil {
		return domain.Reservation{}, err
	}

	transferID, err := s.rail.chargeWallet(ctx, actor.UserID, h.WalletID, x.Currency, amount,
		fmt.Sprintf("reservation-%d", x.ID), fmt.Sprintf("Reservation #%d at %s", x.ID, h.Name))
	if err != nil {
		return domain.Reservation{}, err
	}
	paid, err := s.confirm(ctx, x, amount, domain.MethodWallet, transferID)
	if errors.Is(err, domain.ErrConflict) {
		s.rail.undoWallet(ctx, transferID, fmt.Sprintf("reservation #%d no longer payable", x.ID))
		return domain.Reservation{}, fmt.Errorf("%w: reservation is no longer payable, payment refunded", domain.ErrConflict)
	}
	return paid, err
}

func (s *ReservationService) payWithGateway(ctx context.Context, actor inbound.Principal, x domain.Reservation, amount int64, provider string) (domain.Reservation, error) {
	existing := ""
	if x.PaymentMethod == domain.MethodGateway {
		existing = x.PaymentRef
	}

	p, fresh, err := s.rail.gatewayPay(ctx, existing, outbound.CreateGatewayPayment{
		Provider: strings.ToLower(trim(provider)), Amount: amount, Currency: strings.ToUpper(x.Currency),
		PayerEmail: actor.Email, Description: fmt.Sprintf("Reservation #%d", x.ID),
		IdempotencyKey: fmt.Sprintf("reservation-%d-%d", x.ID, x.Version),
	})
	if err != nil {
		return domain.Reservation{}, err
	}
	if !fresh {
		if p.Status == domain.GatewayCaptured {
			return s.confirm(ctx, x, amount, domain.MethodGateway, existing)
		}
		return x, nil
	}
	ref := strconv.FormatInt(p.ID, 10)
	if x, err = s.reservations.SetPaymentAttempt(ctx, x.ID, domain.MethodGateway, ref); err != nil {
		return domain.Reservation{}, err
	}
	switch p.Status {
	case domain.GatewayCaptured:
		return s.confirm(ctx, x, amount, domain.MethodGateway, ref)
	case domain.GatewayFailed:
		return domain.Reservation{}, domain.ErrPaymentFailed
	}
	return x, nil
}

func (s *ReservationService) syncGateway(ctx context.Context, x domain.Reservation) domain.Reservation {
	if x.Status != domain.ReservationPending || x.PaymentMethod != domain.MethodGateway || x.PaymentRef == "" {
		return x
	}
	p, ok := s.rail.gatewayStatus(ctx, x.PaymentRef)
	if !ok || p.Status != domain.GatewayCaptured {
		return x
	}
	amount, _ := domain.MinorUnits(x.TotalAmount, x.Currency)
	if paid, err := s.confirm(ctx, x, amount, domain.MethodGateway, x.PaymentRef); err == nil {
		return paid
	}
	return x
}

// ---- cancellation, refund, rejection
func (s *ReservationService) cancellationTerms(x domain.Reservation, h domain.Hotel, privileged bool) (feePercent int, fee, refund string, err error) {
	if !privileged {
		hoursLeft := x.Start.Sub(s.now()).Hours()
		feePercent = domain.FeePercentFor(h.EffectivePolicy(), hoursLeft)
	}
	if fee, err = domain.FeeAmount(x.TotalAmount, feePercent, x.Currency); err != nil {
		return 0, "", "", err
	}
	totalMinor, err := domain.MinorUnits(x.TotalAmount, x.Currency)
	if err != nil {
		return 0, "", "", err
	}
	feeMinor, _ := domain.MinorUnits(fee, x.Currency)
	return feePercent, fee, domain.FromMinorUnits(totalMinor-feeMinor, x.Currency), nil
}

func (s *ReservationService) CancellationQuote(ctx context.Context, actor inbound.Principal, id int64) (inbound.CancellationQuote, error) {
	x, privileged, err := s.load(ctx, actor, id)
	if err != nil {
		return inbound.CancellationQuote{}, err
	}
	x = s.syncGateway(ctx, x)
	h, err := s.hotels.Get(ctx, x.HotelID)
	if err != nil {
		return inbound.CancellationQuote{}, err
	}
	q := inbound.CancellationQuote{Status: x.Status, HoursLeft: x.Start.Sub(s.now()).Hours(), Policy: h.EffectivePolicy(), Fee: "0.0000", Refund: "0.0000"}
	switch x.Status {
	case domain.ReservationPending:
		q.Cancellable = true
	case domain.ReservationPaid:
		pct, fee, refund, err := s.cancellationTerms(x, h, privileged)
		if err != nil {
			return q, err
		}
		q.FeePercent, q.Fee, q.Refund = pct, fee, refund
		q.Cancellable = pct < 100
		if !q.Cancellable {
			q.Reason = "the reservation is non-refundable this close to check-in; please contact the hotel"
		}
	default:
		q.Reason = "the reservation is " + x.Status.Name()
	}
	return q, nil
}

func (s *ReservationService) Cancel(ctx context.Context, actor inbound.Principal, id int64) (domain.Reservation, error) {
	x, privileged, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	x = s.syncGateway(ctx, x)
	note := "cancelled by the guest"
	if privileged {
		note = "cancelled by the hotel"
	}

	switch x.Status {
	case domain.ReservationPending:
		if err := s.rail.refund(ctx, x.PaymentMethod, x.PaymentRef, fmt.Sprintf("reservation #%d cancelled", x.ID), true); err != nil {
			return domain.Reservation{}, err
		}
		done, err := s.transition(ctx, outbound.TransitionParams{ID: x.ID, From: []domain.ReservationStatus{domain.ReservationPending},
			To: domain.ReservationCanceled, Note: note, Actor: actor.UserID})
		if err == nil {
			s.notifyCancelled(ctx, done, privileged)
		}
		return done, err

	case domain.ReservationPaid:
		h, err := s.hotels.Get(ctx, x.HotelID)
		if err != nil {
			return domain.Reservation{}, err
		}
		pct, fee, refund, err := s.cancellationTerms(x, h, privileged)
		if err != nil {
			return domain.Reservation{}, err
		}
		if pct >= 100 {
			return domain.Reservation{}, fmt.Errorf("%w: the reservation is non-refundable this close to check-in; please contact the hotel", domain.ErrConflict)
		}
		fee, refund, note, err = s.payBack(ctx, x, h, fee, refund, note)
		if err != nil {
			return domain.Reservation{}, err
		}
		done, err := s.transition(ctx, outbound.TransitionParams{ID: x.ID, From: []domain.ReservationStatus{domain.ReservationPaid},
			To: domain.ReservationRefunded, Note: note, Actor: actor.UserID, Fee: fee, Refund: refund})
		if err == nil {
			s.notifyCancelled(ctx, done, privileged)
		}
		return done, err
	}
	return x, nil
}

func (s *ReservationService) payBack(ctx context.Context, x domain.Reservation, h domain.Hotel, fee, refund, note string) (string, string, string, error) {
	reason := fmt.Sprintf("reservation #%d cancelled", x.ID)
	feeMinor, _ := domain.MinorUnits(fee, x.Currency)
	refundMinor, _ := domain.MinorUnits(refund, x.Currency)

	if feeMinor == 0 {
		return fee, refund, note, s.rail.refund(ctx, x.PaymentMethod, x.PaymentRef, reason, false)
	}
	switch x.PaymentMethod {
	case domain.MethodGateway:
		if err := s.rail.refundGatewayPartial(ctx, x.PaymentRef, refundMinor); err != nil {
			return "", "", note, err
		}
	case domain.MethodWallet:
		if err := s.rail.refund(ctx, x.PaymentMethod, x.PaymentRef, reason, false); err != nil {
			return "", "", note, err
		}
		_, err := s.rail.chargeWallet(ctx, x.GuestID, h.WalletID, x.Currency, feeMinor,
			fmt.Sprintf("reservation-%d-fee", x.ID), fmt.Sprintf("Cancellation fee for reservation #%d", x.ID))
		if errors.Is(err, domain.ErrInsufficientFunds) {
			s.opts.Log.Warn("cancellation fee not collected", "reservation", x.ID, "fee", fee)
			return "0.0000", domain.FromMinorUnits(feeMinor+refundMinor, x.Currency), note + "; the cancellation fee could not be collected", nil
		}
		if err != nil {
			return "", "", note, err
		}
	}
	return fee, refund, note, nil
}

func (s *ReservationService) Reject(ctx context.Context, actor inbound.Principal, id int64, reason string) (domain.Reservation, error) {
	x, privileged, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	if !privileged {
		return domain.Reservation{}, domain.ErrForbidden
	}
	if reason = trim(reason); reason == "" {
		return domain.Reservation{}, fmt.Errorf("%w: give the guest a reason", domain.ErrInvalid)
	}
	x = s.syncGateway(ctx, x)
	if x.Status != domain.ReservationPending {
		return domain.Reservation{}, fmt.Errorf("%w: reservation is %s; only a pending one can be rejected (cancel a paid one to refund it)", domain.ErrConflict, x.Status.Name())
	}
	if err := s.rail.refund(ctx, x.PaymentMethod, x.PaymentRef, fmt.Sprintf("reservation #%d rejected", x.ID), true); err != nil {
		return domain.Reservation{}, err
	}
	done, err := s.transition(ctx, outbound.TransitionParams{ID: x.ID, From: []domain.ReservationStatus{domain.ReservationPending},
		To: domain.ReservationRejected, Note: reason, Actor: actor.UserID})
	if err == nil {
		s.notifyRejected(ctx, done, reason)
	}
	return done, err
}

func (s *ReservationService) transition(ctx context.Context, p outbound.TransitionParams) (domain.Reservation, error) {
	x, err := s.reservations.Transition(ctx, p)
	if errors.Is(err, domain.ErrConflict) {
		return domain.Reservation{}, fmt.Errorf("%w: the reservation changed while you were acting on it; look at it again", domain.ErrConflict)
	}
	if err == nil {
		s.roomsFreed(ctx, x.RoomTypeID)
	}
	return x, err
}

func (s *ReservationService) roomsFreed(ctx context.Context, roomTypeID int64) {
	s.opts.SoldOut.Release(roomTypeID)
	if s.opts.Promoter != nil {
		if _, err := s.opts.Promoter.Promote(context.WithoutCancel(ctx), roomTypeID); err != nil {
			s.opts.Log.Warn("waiting list not offered the freed rooms", "room_type", roomTypeID, "err", err)
		}
	}
}

func (s *ReservationService) ReleaseExpired(ctx context.Context) (int, error) {
	now := s.now()
	expired, err := s.reservations.ListExpired(ctx, now, expireBatch)
	if err != nil {
		return 0, err
	}
	released := 0
	for _, x := range expired {
		if x.PaymentMethod == domain.MethodGateway && x.PaymentRef != "" {
			switch s.rail.judgeStale(ctx, x.PaymentRef, x.ExpiresAt, s.opts.HoldTTL, now) {
			case staleKeep:
				continue
			case stalePaid:
				amount, _ := domain.MinorUnits(x.TotalAmount, x.Currency)
				if _, err := s.confirm(ctx, x, amount, domain.MethodGateway, x.PaymentRef); err != nil {
					s.opts.Log.Error("confirm captured payment", "reservation", x.ID, "err", err)
				}
				continue
			}
		}
		_, err := s.reservations.Transition(ctx, outbound.TransitionParams{ID: x.ID, From: []domain.ReservationStatus{domain.ReservationPending},
			To: domain.ReservationCanceled, Note: "hold expired"})
		switch {
		case err == nil:
			s.roomsFreed(ctx, x.RoomTypeID)
			s.notifyExpired(ctx, x)
			released++
		case errors.Is(err, domain.ErrConflict): 
		default:
			s.opts.Log.Error("expire reservation", "reservation", x.ID, "err", err)
		}
	}
	return released, nil
}
