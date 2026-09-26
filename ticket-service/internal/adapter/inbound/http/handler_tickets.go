package http

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

func (h *Handler) listTickets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	status := domain.ParseTicketStatus(q.Get("status"))
	if q.Get("status") != "" && status == 0 {
		h.fail(w, fmt.Errorf("%w: unknown status", domain.ErrInvalid))
		return
	}
	p, err := h.tickets.List(r.Context(), mustPrincipal(r), inbound.TicketListQuery{
		EventID: int64(atoi(q.Get("event_id"))), Status: status, AfterID: after, Limit: atoi(q.Get("limit"))})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, toTicketResponse))
}

func (h *Handler) ticketStep(step func(r *http.Request, id int64) (domain.Ticket, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.pathID(w, r, "id")
		if !ok {
			return
		}
		t, err := step(r, id)
		if err != nil {
			h.fail(w, err)
			return
		}
		h.write(w, http.StatusOK, toTicketResponse(t))
	}
}

func (h *Handler) getTicket(w http.ResponseWriter, r *http.Request) {
	h.ticketStep(func(r *http.Request, id int64) (domain.Ticket, error) {
		return h.tickets.Get(r.Context(), mustPrincipal(r), id)
	})(w, r)
}

func (h *Handler) reissueTicket(w http.ResponseWriter, r *http.Request) {
	h.ticketStep(func(r *http.Request, id int64) (domain.Ticket, error) {
		return h.tickets.Reissue(r.Context(), mustPrincipal(r), id)
	})(w, r)
}

func (h *Handler) revertCheckIn(w http.ResponseWriter, r *http.Request) {
	h.ticketStep(func(r *http.Request, id int64) (domain.Ticket, error) {
		return h.tickets.RevertCheckIn(r.Context(), mustPrincipal(r), id)
	})(w, r)
}

func (h *Handler) setHolder(w http.ResponseWriter, r *http.Request) {
	var req holderRequest
	if !h.decode(w, r, &req) {
		return
	}
	h.ticketStep(func(r *http.Request, id int64) (domain.Ticket, error) {
		return h.tickets.SetHolder(r.Context(), mustPrincipal(r), id, inbound.HolderInput{Name: req.Name, Email: req.Email})
	})(w, r)
}

func (h *Handler) voidTicket(w http.ResponseWriter, r *http.Request) {
	var req reasonRequest
	if !h.decode(w, r, &req) {
		return
	}
	h.ticketStep(func(r *http.Request, id int64) (domain.Ticket, error) {
		return h.tickets.Void(r.Context(), mustPrincipal(r), id, req.Reason)
	})(w, r)
}

func (h *Handler) ticketHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	rows, err := h.tickets.History(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]ticketEventResponse, len(rows))
	for i, e := range rows {
		out[i] = toTicketEventResponse(e)
	}
	h.write(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) offerTicket(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req transferRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.tickets.Offer(r.Context(), mustPrincipal(r), id, inbound.TransferCommand{ToUserID: req.ToUserID, ToEmail: req.ToEmail, Message: req.Message})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toTransferResponse(x))
}

func (h *Handler) listTransfers(w http.ResponseWriter, r *http.Request) {
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	incoming := q.Get("direction") == "incoming"
	if d := q.Get("direction"); d != "" && d != "incoming" && d != "outgoing" {
		h.fail(w, fmt.Errorf("%w: direction must be incoming or outgoing", domain.ErrInvalid))
		return
	}
	p, err := h.tickets.Transfers(r.Context(), mustPrincipal(r), incoming, after, atoi(q.Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, toTransferResponse))
}

func (h *Handler) acceptTransfer(w http.ResponseWriter, r *http.Request) {
	h.ticketStep(func(r *http.Request, id int64) (domain.Ticket, error) {
		return h.tickets.AcceptTransfer(r.Context(), mustPrincipal(r), id)
	})(w, r)
}

