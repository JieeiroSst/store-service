package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type fakeHomestays struct {
	outbound.HomestayRepository
	h domain.Homestay
}

func (f fakeHomestays) Get(context.Context, int64) (domain.Homestay, error) { return f.h, nil }

type fakeBookings struct {
	outbound.BookingRepository
	reserved *outbound.ReserveParams
}

func (f *fakeBookings) Reserve(_ context.Context, p outbound.ReserveParams) (domain.Booking, error) {
	f.reserved = &p
	return domain.Booking{ID: 1, UserID: p.UserID}, nil
}

func newBookingSvc() (*BookingService, *fakeBookings) {
	b := &fakeBookings{}
	s := NewBookingService(b, fakeHomestays{h: domain.Homestay{ID: 1, Status: domain.HomestayActive, Guests: 4, HostID: 50}}, NewPaymentRail(nil, nil, nil, nil), BookingOptions{})
	s.now = func() time.Time { return time.Date(2026, 1, 10, 15, 0, 0, 0, time.UTC) }
	return s, b
}

func date(d int) time.Time { return time.Date(2026, 1, d, 0, 0, 0, 0, time.UTC) }

func TestBook(t *testing.T) {
	guest := inbound.Principal{UserID: 7}
	owner := inbound.Principal{UserID: 50}
	admin := inbound.Principal{UserID: 1, Admin: true}
	ok := inbound.BookCommand{HomestayID: 1, CheckIn: date(12), CheckOut: date(14), Guests: 2}

	tests := []struct {
		name  string
		actor inbound.Principal
		mod   func(*inbound.BookCommand)
		want  error
	}{
		{"ok", guest, nil, nil},
		{"owner cannot rent their own homestay", owner, nil, domain.ErrForbidden},
		{"admins may rent like anyone", admin, nil, nil},
		{"past check-in", guest, func(c *inbound.BookCommand) { c.CheckIn, c.CheckOut = date(1), date(3) }, domain.ErrInvalid},
		{"checkout before checkin", guest, func(c *inbound.BookCommand) { c.CheckOut = date(12) }, domain.ErrInvalid},
		{"too many guests", guest, func(c *inbound.BookCommand) { c.Guests = 5 }, domain.ErrInvalid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newBookingSvc()
			cmd := ok
			if tc.mod != nil {
				tc.mod(&cmd)
			}
			_, err := svc.Book(context.Background(), tc.actor, cmd)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if tc.want == nil {
				if repo.reserved == nil || repo.reserved.Currency != "VND" || repo.reserved.RequestID == "" {
					t.Fatalf("defaults not applied: %+v", repo.reserved)
				}
			} else if repo.reserved != nil {
				t.Fatal("repository called despite failure")
			}
		})
	}
}
