package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

const (
	maxPageSize     = 100
	defaultPageSize = 20
	maxRangeDays    = 366
	expireBatch     = 100
)

func page(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func requireAdmin(p inbound.Principal) error {
	if !p.IsAdmin() {
		return domain.ErrForbidden
	}
	return nil
}

func canManage(a inbound.Principal, h domain.Hotel) bool {
	return a.IsAdmin() || (a.UserID != 0 && h.OwnerID == a.UserID)
}

func validRange(from, to time.Time) error {
	if !to.After(from) {
		return fmt.Errorf("%w: end date must be after start date", domain.ErrInvalid)
	}
	if to.Sub(from) > maxRangeDays*24*time.Hour {
		return fmt.Errorf("%w: range cannot exceed %d days", domain.ErrInvalid, maxRangeDays)
	}
	return nil
}

func reservationAccess(ctx context.Context, hotels outbound.HotelRepository, actor inbound.Principal, guestID, hotelID int64) (allowed, privileged bool) {
	if actor.IsAdmin() {
		return true, true
	}
	if actor.UserID == guestID {
		return true, false
	}
	h, err := hotels.Get(ctx, hotelID)
	if err != nil || h.OwnerID == 0 || h.OwnerID != actor.UserID {
		return false, false
	}
	return true, true
}

func defaultLog(l *slog.Logger) *slog.Logger {
	if l == nil {
		return slog.Default()
	}
	return l
}

func trim(s string) string { return strings.TrimSpace(s) }