func (h *Handler) transferStep(step func(r *http.Request, id int64) (domain.Transfer, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.pathID(w, r, "id")
		if !ok {
			return
		}
		x, err := step(r, id)
		if err != nil {
			h.fail(w, err)
			return
		}
		h.write(w, http.StatusOK, toTransferResponse(x))
	}
}

func (h *Handler) declineTransfer(w http.ResponseWriter, r *http.Request) {
	h.transferStep(func(r *http.Request, id int64) (domain.Transfer, error) {
		return h.tickets.DeclineTransfer(r.Context(), mustPrincipal(r), id)
	})(w, r)
}

func (h *Handler) cancelTransfer(w http.ResponseWriter, r *http.Request) {
	h.transferStep(func(r *http.Request, id int64) (domain.Transfer, error) {
		return h.tickets.CancelTransfer(r.Context(), mustPrincipal(r), id)
	})(w, r)
}

func (h *Handler) lookupTicket(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	t, err := h.tickets.Lookup(r.Context(), mustPrincipal(r), id, r.PathValue("code"))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toTicketResponse(t))
}

func (h *Handler) batchCheckIn(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req scanRequest
	if !h.decode(w, r, &req) {
		return
	}
	items := make([]inbound.ScanItem, len(req.Scans))
	for i, s := range req.Scans {
		items[i] = inbound.ScanItem{Code: s.Code, ScannedAt: s.ScannedAt}
	}
	res, err := h.gate.BatchCheckIn(r.Context(), mustPrincipal(r), id, items)
	out := make([]scanResponse, len(res))
	admitted := 0
	for i, x := range res {
		out[i] = scanResponse{Code: x.Code, Result: x.Result}
		if x.Ticket.ID != 0 {
			t := toTicketResponse(x.Ticket)
			t.Code = ""
			out[i].Ticket = &t
		}
		if x.Result == "admitted" {
			admitted++
		}
	}
	if err != nil && len(res) == 0 {
		h.fail(w, err)
		return
	}
	status := http.StatusOK
	body := map[string]any{"results": out, "admitted": admitted}
	if err != nil {
		status = http.StatusServiceUnavailable
		body["error"] = "stopped early; send the remaining scans again"
		h.log.Error("batch check-in stopped", "err", err)
	}
	h.write(w, status, body)
}

// ---- staff, invitations, waiting lists

func (h *Handler) addStaff(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req staffRequest
	if !h.decode(w, r, &req) {
		return
	}
	if err := h.staff.Add(r.Context(), mustPrincipal(r), id, req.UserID); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) removeStaff(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	userID, ok := h.pathID(w, r, "user")
	if !ok {
		return
	}
	if err := h.staff.Remove(r.Context(), mustPrincipal(r), id, userID); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listStaff(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	rows, err := h.staff.List(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]staffResponse, len(rows))
	for i, m := range rows {
		out[i] = staffResponse{UserID: m.UserID, AddedBy: m.AddedBy, CreatedAt: m.CreatedAt}
	}
	h.write(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) staffEvents(w http.ResponseWriter, r *http.Request) {
	rows, err := h.staff.Events(r.Context(), mustPrincipal(r))
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

func (h *Handler) invite(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req inviteRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.orders.Invite(r.Context(), mustPrincipal(r), id, inbound.InviteCommand{TicketTypeID: req.TicketTypeID, Quantity: req.Quantity,
		SeatIDs: req.SeatIDs, UserID: req.UserID, Name: req.Name, Email: req.Email, Note: req.Note})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toOrderResponse(x))
}

func (h *Handler) joinWaitlist(w http.ResponseWriter, r *http.Request) {
	eventID, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	typeID, ok := h.pathID(w, r, "type")
	if !ok {
		return
	}
	if err := h.waitlist.Join(r.Context(), mustPrincipal(r), eventID, typeID); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) leaveWaitlist(w http.ResponseWriter, r *http.Request) {
	typeID, ok := h.pathID(w, r, "type")
	if !ok {
		return
	}
	if err := h.waitlist.Leave(r.Context(), mustPrincipal(r), typeID); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) myWaitlist(w http.ResponseWriter, r *http.Request) {
	rows, err := h.waitlist.Mine(r.Context(), mustPrincipal(r))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]waitlistResponse, len(rows))
	for i, x := range rows {
		out[i] = waitlistResponse{TicketTypeID: x.TicketTypeID, EventID: x.EventID, EventTitle: x.EventTitle, TypeName: x.TypeName, CreatedAt: x.CreatedAt}
	}
	h.write(w, http.StatusOK, map[string]any{"items": out})
}

