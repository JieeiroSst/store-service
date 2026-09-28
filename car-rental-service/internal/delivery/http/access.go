package http

import (
	"context"
	"strconv"

	"github.com/JIeeiroSst/car-rental-service/internal/auth"
	"github.com/JIeeiroSst/car-rental-service/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) actingUser(ctx context.Context, requested string) (string, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return "", status.Error(codes.Unauthenticated, "authentication required")
	}
	if requested == "" || requested == claims.Subject {
		return claims.Subject, nil
	}
	if !h.auth.IsStaff(claims) {
		return "", status.Error(codes.PermissionDenied, "cannot act on another user")
	}
	return requested, nil
}

func (h *Handler) canAccess(ctx context.Context, ownerID int64) bool {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return false
	}
	return h.auth.IsStaff(claims) || claims.Subject == strconv.FormatInt(ownerID, 10)
}

func (h *Handler) authorizeReservation(ctx context.Context, reservationID string) error {
	r, err := h.usecase.GetReservation(ctx, reservationID)
	if err != nil {
		return grpcError(err)
	}
	if !h.canAccess(ctx, r.UserID) {
		return grpcError(model.ErrNotFound)
	}
	return nil
}

func staffID(ctx context.Context) string {
	if c := auth.FromContext(ctx); c != nil {
		return c.Subject
	}
	return ""
}
