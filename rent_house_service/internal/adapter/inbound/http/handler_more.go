package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

func today() time.Time { return time.Now().UTC().Truncate(24 * time.Hour) }

// ---- rates (the manager prices each rental model)

func (h *Handler) listRates(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	rs, err := h.homestays.Rates(r.Context(), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toRateResponses(rs))
}

func (h *Handler) setRate(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	var req rateRequest
	if !h.decode(w, r, &req) {
		return
	}
	active := req.Active == nil || *req.Active
	err := h.homestays.SetRate(r.Context(), mustPrincipal(r), domain.Rate{
		HomestayID: id, Model: domain.RentalModel(r.PathValue("model")), Price: req.Price.String(),
		Currency: req.Currency, MinPeriods: req.MinPeriods, Active: active,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteRate(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	if err := h.homestays.DeleteRate(r.Context(), mustPrincipal(r), id, domain.RentalModel(r.PathValue("model"))); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- leases

func (h *Handler) startLease(w http.ResponseWriter, r *http.Request) {
	var req leaseRequest
	if !h.decode(w, r, &req) {
		return
	}
	start, err := parseDate("start_date", req.StartDate)
	if err != nil {
		h.fail(w, err)
		return
	}
	if k := r.Header.Get("Idempotency-Key"); k != "" {
		req.RequestID = k
	}
	l, err := h.leases.Start(r.Context(), mustPrincipal(r), inbound.LeaseCommand{
		HomestayID: req.HomestayID, Model: domain.RentalModel(req.Model), Start: start, Periods: req.Periods,
		BillingDay: req.BillingDay, Note: req.Note, RequestID: req.RequestID,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toLeaseResponse(l, today()))
}

func (h *Handler) getLease(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	l, err := h.leases.Get(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toLeaseResponse(l, today()))
}

func (h *Handler) listLeases(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.leases.List(r.Context(), mustPrincipal(r), int64(atoi(q.Get("user_id"))), q.Get("as") == "host", atoi(q.Get("limit")), atoi(q.Get("offset")))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]leaseResponse, len(list))
	for i, l := range list {
		out[i] = toLeaseResponse(l, today())
	}
	h.write(w, http.StatusOK, out)
}

func (h *Handler) payInvoice(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	period := atoi(r.PathValue("period"))
	if period < 1 {
		h.fail(w, domain.ErrInvalid)
		return
	}
	var req payRequest
	if !h.decode(w, r, &req) {
		return
	}
	l, err := h.leases.PayInvoice(r.Context(), mustPrincipal(r), id, period-1, inbound.PayCommand{
		Method: domain.PaymentMethod(strings.ToLower(req.Method)), Provider: req.Provider,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toLeaseResponse(l, today()))
}

func (h *Handler) cancelLease(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	l, err := h.leases.Cancel(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toLeaseResponse(l, today()))
}

func (h *Handler) terminateLease(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	l, err := h.leases.Terminate(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toLeaseResponse(l, today()))
}

func (h *Handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := inbound.InvoiceQuery{Limit: atoi(q.Get("limit")), Offset: atoi(q.Get("offset"))}
	switch q.Get("status") {
	case "":
	case "unpaid":
		f.Status = domain.InvoiceUnpaid
	case "paid":
		f.Status = domain.InvoicePaid
	case "void":
		f.Status = domain.InvoiceVoid
	case "refunded":
		f.Status = domain.InvoiceRefunded
	default:
		h.fail(w, domain.ErrInvalid)
		return
	}
	if v := q.Get("due_before"); v != "" {
		d, err := parseDate("due_before", v)
		if err != nil {
			h.fail(w, err)
			return
		}
		f.DueBefore = d
	}
	list, err := h.leases.Invoices(r.Context(), mustPrincipal(r), f)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]invoiceResponse, len(list))
	for i, inv := range list {
		out[i] = toInvoiceResponse(inv, today())
	}
	h.write(w, http.StatusOK, out)
}

// ---- reviews

func (h *Handler) listReviews(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	list, err := h.reviews.List(r.Context(), id, atoi(q.Get("limit")), atoi(q.Get("offset")))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]reviewResponse, len(list))
	for i, rv := range list {
		out[i] = toReviewResponse(rv)
	}
	h.write(w, http.StatusOK, out)
}

func (h *Handler) createReview(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	var req reviewRequest
	if !h.decode(w, r, &req) {
		return
	}
	rv, err := h.reviews.Create(r.Context(), mustPrincipal(r), inbound.ReviewCommand{
		HomestayID: id, BookingID: req.BookingID, LeaseID: req.LeaseID, Rating: req.Rating, Comment: req.Comment,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toReviewResponse(rv))
}

func toReviewResponse(rv domain.Review) reviewResponse {
	return reviewResponse{ID: rv.ID, HomestayID: rv.HomestayID, UserID: rv.UserID, BookingID: rv.BookingID,
		LeaseID: rv.LeaseID, Rating: rv.Rating, Comment: rv.Comment, CreatedAt: rv.CreatedAt}
}

// ---- wishlist and loyalty

func (h *Handler) addWishlist(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
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
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	if err := h.wishlist.Remove(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listWishlist(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.wishlist.List(r.Context(), mustPrincipal(r), atoi(q.Get("limit")), atoi(q.Get("offset")))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]homestayResponse, len(list))
	for i, x := range list {
		out[i] = toHomestayResponse(x)
	}
	h.write(w, http.StatusOK, out)
}

func (h *Handler) loyaltyStatus(w http.ResponseWriter, r *http.Request) {
	s, err := h.loyalty.Status(r.Context(), mustPrincipal(r))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, loyaltyResponse{Points: s.Points, Tier: s.Tier, TierName: s.TierName})
}

// ---- registration: any signed-in user registers a homestay, hotel or rental house and owns it

func (h *Handler) myHomestays(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.homestays.Mine(r.Context(), mustPrincipal(r), atoi(q.Get("limit")), atoi(q.Get("offset")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeHomestays(w, list)
}

// reviewQueue is the admin's list of registrations by status (default: waiting for approval).
func (h *Handler) reviewQueue(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := map[string]domain.HomestayStatus{"": 0, "pending": domain.HomestayPending, "rejected": domain.HomestayRejected,
		"active": domain.HomestayActive, "inactive": domain.HomestayInactive}
	st, ok := status[q.Get("status")]
	if !ok {
		h.fail(w, domain.ErrInvalid)
		return
	}
	list, err := h.homestays.ForReview(r.Context(), mustPrincipal(r), st, atoi(q.Get("limit")), atoi(q.Get("offset")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeHomestays(w, list)
}

func (h *Handler) approveHomestay(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	x, err := h.homestays.Approve(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toHomestayResponse(x))
}

func (h *Handler) rejectHomestay(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	var req rejectRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.homestays.Reject(r.Context(), mustPrincipal(r), id, req.Reason)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toHomestayResponse(x))
}

func (h *Handler) writeHomestays(w http.ResponseWriter, list []domain.Homestay) {
	out := make([]homestayResponse, len(list))
	for i, x := range list {
		out[i] = toHomestayResponse(x)
	}
	h.write(w, http.StatusOK, out)
}