// ---- exports

const maxCSVRows = 100_000

func csvCell(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

func (h *Handler) attendeesCSV(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	p := mustPrincipal(r)
	first, err := h.events.Attendees(r.Context(), p, id, 0, 100)
	if err != nil { // authorization is checked here, before any byte of the file is written
		h.fail(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="attendees-`+strconv.FormatInt(id, 10)+`.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"ticket_id", "order_id", "ticket_type", "seat", "status", "holder_name", "holder_email", "buyer_name", "buyer_email", "checked_in_at"})
	rows := 0
	for page := first; ; {
		for _, a := range page.Items {
			at := ""
			if a.CheckedInAt != nil {
				at = a.CheckedInAt.UTC().Format("2006-01-02T15:04:05Z")
			}
			_ = cw.Write([]string{strconv.FormatInt(a.ID, 10), strconv.FormatInt(a.OrderID, 10), csvCell(a.TypeName), csvCell(a.SeatLabel),
				a.Status.Name(), csvCell(a.HolderName), csvCell(a.HolderEmail), csvCell(a.BuyerName), csvCell(a.BuyerEmail), at})
			rows++
		}
		if page.NextID == 0 || rows >= maxCSVRows {
			break
		}
		if page, err = h.events.Attendees(r.Context(), p, id, page.NextID, 100); err != nil {
			h.log.Error("attendees export stopped", "err", err)
			break
		}
	}
	cw.Flush()
}

// ---- the internal API

func (h *Handler) internalAuth(key string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !keyMatches(r.Header.Get("X-Internal-Key"), key) {
			h.write(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		acting, _ := strconv.ParseInt(r.Header.Get("X-Acting-User"), 10, 64)
		next(w, r.WithContext(withPrincipal(r.Context(), inbound.Principal{UserID: max(acting, 0), Admin: true})))
	})
}

func (h *Handler) internalTicketByCode(w http.ResponseWriter, r *http.Request) {
	t, err := h.tickets.Lookup(r.Context(), mustPrincipal(r), int64(atoi(r.URL.Query().Get("event_id"))), r.PathValue("code"))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toTicketResponse(t))
}

func (h *Handler) internalUserTickets(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.pathID(w, r, "user")
	if !ok {
		return
	}
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	status := domain.ParseTicketStatus(q.Get("status"))
	if q.Get("status") != "" && status == 0 {
		h.fail(w, errors.New("invalid input: unknown status"))
		return
	}
	p, err := h.tickets.List(r.Context(), inbound.Principal{UserID: userID}, inbound.TicketListQuery{
		EventID: int64(atoi(q.Get("event_id"))), Status: status, AfterID: after, Limit: atoi(q.Get("limit"))})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, toTicketResponse))
}

func (h *Handler) setSeatMap(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	typeID, ok := h.pathID(w, r, "type")
	if !ok {
		return
	}
	var req seatMapRequest
	if !h.decode(w, r, &req) {
		return
	}
	seats := make([]domain.SeatPosition, len(req.Seats))
	for i, p := range req.Seats {
		seats[i] = domain.SeatPosition{Row: strings.ToUpper(strings.TrimSpace(p.Row)), Number: p.Number, X: p.X, Y: p.Y}
	}
	n, err := h.events.SetSeatMap(r.Context(), mustPrincipal(r), id, typeID, req.SeatMapLayout, seats)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, map[string]int{"seats_placed": n})
}

