package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Expiry struct {
	Month int
	Year  int
}

func NewExpiry(issuedAt time.Time, validityYears int) Expiry {
	t := issuedAt.AddDate(validityYears, 0, 0)
	return Expiry{Month: int(t.Month()), Year: t.Year()}
}

func ParseExpiry(s string) (Expiry, error) {
	mm, yy, ok := strings.Cut(strings.TrimSpace(s), "/")
	if !ok || len(mm) != 2 || len(yy) != 2 || !isDigits(mm) || !isDigits(yy) {
		return Expiry{}, Invalid("expiry must be MM/YY, got %q", s)
	}
	m, _ := strconv.Atoi(mm)
	y, _ := strconv.Atoi(yy)
	if m < 1 || m > 12 {
		return Expiry{}, Invalid("expiry month must be 01-12, got %q", mm)
	}
	return Expiry{Month: m, Year: 2000 + y}, nil
}

func (e Expiry) String() string {
	return fmt.Sprintf("%02d/%02d", e.Month, e.Year%100)
}

func (e Expiry) YYMM() string {
	return fmt.Sprintf("%02d%02d", e.Year%100, e.Month)
}

func (e Expiry) Expired(now time.Time) bool {
	firstDayAfter := time.Date(e.Year, time.Month(e.Month)+1, 1, 0, 0, 0, 0, time.UTC)
	return !now.UTC().Before(firstDayAfter)
}
