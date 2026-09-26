package service

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/mail"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

const (
	maxPageSize     = 100
	defaultPageSize = 20
	expireBatch     = 200
)

func page(limit int) int {
	if limit <= 0 {
		return defaultPageSize
	}
	return min(limit, maxPageSize)
}

func formatMoney(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		return "-" + s
	}
	return s
}

func trim(s string) string { return strings.TrimSpace(s) }

func normalizeCode(c string) string { return strings.ToUpper(strings.TrimSpace(c)) }

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

var codeEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

func newTicketCode() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return codeEncoding.EncodeToString(b), nil
}

func requireAdmin(p inbound.Principal) error {
	if !p.IsAdmin() {
		return domain.ErrForbidden
	}
	return nil
}

func canManage(a inbound.Principal, e domain.Event) bool {
	return a.IsAdmin() || (a.UserID != 0 && e.OrganizerID == a.UserID)
}

func notYours(e domain.Event) error {
	if e.Status == domain.EventPublished {
		return domain.ErrForbidden
	}
	return domain.ErrNotFound
}

func invalid(format string, a ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{domain.ErrInvalid}, a...)...)
}

func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s
}

func defaultLog(l *slog.Logger) *slog.Logger {
	if l == nil {
		return slog.Default()
	}
	return l
}

func showtime(e domain.Event, id int64) domain.Session {
	if s, ok := sessionOf(e, id); ok {
		return s
	}
	return domain.Session{EventID: e.ID, StartsAt: e.StartsAt, EndsAt: e.EndsAt, Status: domain.SessionScheduled}
}

func titleOf(e domain.Event, s domain.Session) string {
	if s.Label != "" {
		return e.Title + " - " + s.Label
	}
	return e.Title
}