func (h *Handler) duplicateEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req duplicateRequest
	if !h.decode(w, r, &req) {
		return
	}
	e, err := h.events.Duplicate(r.Context(), mustPrincipal(r), id, inbound.DuplicateInput{Title: req.Title, StartsAt: req.StartsAt, EndsAt: req.EndsAt, CopyPromotions: req.CopyPromotions})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toEventResponse(e))
}

func (h *Handler) eventSeries(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	rows, err := h.catalog.Series(r.Context(), id)
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

// ---- resale

func (h *Handler) browseResale(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	p, err := h.resale.Browse(r.Context(), inbound.ResaleQuery{EventID: id, TypeID: int64(atoi(q.Get("type"))),
		CheapestFirst: q.Get("sort") == "price", AfterID: after, Limit: atoi(q.Get("limit"))})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, func(l domain.ResaleListing) resaleResponse { return toResaleResponse(l, false) }))
}

func (h *Handler) myResale(w http.ResponseWriter, r *http.Request) {
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	status := domain.ParseResaleStatus(q.Get("status"))
	if q.Get("status") != "" && status == 0 {
		h.fail(w, fmt.Errorf("%w: unknown status", domain.ErrInvalid))
		return
	}
	p, err := h.resale.Mine(r.Context(), mustPrincipal(r), status, after, atoi(q.Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, func(l domain.ResaleListing) resaleResponse { return toResaleResponse(l, true) }))
}

func (h *Handler) listForResale(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req resaleRequest
	if !h.decode(w, r, &req) {
		return
	}
	l, err := h.resale.List(r.Context(), mustPrincipal(r), id, req.Price)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toResaleResponse(l, true))
}

func (h *Handler) cancelResale(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	l, err := h.resale.Cancel(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toResaleResponse(l, true))
}

func (h *Handler) buyResale(w http.ResponseWriter, r *http.Request) {
	h.ticketStep(func(r *http.Request, id int64) (domain.Ticket, error) {
		return h.resale.Buy(r.Context(), mustPrincipal(r), id)
	})(w, r)
}

// ---- sessions (showtimes)

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req sessionRequest
	if !h.decode(w, r, &req) {
		return
	}
	sess, err := h.events.CreateSession(r.Context(), mustPrincipal(r), id, inbound.SessionInput{StartsAt: req.StartsAt, EndsAt: req.EndsAt, Label: req.Label, CopyFrom: req.CopyFrom})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toSessionResponse(sess))
}

func (h *Handler) updateSession(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	sid, ok := h.pathID(w, r, "session")
	if !ok {
		return
	}
	var req sessionRequest
	if !h.decode(w, r, &req) {
		return
	}
	sess, err := h.events.UpdateSession(r.Context(), mustPrincipal(r), id, sid, inbound.SessionInput{StartsAt: req.StartsAt, EndsAt: req.EndsAt, Label: req.Label})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toSessionResponse(sess))
}

func (h *Handler) deleteSession(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	sid, ok := h.pathID(w, r, "session")
	if !ok {
		return
	}
	if err := h.events.DeleteSession(r.Context(), mustPrincipal(r), id, sid); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) cancelSession(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	sid, ok := h.pathID(w, r, "session")
	if !ok {
		return
	}
	var req reasonRequest
	if !h.decode(w, r, &req) {
		return
	}
	sess, err := h.events.CancelSession(r.Context(), mustPrincipal(r), id, sid, req.Reason)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toSessionResponse(sess))
}

// listSessions is the showtimes of a published event (an organizer sees them on the event, whatever its status).
func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
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
	out := make([]sessionResponse, len(e.Sessions))
	for i, x := range e.Sessions {
		out[i] = toSessionResponse(x)
	}
	h.write(w, http.StatusOK, map[string]any{"items": out})
}
