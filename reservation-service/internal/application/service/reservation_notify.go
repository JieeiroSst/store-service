package service

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

const dateOnly = "2 Jan 2006"

func stayText(x domain.Reservation) string {
	return fmt.Sprintf("%s to %s (%d night(s), %d room(s))", x.Start.Format(dateOnly), x.End.Format(dateOnly), x.Nights(), x.Rooms)
}

func (s *ReservationService) notifyCreated(ctx context.Context, x domain.Reservation, h domain.Hotel) {
	s.opts.Notifier.Send(ctx, h.OwnerID, "reservation.created",
		fmt.Sprintf("New reservation #%d at %s", x.ID, h.Name),
		fmt.Sprintf("A guest is holding rooms for %s, total %s %s. It is pending until they pay.", stayText(x), x.TotalAmount, x.Currency),
		x.ID, h.ID)
}

func (s *ReservationService) notifyPaid(ctx context.Context, x domain.Reservation) {
	h, err := s.hotels.Get(ctx, x.HotelID)
	if err != nil {
		return
	}
	s.opts.Notifier.Send(ctx, x.GuestID, "reservation.confirmed",
		fmt.Sprintf("Reservation #%d confirmed at %s", x.ID, h.Name),
		fmt.Sprintf("Your stay %s is confirmed. Check-in from %s.", stayText(x), h.CheckInTime), x.ID, h.ID)
	s.opts.Notifier.Send(ctx, h.OwnerID, "reservation.paid",
		fmt.Sprintf("Reservation #%d paid", x.ID),
		fmt.Sprintf("%s, %s %s received.", stayText(x), x.TotalAmount, x.Currency), x.ID, h.ID)
}

func (s *ReservationService) notifyCancelled(ctx context.Context, x domain.Reservation, byHotel bool) {
	h, err := s.hotels.Get(ctx, x.HotelID)
	if err != nil {
		return
	}
	detail := ""
	if x.Status == domain.ReservationRefunded {
		detail = fmt.Sprintf(" %s %s was paid back", x.RefundAmount, x.Currency)
		if x.FeeAmount != "" && x.FeeAmount != "0.0000" {
			detail += fmt.Sprintf(", %s %s kept as the cancellation fee", x.FeeAmount, x.Currency)
		}
		detail += "."
	}
	if byHotel {
		s.opts.Notifier.Send(ctx, x.GuestID, "reservation.cancelled_by_hotel",
			fmt.Sprintf("Reservation #%d was cancelled by %s", x.ID, h.Name),
			fmt.Sprintf("Your stay %s was cancelled by the hotel.%s", stayText(x), detail), x.ID, h.ID)
		return
	}
	s.opts.Notifier.Send(ctx, h.OwnerID, "reservation.cancelled",
		fmt.Sprintf("Reservation #%d cancelled by the guest", x.ID),
		fmt.Sprintf("%s. The rooms are available again.%s", stayText(x), detail), x.ID, h.ID)
	if x.Status == domain.ReservationRefunded {
		s.opts.Notifier.Send(ctx, x.GuestID, "reservation.refunded",
			fmt.Sprintf("Reservation #%d cancelled", x.ID), fmt.Sprintf("Your cancellation is done.%s", detail), x.ID, h.ID)
	}
}

func (s *ReservationService) notifyRejected(ctx context.Context, x domain.Reservation, reason string) {
	h, err := s.hotels.Get(ctx, x.HotelID)
	if err != nil {
		return
	}
	s.opts.Notifier.Send(ctx, x.GuestID, "reservation.rejected",
		fmt.Sprintf("Reservation #%d was declined by %s", x.ID, h.Name),
		fmt.Sprintf("Your request for %s was declined: %s", stayText(x), reason), x.ID, h.ID)
}

func (s *ReservationService) notifyExpired(ctx context.Context, x domain.Reservation) {
	s.opts.Notifier.Send(ctx, x.GuestID, "reservation.expired",
		fmt.Sprintf("Reservation #%d expired", x.ID),
		fmt.Sprintf("The rooms held for %s were released because payment did not arrive in time.", stayText(x)), x.ID, x.HotelID)
}

func (s *ReservationService) SendReminders(ctx context.Context) (int, error) {
	if s.opts.Notifier == nil {
		return 0, nil
	}
	today := s.today()
	due, err := s.reservations.DueForReminder(ctx, today, today.Add(48*time.Hour), 200)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, x := range due {
		if ok, err := s.reservations.MarkReminded(ctx, x.ID); err != nil || !ok {
			continue
		}
		h, err := s.hotels.Get(ctx, x.HotelID)
		if err != nil {
			continue
		}
		when := "tomorrow"
		if x.Start.Equal(today) {
			when = "today"
		}
		s.opts.Notifier.Send(ctx, x.GuestID, "reservation.reminder",
			fmt.Sprintf("Your stay at %s starts %s", h.Name, when),
			fmt.Sprintf("%s. Check-in from %s at %s, %s. Phone %s.", stayText(x), h.CheckInTime, h.Name, h.Address, h.PhoneNumber), x.ID, h.ID)
		sent++
	}
	return sent, nil
}
