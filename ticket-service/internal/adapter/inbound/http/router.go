package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

var errUnauthorized = domain.ErrUnauthorized

type Pinger func(ctx context.Context) error

func NewRouter(h *Handler, ping Pinger, maxInFlight int, internalKey string, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := ping(ctx); err != nil {
			log.Error("readiness", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	const v1 = "/api/v1"

	mux.HandleFunc("GET /editor/seat-map", serveEditor)

	mux.HandleFunc("GET "+v1+"/events", h.searchEvents)
	mux.HandleFunc("GET "+v1+"/categories", h.categories)
	mux.Handle("GET "+v1+"/events/{id}", h.optionalAuth(http.HandlerFunc(h.getEvent)))
	mux.HandleFunc("GET "+v1+"/events/{id}/ticket-types/{type}/seats", h.listSeats)
	mux.HandleFunc("GET "+v1+"/events/{id}/series", h.eventSeries)
	mux.Handle("GET "+v1+"/events/{id}/sessions", h.optionalAuth(http.HandlerFunc(h.listSessions)))
	mux.HandleFunc("GET "+v1+"/events/{id}/seat-map", h.eventSeatMap)
	mux.HandleFunc("GET "+v1+"/venue-templates", h.venueTemplates)
	mux.HandleFunc("GET "+v1+"/events/{id}/resale", h.browseResale)

	mux.Handle("POST "+v1+"/events", h.requireAuth(h.createEvent))
	mux.Handle("GET "+v1+"/my/events", h.requireAuth(h.myEvents))
	mux.Handle("PUT "+v1+"/events/{id}", h.requireAuth(h.updateEvent))
	mux.Handle("POST "+v1+"/events/{id}/submit", h.requireAuth(h.eventStep(func(p inbound.Principal, r *http.Request, id int64) (domain.Event, error) {
		return h.events.Submit(r.Context(), p, id)
	})))
	mux.Handle("POST "+v1+"/venues/preview", h.requireAuth(h.previewVenue))
	mux.Handle("POST "+v1+"/venues", h.requireAuth(h.createVenue))
	mux.Handle("GET "+v1+"/venues", h.requireAuth(h.listVenues))
	mux.Handle("GET "+v1+"/venues/{id}", h.requireAuth(h.getVenue))
	mux.Handle("PUT "+v1+"/venues/{id}", h.requireAuth(h.updateVenue))
	mux.Handle("DELETE "+v1+"/venues/{id}", h.requireAuth(h.deleteVenue))
	mux.Handle("POST "+v1+"/venues/{id}/copy", h.requireAuth(h.copyVenue))
	mux.Handle("PUT "+v1+"/venues/{id}/shared", h.requireAuth(h.shareVenue))
	mux.Handle("POST "+v1+"/events/{id}/seat-map/apply", h.requireAuth(h.applyVenue))
	mux.Handle("POST "+v1+"/events/{id}/sessions", h.requireAuth(h.createSession))
	mux.Handle("PUT "+v1+"/events/{id}/sessions/{session}", h.requireAuth(h.updateSession))
	mux.Handle("DELETE "+v1+"/events/{id}/sessions/{session}", h.requireAuth(h.deleteSession))
	mux.Handle("POST "+v1+"/events/{id}/sessions/{session}/cancel", h.requireAuth(h.cancelSession))
	mux.Handle("POST "+v1+"/events/{id}/duplicate", h.requireAuth(h.duplicateEvent))
	mux.Handle("POST "+v1+"/events/{id}/cancel", h.requireAuth(h.cancelEvent))
	mux.Handle("POST "+v1+"/events/{id}/ticket-types", h.requireAuth(h.createTicketType))
	mux.Handle("PUT "+v1+"/events/{id}/ticket-types/{type}", h.requireAuth(h.updateTicketType))
	mux.Handle("POST "+v1+"/events/{id}/ticket-types/{type}/seats", h.requireAuth(h.generateSeats))
	mux.Handle("PUT "+v1+"/events/{id}/ticket-types/{type}/seat-map", h.requireAuth(h.setSeatMap))
	mux.Handle("GET "+v1+"/events/{id}/promotions", h.requireAuth(h.listPromotions))
	mux.Handle("POST "+v1+"/events/{id}/promotions", h.requireAuth(h.createPromotion))
	mux.Handle("PUT "+v1+"/events/{id}/promotions/{promo}", h.requireAuth(h.updatePromotion))
	mux.Handle("GET "+v1+"/events/{id}/report", h.requireAuth(h.eventReport))
	mux.Handle("GET "+v1+"/events/{id}/orders", h.requireAuth(h.listEventOrders))
	mux.Handle("GET "+v1+"/events/{id}/attendees", h.requireAuth(h.attendees))
	mux.Handle("POST "+v1+"/events/{id}/check-in", h.requireAuth(h.checkIn))
	mux.Handle("POST "+v1+"/events/{id}/check-in/batch", h.requireAuth(h.batchCheckIn))
	mux.Handle("GET "+v1+"/events/{id}/tickets/{code}", h.requireAuth(h.lookupTicket))
	mux.Handle("GET "+v1+"/events/{id}/attendees.csv", h.requireAuth(h.attendeesCSV))
	mux.Handle("POST "+v1+"/events/{id}/invitations", h.requireAuth(h.invite))
	mux.Handle("GET "+v1+"/events/{id}/staff", h.requireAuth(h.listStaff))
	mux.Handle("POST "+v1+"/events/{id}/staff", h.requireAuth(h.addStaff))
	mux.Handle("DELETE "+v1+"/events/{id}/staff/{user}", h.requireAuth(h.removeStaff))
	mux.Handle("GET "+v1+"/my/staff-events", h.requireAuth(h.staffEvents))

	// Admin: moderation.
	mux.Handle("GET "+v1+"/admin/events", h.requireAuth(h.reviewQueue))
	mux.Handle("POST "+v1+"/events/{id}/approve", h.requireAuth(h.eventStep(func(p inbound.Principal, r *http.Request, id int64) (domain.Event, error) {
		return h.events.Approve(r.Context(), p, id)
	})))
	mux.Handle("POST "+v1+"/events/{id}/reject", h.requireAuth(h.rejectEvent))
	mux.Handle("PUT "+v1+"/events/{id}/featured", h.requireAuth(h.setFeatured))

	// Orders: pending -> paid | expired | cancelled; paid -> refunded.
	mux.HandleFunc("POST "+v1+"/orders", h.reserve)
	mux.Handle("GET "+v1+"/orders", h.requireAuth(h.listOrders))
	mux.Handle("GET "+v1+"/orders/{id}", h.requireAuth(h.getOrder))
	mux.Handle("GET "+v1+"/orders/{id}/invoice", h.requireAuth(h.getInvoice))
	mux.Handle("GET "+v1+"/orders/{id}/documents", h.requireAuth(h.listDocuments))
	mux.Handle("GET "+v1+"/orders/{id}/documents/{kind}", h.requireAuth(h.downloadDocument))
	mux.Handle("GET "+v1+"/tickets/{id}/pdf", h.requireAuth(h.ticketPDF))
	mux.Handle("POST "+v1+"/orders/{id}/pay", h.requireAuth(h.payOrder))
	mux.Handle("POST "+v1+"/orders/{id}/cancel", h.requireAuth(h.cancelOrder))

	// Tickets: one per admission, through their life: valid, transferring, used, expired, void.
	mux.Handle("GET "+v1+"/tickets", h.requireAuth(h.listTickets))
	mux.Handle("GET "+v1+"/my/tickets", h.requireAuth(h.listTickets))
	mux.Handle("GET "+v1+"/tickets/{id}", h.requireAuth(h.getTicket))
	mux.Handle("GET "+v1+"/tickets/{id}/history", h.requireAuth(h.ticketHistory))
	mux.Handle("PUT "+v1+"/tickets/{id}/holder", h.requireAuth(h.setHolder))
	mux.Handle("POST "+v1+"/tickets/{id}/reissue", h.requireAuth(h.reissueTicket))
	mux.Handle("POST "+v1+"/tickets/{id}/resale", h.requireAuth(h.listForResale))
	mux.Handle("GET "+v1+"/resale", h.requireAuth(h.myResale))
	mux.Handle("DELETE "+v1+"/resale/{id}", h.requireAuth(h.cancelResale))
	mux.Handle("POST "+v1+"/resale/{id}/buy", h.requireAuth(h.buyResale))
	mux.Handle("POST "+v1+"/tickets/{id}/transfer", h.requireAuth(h.offerTicket))
	mux.Handle("GET "+v1+"/transfers", h.requireAuth(h.listTransfers))
	mux.Handle("POST "+v1+"/transfers/{id}/accept", h.requireAuth(h.acceptTransfer))
	mux.Handle("POST "+v1+"/transfers/{id}/decline", h.requireAuth(h.declineTransfer))
	mux.Handle("POST "+v1+"/transfers/{id}/cancel", h.requireAuth(h.cancelTransfer))
	mux.Handle("POST "+v1+"/tickets/{id}/void", h.requireAuth(h.voidTicket))
	mux.Handle("POST "+v1+"/tickets/{id}/revert-check-in", h.requireAuth(h.revertCheckIn))

	// Waiting lists: "tell me when tickets are back".
	mux.Handle("GET "+v1+"/waitlist", h.requireAuth(h.myWaitlist))
	mux.Handle("PUT "+v1+"/events/{id}/ticket-types/{type}/waitlist", h.requireAuth(h.joinWaitlist))
	mux.Handle("DELETE "+v1+"/events/{id}/ticket-types/{type}/waitlist", h.requireAuth(h.leaveWaitlist))

	// Favourite events.
	mux.Handle("GET "+v1+"/wishlist", h.requireAuth(h.listWishlist))
	mux.Handle("PUT "+v1+"/wishlist/{id}", h.requireAuth(h.addWishlist))
	mux.Handle("DELETE "+v1+"/wishlist/{id}", h.requireAuth(h.removeWishlist))

	// Notifications: the caller's inbox.
	mux.Handle("GET "+v1+"/notifications", h.requireAuth(h.listNotifications))
	mux.Handle("GET "+v1+"/notifications/unread-count", h.requireAuth(h.unreadCount))
	mux.Handle("POST "+v1+"/notifications/read-all", h.requireAuth(h.markAllRead))
	mux.Handle("POST "+v1+"/notifications/{id}/read", h.requireAuth(h.markRead))

	if internalKey != "" {
		const in = "/internal/v1"
		key := func(f http.HandlerFunc) http.Handler { return h.internalAuth(internalKey, f) }
		mux.Handle("GET "+in+"/ticket-codes/{code}", key(h.internalTicketByCode))
		mux.Handle("GET "+in+"/tickets/{id}", key(h.getTicket))
		mux.Handle("GET "+in+"/tickets/{id}/history", key(h.ticketHistory))
		mux.Handle("POST "+in+"/tickets/{id}/void", key(h.voidTicket))
		mux.Handle("POST "+in+"/tickets/{id}/revert-check-in", key(h.revertCheckIn))
		mux.Handle("GET "+in+"/users/{user}/tickets", key(h.internalUserTickets))
		mux.Handle("GET "+in+"/orders/{id}", key(h.getOrder))
		mux.Handle("GET "+in+"/orders/{id}/invoice", key(h.getInvoice))
		mux.Handle("GET "+in+"/orders/{id}/documents", key(h.listDocuments))
		mux.Handle("GET "+in+"/orders/{id}/documents/{kind}", key(h.downloadDocument))
		mux.Handle("POST "+in+"/orders/{id}/cancel", key(h.cancelOrder))
		mux.Handle("GET "+in+"/events/{id}", key(h.getEvent))
		mux.Handle("GET "+in+"/events/{id}/report", key(h.eventReport))
		mux.Handle("POST "+in+"/events/{id}/check-in", key(h.checkIn))
		mux.Handle("POST "+in+"/events/{id}/check-in/batch", key(h.batchCheckIn))
		mux.Handle("POST "+in+"/events/{id}/invitations", key(h.invite))
	}

	return limitInFlight(maxInFlight, logging(log, mux))
}
