package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

const maxBodyBytes = 1 << 20

type Handler struct {
	auth          inbound.AuthUseCase
	events        inbound.EventUseCase
	catalog       inbound.CatalogUseCase
	orders        inbound.OrderUseCase
	gate          inbound.GateUseCase
	tickets       inbound.TicketUseCase
	staff         inbound.StaffUseCase
	waitlist      inbound.WaitlistUseCase
	resale        inbound.ResaleUseCase
	venues        inbound.VenueUseCase
	documents     inbound.DocumentUseCase
	wishlist      inbound.WishlistUseCase
	notifications inbound.NotificationUseCase
	log           *slog.Logger
}

func NewHandler(a inbound.AuthUseCase, e inbound.EventUseCase, c inbound.CatalogUseCase, o inbound.OrderUseCase,
	g inbound.GateUseCase, t inbound.TicketUseCase, st inbound.StaffUseCase, wl inbound.WaitlistUseCase, rs inbound.ResaleUseCase, vn inbound.VenueUseCase, dc inbound.DocumentUseCase,
	w inbound.WishlistUseCase, n inbound.NotificationUseCase, log *slog.Logger) *Handler {
	return &Handler{auth: a, events: e, catalog: c, orders: o, gate: g, tickets: t, staff: st, waitlist: wl, resale: rs, venues: vn, documents: dc, wishlist: w, notifications: n, log: log}
}

// ---- plumbing

func (h *Handler) write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.log.Warn("write response", "err", err)
	}
}

func (h *Handler) decode(w http.ResponseWriter, r *http.Request, into any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(into); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			h.write(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body is too large"})
			return false
		}
		h.fail(w, fmt.Errorf("%w: bad JSON body: %v", domain.ErrInvalid, err))
		return false
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		h.fail(w, fmt.Errorf("%w: bad JSON body: unexpected data after the object", domain.ErrInvalid))
		return false
	}
	return true
}

func (h *Handler) pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		h.fail(w, fmt.Errorf("%w: %s must be a positive integer", domain.ErrInvalid, name))
		return 0, false
	}
	return id, true
}

func (h *Handler) fail(w http.ResponseWriter, err error) {
	status, msg := http.StatusInternalServerError, "internal error"
	switch {
	case errors.Is(err, domain.ErrInvalid):
		status, msg = http.StatusBadRequest, cleanMsg(err)
	case errors.Is(err, domain.ErrUnauthorized):
		status, msg = http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, domain.ErrForbidden):
		status, msg = http.StatusForbidden, cleanMsg(err)
	case errors.Is(err, domain.ErrNotFound):
		status, msg = http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrSoldOut):
		status, msg = http.StatusConflict, cleanMsg(err)
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyCheckedIn):
		status, msg = http.StatusConflict, cleanMsg(err)
	case errors.Is(err, domain.ErrBusy):
		w.Header().Set("Retry-After", "1")
		status, msg = http.StatusTooManyRequests, "too many requests right now, try again in a moment"
	case errors.Is(err, domain.ErrInsufficientFunds):
		status, msg = http.StatusPaymentRequired, "insufficient wallet balance"
	case errors.Is(err, domain.ErrPaymentFailed):
		status, msg = http.StatusPaymentRequired, "payment was declined by the provider"
	case errors.Is(err, domain.ErrUpstreamUnavailable):
		status, msg = http.StatusServiceUnavailable, "dependent service unavailable"
	}
	if status == http.StatusInternalServerError || status == http.StatusServiceUnavailable {
		h.log.Error("request failed", "err", err)
	}
	h.write(w, status, map[string]string{"error": msg})
}

// cleanMsg drops the sentinel prefix ("invalid input: ", "conflict: ", "forbidden: ") from wrapped messages.
func cleanMsg(err error) string {
	m := err.Error()
	for _, p := range []string{"invalid input: ", "conflict: ", "forbidden: "} {
		if i := strings.Index(m, p); i >= 0 {
			return m[i+len(p):]
		}
	}
	return m
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func (h *Handler) cursor(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := decodeIDCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		h.fail(w, err)
		return 0, false
	}
	return id, true
}

func parseTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("%w: %q is not a date (use 2006-01-02 or RFC 3339)", domain.ErrInvalid, s)
}

func parsePrice(s string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return nil, fmt.Errorf("%w: prices are whole numbers of minor units", domain.ErrInvalid)
	}
	return &n, nil
}

// ---- catalog

func (h *Handler) searchEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := domain.EventFilter{Query: q.Get("q"), City: q.Get("city"), Category: q.Get("category"), Sort: q.Get("sort"),
		Featured: q.Get("featured") == "true", Limit: atoi(q.Get("limit")), Offset: atoi(q.Get("offset"))}
	var err error
	if f.From, err = parseTime(q.Get("from")); err != nil {
		h.fail(w, err)
		return
	}
	if f.To, err = parseTime(q.Get("to")); err != nil {
		h.fail(w, err)
		return
	}
	if f.MinPrice, err = parsePrice(q.Get("min_price")); err != nil {
		h.fail(w, err)
		return
	}
	if f.MaxPrice, err = parsePrice(q.Get("max_price")); err != nil {
		h.fail(w, err)
		return
	}
	rows, err := h.catalog.Search(r.Context(), f)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]eventResponse, len(rows))
	for i, e := range rows {
		out[i] = toEventResponse(e)
	}
	h.write(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) categories(w http.ResponseWriter, _ *http.Request) {
	h.write(w, http.StatusOK, map[string]any{"items": domain.Categories})
}

