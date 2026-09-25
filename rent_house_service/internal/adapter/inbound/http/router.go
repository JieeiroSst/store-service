package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

var errUnauthorized = domain.ErrUnauthorized

type Pinger func(ctx context.Context) error

func NewRouter(h *Handler, ping Pinger, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := ping(ctx); err != nil {
			log.Error("readiness", "err", errors.Unwrap(err))
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	const v1 = "/api/v1"
	// Accounts, login and sessions live in user-service; callers send its access token as a Bearer token.

	// Public browsing.
	mux.Handle("GET "+v1+"/homestays/{id}", h.optionalAuth(http.HandlerFunc(h.getHomestay)))
	mux.HandleFunc("GET "+v1+"/homestays", h.listHomestays)
	mux.HandleFunc("GET "+v1+"/homestays/{id}/availability", h.availability)
	mux.HandleFunc("GET "+v1+"/amenities", h.listAmenities)

	// Registration: any signed-in user registers a homestay, hotel or rental house and owns it;
	// an admin approves it before it goes public. Owners manage only what they own.
	mux.Handle("GET "+v1+"/my/homestays", h.requireAuth(h.myHomestays))
	mux.Handle("GET "+v1+"/admin/homestays", h.requireAuth(h.reviewQueue))
	mux.Handle("POST "+v1+"/homestays/{id}/approve", h.requireAuth(h.approveHomestay))
	mux.Handle("POST "+v1+"/homestays/{id}/reject", h.requireAuth(h.rejectHomestay))

	// Owner management (authorization is enforced in the use case).
	mux.Handle("POST "+v1+"/homestays", h.requireAuth(h.createHomestay))
	mux.Handle("PUT "+v1+"/homestays/{id}", h.requireAuth(h.updateHomestay))
	mux.Handle("DELETE "+v1+"/homestays/{id}", h.requireAuth(h.deactivateHomestay))
	mux.Handle("PUT "+v1+"/homestays/{id}/availability", h.requireAuth(h.setAvailability))
	mux.Handle("POST "+v1+"/amenities", h.requireAuth(h.createAmenity))

	// Renting.
	mux.Handle("POST "+v1+"/bookings", h.requireAuth(h.book))
	mux.Handle("GET "+v1+"/bookings", h.requireAuth(h.listBookings))
	mux.Handle("GET "+v1+"/bookings/{id}", h.requireAuth(h.getBooking))
	mux.Handle("POST "+v1+"/bookings/{id}/cancel", h.requireAuth(h.cancelBooking))
	mux.Handle("POST "+v1+"/bookings/{id}/pay", h.requireAuth(h.payBooking))

	// Rental models: the manager prices each one; week/month/year are leases billed per period.
	mux.HandleFunc("GET "+v1+"/homestays/{id}/rates", h.listRates)
	mux.Handle("PUT "+v1+"/homestays/{id}/rates/{model}", h.requireAuth(h.setRate))
	mux.Handle("DELETE "+v1+"/homestays/{id}/rates/{model}", h.requireAuth(h.deleteRate))
	mux.Handle("POST "+v1+"/leases", h.requireAuth(h.startLease))
	mux.Handle("GET "+v1+"/leases", h.requireAuth(h.listLeases))
	mux.Handle("GET "+v1+"/leases/{id}", h.requireAuth(h.getLease))
	mux.Handle("POST "+v1+"/leases/{id}/cancel", h.requireAuth(h.cancelLease))
	mux.Handle("POST "+v1+"/leases/{id}/terminate", h.requireAuth(h.terminateLease))
	mux.Handle("POST "+v1+"/leases/{id}/invoices/{period}/pay", h.requireAuth(h.payInvoice))
	mux.Handle("GET "+v1+"/invoices", h.requireAuth(h.listInvoices))

	// Reviews, wishlist and loyalty.
	mux.HandleFunc("GET "+v1+"/homestays/{id}/reviews", h.listReviews)
	mux.Handle("POST "+v1+"/homestays/{id}/reviews", h.requireAuth(h.createReview))
	mux.Handle("GET "+v1+"/wishlist", h.requireAuth(h.listWishlist))
	mux.Handle("PUT "+v1+"/wishlist/{id}", h.requireAuth(h.addWishlist))
	mux.Handle("DELETE "+v1+"/wishlist/{id}", h.requireAuth(h.removeWishlist))
	mux.Handle("GET "+v1+"/loyalty", h.requireAuth(h.loyaltyStatus))

	// The caller's own wallet in payment-wallet-service.
	mux.Handle("GET "+v1+"/wallet", h.requireAuth(h.getWallet))
	mux.Handle("POST "+v1+"/wallet", h.requireAuth(h.createWallet))
	mux.Handle("GET "+v1+"/wallet/transactions", h.requireAuth(h.walletTransactions))

	return logging(log, mux)
}
