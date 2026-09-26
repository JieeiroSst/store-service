package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

var errUnauthorized = domain.ErrUnauthorized

type Pinger func(ctx context.Context) error

func NewRouter(h *Handler, ping Pinger, maxInFlight int, log *slog.Logger) http.Handler {
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

	// Public: browse hotels, their rooms, availability and prices.
	mux.HandleFunc("GET "+v1+"/hotels", h.searchHotels)
	mux.Handle("GET "+v1+"/hotels/{id}", h.optionalAuth(http.HandlerFunc(h.getHotel)))
	mux.HandleFunc("GET "+v1+"/hotels/{id}/room-types", h.listRoomTypes)
	mux.HandleFunc("GET "+v1+"/hotels/{id}/room-types/{type}/inventory", h.getInventory)
	mux.HandleFunc("GET "+v1+"/hotels/{id}/quotes", h.quotes)
	mux.HandleFunc("GET "+v1+"/hotels/{id}/services", h.listServices)
	mux.HandleFunc("GET "+v1+"/hotels/{id}/reviews", h.listReviews)

	// Registration: any signed-in user registers a 5-star hotel and manages it; an admin verifies it
	// before it goes public.
	mux.Handle("POST "+v1+"/hotels", h.requireAuth(h.registerHotel))
	mux.Handle("GET "+v1+"/my/hotels", h.requireAuth(h.myHotels))
	mux.Handle("GET "+v1+"/admin/hotels", h.requireAuth(h.reviewQueue))
	mux.Handle("POST "+v1+"/hotels/{id}/approve", h.requireAuth(h.approveHotel))
	mux.Handle("POST "+v1+"/hotels/{id}/reject", h.requireAuth(h.rejectHotel))

	// Hotel management (authorization is enforced in the use case: the hotel's manager or an admin).
	mux.Handle("PUT "+v1+"/hotels/{id}", h.requireAuth(h.updateHotel))
	mux.Handle("DELETE "+v1+"/hotels/{id}", h.requireAuth(h.deactivateHotel))
	mux.Handle("POST "+v1+"/hotels/{id}/room-types", h.requireAuth(h.createRoomType))
	mux.Handle("PUT "+v1+"/hotels/{id}/room-types/{type}", h.requireAuth(h.updateRoomType))
	mux.Handle("PUT "+v1+"/hotels/{id}/room-types/{type}/inventory", h.requireAuth(h.setInventory))
	mux.Handle("GET "+v1+"/hotels/{id}/rooms", h.requireAuth(h.listRooms))
	mux.Handle("POST "+v1+"/hotels/{id}/rooms", h.requireAuth(h.createRoom))
	mux.Handle("PUT "+v1+"/hotels/{id}/rooms/{room}", h.requireAuth(h.updateRoom))
	mux.Handle("GET "+v1+"/hotels/{id}/promotions", h.requireAuth(h.listPromotions))
	mux.Handle("POST "+v1+"/hotels/{id}/promotions", h.requireAuth(h.createPromotion))
	mux.Handle("PUT "+v1+"/hotels/{id}/promotions/{promo}", h.requireAuth(h.updatePromotion))
	mux.Handle("GET "+v1+"/hotels/{id}/report", h.requireAuth(h.hotelReport))
	mux.Handle("POST "+v1+"/hotels/{id}/services", h.requireAuth(h.createService))
	mux.Handle("PUT "+v1+"/hotels/{id}/services/{service}", h.requireAuth(h.updateService))

	// Reservations: pending -> paid | canceled | rejected; paid -> refunded.
	// Not wrapped in requireAuth: reserve turns away a sold-out request before it spends an identity lookup on it.
	mux.HandleFunc("POST "+v1+"/reservations", h.reserve)
	mux.Handle("GET "+v1+"/reservations", h.requireAuth(h.listReservations))
	mux.Handle("GET "+v1+"/reservations/{id}", h.requireAuth(h.getReservation))
	mux.Handle("POST "+v1+"/reservations/{id}/pay", h.requireAuth(h.payReservation))
	mux.Handle("GET "+v1+"/reservations/{id}/cancellation-quote", h.requireAuth(h.cancellationQuote))
	mux.Handle("POST "+v1+"/reservations/{id}/cancel", h.requireAuth(h.cancelReservation))
	mux.Handle("POST "+v1+"/reservations/{id}/reject", h.requireAuth(h.rejectReservation))
	mux.Handle("POST "+v1+"/reservations/{id}/check-in", h.requireAuth(h.stayStep(h.reservations.CheckIn)))
	mux.Handle("POST "+v1+"/reservations/{id}/check-out", h.requireAuth(h.stayStep(h.reservations.CheckOut)))
	mux.Handle("POST "+v1+"/reservations/{id}/no-show", h.requireAuth(h.stayStep(h.reservations.NoShow)))
	mux.Handle("GET "+v1+"/reservations/{id}/history", h.requireAuth(h.reservationHistory))
	mux.Handle("POST "+v1+"/hotels/{id}/reviews", h.requireAuth(h.createReview))

	// The waiting list for sold-out rooms: the oldest guest whose stay fits is offered rooms as they come back.
	mux.Handle("POST "+v1+"/waitlist", h.requireAuth(h.joinWaitlist))
	mux.Handle("GET "+v1+"/waitlist", h.requireAuth(h.listWaitlist))
	mux.Handle("DELETE "+v1+"/waitlist/{id}", h.requireAuth(h.leaveWaitlist))

	// Favourite hotels.
	mux.Handle("GET "+v1+"/wishlist", h.requireAuth(h.listFavourites))
	mux.Handle("PUT "+v1+"/wishlist/{id}", h.requireAuth(h.addFavourite))
	mux.Handle("DELETE "+v1+"/wishlist/{id}", h.requireAuth(h.removeFavourite))

	// Notifications: the caller's inbox.
	mux.Handle("GET "+v1+"/notifications", h.requireAuth(h.listNotifications))
	mux.Handle("GET "+v1+"/notifications/unread-count", h.requireAuth(h.unreadCount))
	mux.Handle("POST "+v1+"/notifications/read-all", h.requireAuth(h.markAllRead))
	mux.Handle("POST "+v1+"/notifications/{id}/read", h.requireAuth(h.markRead))

	// Loyalty points and the caller's own wallet in payment-wallet-service.
	mux.Handle("GET "+v1+"/loyalty", h.requireAuth(h.loyaltyStatus))
	mux.Handle("GET "+v1+"/wallet", h.requireAuth(h.getWallet))
	mux.Handle("POST "+v1+"/wallet", h.requireAuth(h.createWallet))
	mux.Handle("GET "+v1+"/wallet/transactions", h.requireAuth(h.walletTransactions))

	return limitInFlight(maxInFlight, logging(log, mux))
}