func (h *Handler) getEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var viewer *inbound.Principal
	if p, ok := principalFrom(r.Context()); ok {
		viewer = &p
	}
	e, err := h.events.View(r.Context(), viewer, id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toEventResponse(e))
}

func (h *Handler) listSeats(w http.ResponseWriter, r *http.Request) {
	eventID, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	typeID, ok := h.pathID(w, r, "type")
	if !ok {
		return
	}
	m, err := h.catalog.Seats(r.Context(), eventID, typeID)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]seatResponse, len(m.Seats))
	for i, s := range m.Seats {
		out[i] = toSeatResponse(s)
	}
	h.write(w, http.StatusOK, map[string]any{"layout": m.Layout, "items": out})
}

// ---- organizer and admin: events

func (h *Handler) createEvent(w http.ResponseWriter, r *http.Request) {
	var req eventRequest
	if !h.decode(w, r, &req) {
		return
	}
	e, err := h.events.Create(r.Context(), mustPrincipal(r), req.input())
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toEventResponse(e))
}

func (h *Handler) updateEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req eventRequest
	if !h.decode(w, r, &req) {
		return
	}
	e, err := h.events.Update(r.Context(), mustPrincipal(r), id, req.input())
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toEventResponse(e))
}

func (h *Handler) eventStep(step func(p inbound.Principal, r *http.Request, id int64) (domain.Event, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.pathID(w, r, "id")
		if !ok {
			return
		}
		e, err := step(mustPrincipal(r), r, id)
		if err != nil {
			h.fail(w, err)
			return
		}
		h.write(w, http.StatusOK, toEventResponse(e))
	}
}

type reasonRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) cancelEvent(w http.ResponseWriter, r *http.Request) {
	var req reasonRequest
	if r.ContentLength != 0 && !h.decode(w, r, &req) {
		return
	}
	h.eventStep(func(p inbound.Principal, r *http.Request, id int64) (domain.Event, error) {
		return h.events.Cancel(r.Context(), p, id, req.Reason)
	})(w, r)
}

func (h *Handler) rejectEvent(w http.ResponseWriter, r *http.Request) {
	var req reasonRequest
	if !h.decode(w, r, &req) {
		return
	}
	h.eventStep(func(p inbound.Principal, r *http.Request, id int64) (domain.Event, error) {
		return h.events.Reject(r.Context(), p, id, req.Reason)
	})(w, r)
}

func (h *Handler) setFeatured(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Featured bool `json:"featured"`
	}
	if !h.decode(w, r, &req) {
		return
	}
	h.eventStep(func(p inbound.Principal, r *http.Request, id int64) (domain.Event, error) {
		return h.events.SetFeatured(r.Context(), p, id, req.Featured)
	})(w, r)
}

func (h *Handler) myEvents(w http.ResponseWriter, r *http.Request) {
	h.eventPage(w, r, h.events.ListMine)
}

func (h *Handler) reviewQueue(w http.ResponseWriter, r *http.Request) {
	h.eventPage(w, r, h.events.ReviewQueue)
}

func (h *Handler) eventPage(w http.ResponseWriter, r *http.Request, list func(ctx context.Context, p inbound.Principal, after int64, limit int) (domain.Page[domain.Event], error)) {
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	p, err := list(r.Context(), mustPrincipal(r), after, atoi(r.URL.Query().Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, toEventResponse))
}

func (h *Handler) createTicketType(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req ticketTypeRequest
	if !h.decode(w, r, &req) {
		return
	}
	t, err := h.events.CreateTicketType(r.Context(), mustPrincipal(r), id, req.input())
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toTypeResponse(t, time.Now()))
}

func (h *Handler) updateTicketType(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	typeID, ok := h.pathID(w, r, "type")
	if !ok {
		return
	}
	var req ticketTypeRequest
	if !h.decode(w, r, &req) {
		return
	}
	t, err := h.events.UpdateTicketType(r.Context(), mustPrincipal(r), id, typeID, req.input())
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toTypeResponse(t, time.Now()))
}

func (h *Handler) generateSeats(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	typeID, ok := h.pathID(w, r, "type")
	if !ok {
		return
	}
	var req seatLayoutRequest
	if !h.decode(w, r, &req) {
		return
	}
	n, err := h.events.GenerateSeats(r.Context(), mustPrincipal(r), id, typeID, inbound.SeatLayout{
		Section: req.Section, Rows: req.Rows, SeatsPerRow: req.SeatsPerRow, FirstRow: req.FirstRow, OffsetX: req.OffsetX, OffsetY: req.OffsetY})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, map[string]int{"seats_added": n})
}

func (h *Handler) listPromotions(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	rows, err := h.events.ListPromotions(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]promotionResponse, len(rows))
	for i, p := range rows {
		out[i] = toPromotionResponse(p)
	}
	h.write(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) createPromotion(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req promotionRequest
	if !h.decode(w, r, &req) {
		return
	}
	p, err := h.events.CreatePromotion(r.Context(), mustPrincipal(r), id, req.input())
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toPromotionResponse(p))
}

func (h *Handler) updatePromotion(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	promoID, ok := h.pathID(w, r, "promo")
	if !ok {
		return
	}
	var req promotionRequest
	if !h.decode(w, r, &req) {
		return
	}
	p, err := h.events.UpdatePromotion(r.Context(), mustPrincipal(r), id, promoID, req.input())
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPromotionResponse(p))
}

func (h *Handler) eventReport(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	rep, err := h.events.Report(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toReportResponse(rep))
}

func (h *Handler) attendees(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	p, err := h.events.Attendees(r.Context(), mustPrincipal(r), id, after, atoi(r.URL.Query().Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, func(a domain.Attendee) attendeeResponse {
		return attendeeResponse{ticketResponse: toTicketResponse(a.Ticket), BuyerName: a.BuyerName, BuyerEmail: a.BuyerEmail}
	}))
}

func (h *Handler) checkIn(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if !h.decode(w, r, &req) {
		return
	}
	t, err := h.gate.CheckIn(r.Context(), mustPrincipal(r), id, req.Code)
	if errors.Is(err, domain.ErrAlreadyCheckedIn) && t.ID != 0 {
		// The gate needs to see who was admitted and when, to tell a genuine second scan from a copied QR code.
		h.write(w, http.StatusConflict, map[string]any{"error": "ticket already checked in", "ticket": toTicketResponse(t)})
		return
	}
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toTicketResponse(t))
}

// ---- orders

func (h *Handler) reserve(w http.ResponseWriter, r *http.Request) {
	var req reserveRequest
	if !h.decode(w, r, &req) {
		return
	}
	cmd := req.command()
	if h.orders.SoldOut(r.Context(), cmd) {
		h.fail(w, &domain.SoldOutError{})
		return
	}
	r, err := h.authenticate(r)
	if err != nil {
		h.fail(w, err)
		return
	}
	x, err := h.orders.Reserve(r.Context(), mustPrincipal(r), cmd)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toOrderResponse(x))
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	x, err := h.orders.Get(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toOrderResponse(x))
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	status := domain.ParseOrderStatus(q.Get("status"))
	if q.Get("status") != "" && status == 0 {
		h.fail(w, fmt.Errorf("%w: unknown status", domain.ErrInvalid))
		return
	}
	p, err := h.orders.List(r.Context(), mustPrincipal(r), inbound.OrderListQuery{
		EventID: int64(atoi(q.Get("event_id"))), Status: status, AfterID: after, Limit: atoi(q.Get("limit"))})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, toOrderResponse))
}

func (h *Handler) listEventOrders(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	status := domain.ParseOrderStatus(r.URL.Query().Get("status"))
	p, err := h.orders.List(r.Context(), mustPrincipal(r), inbound.OrderListQuery{
		EventID: id, Status: status, AfterID: after, Limit: atoi(r.URL.Query().Get("limit"))})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, toOrderResponse))
}

func (h *Handler) payOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req payRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.orders.Pay(r.Context(), mustPrincipal(r), id, inbound.PayCommand{Method: domain.PaymentMethod(req.Method), Provider: req.Provider})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toOrderResponse(x))
}

func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	x, err := h.orders.Cancel(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toOrderResponse(x))
}

// ---- wishlist

func (h *Handler) listWishlist(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, err := h.wishlist.List(r.Context(), mustPrincipal(r), atoi(q.Get("limit")), atoi(q.Get("offset")))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]eventResponse, len(rows))
	for i, e := range rows {
		out[i] = toEventResponse(e)
	}
	h.write(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) addWishlist(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.wishlist.Add(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) removeWishlist(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.wishlist.Remove(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- notifications

func (h *Handler) listNotifications(w http.ResponseWriter, r *http.Request) {
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	p, err := h.notifications.List(r.Context(), mustPrincipal(r), q.Get("unread") == "true", after, atoi(q.Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, func(n domain.Notification) notificationResponse {
		return notificationResponse{ID: n.ID, Kind: n.Kind, Title: n.Title, Body: n.Body, OrderID: n.OrderID, EventID: n.EventID, CreatedAt: n.CreatedAt, ReadAt: n.ReadAt}
	}))
}

func (h *Handler) unreadCount(w http.ResponseWriter, r *http.Request) {
	n, err := h.notifications.UnreadCount(r.Context(), mustPrincipal(r))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, map[string]int{"unread": n})
}

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.notifications.MarkRead(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) markAllRead(w http.ResponseWriter, r *http.Request) {
	n, err := h.notifications.MarkAllRead(r.Context(), mustPrincipal(r))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, map[string]int{"marked": n})
}
