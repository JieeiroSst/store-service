package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type Notifier struct {
	repo outbound.NotificationRepository
	log  *slog.Logger
}

func NewNotifier(repo outbound.NotificationRepository, log *slog.Logger) *Notifier {
	if repo == nil {
		return nil
	}
	return &Notifier{repo: repo, log: defaultLog(log)}
}

func (n *Notifier) Send(ctx context.Context, userID int64, kind, title, body string, reservationID, hotelID int64) {
	if n == nil || userID == 0 {
		return
	}
	if err := n.repo.Add(ctx, domain.Notification{UserID: userID, Kind: kind, Title: title, Body: body, ReservationID: reservationID, HotelID: hotelID}); err != nil {
		n.log.Warn("notification not stored", "user", userID, "kind", kind, "err", err)
	}
}

const pushLease = 30 * time.Second

type NotificationService struct {
	repo outbound.NotificationRepository
	push outbound.PushGateway
	log  *slog.Logger
}

var _ inbound.NotificationUseCase = (*NotificationService)(nil)

func NewNotificationService(repo outbound.NotificationRepository, push outbound.PushGateway, log *slog.Logger) *NotificationService {
	return &NotificationService{repo: repo, push: push, log: defaultLog(log)}
}

func (s *NotificationService) List(ctx context.Context, a inbound.Principal, unreadOnly bool, afterID int64, limit int) (domain.Page[domain.Notification], error) {
	limit, _ = page(limit, 0)
	rows, err := s.repo.List(ctx, a.UserID, unreadOnly, afterID, limit+1)
	if err != nil {
		return domain.Page[domain.Notification]{}, err
	}
	return domain.PageOf(rows, limit, func(n domain.Notification) int64 { return n.ID }), nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, a inbound.Principal) (int, error) {
	return s.repo.UnreadCount(ctx, a.UserID)
}

func (s *NotificationService) MarkRead(ctx context.Context, a inbound.Principal, id int64) error {
	return s.repo.MarkRead(ctx, a.UserID, id)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, a inbound.Principal) (int, error) {
	return s.repo.MarkAllRead(ctx, a.UserID)
}

func (s *NotificationService) PushPending(ctx context.Context) (int, error) {
	if s.push == nil {
		return 0, nil
	}
	pending, err := s.repo.ClaimUnpushed(ctx, 100, pushLease)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, n := range pending {
		if err := s.push.Push(ctx, n); err != nil {
			s.log.Warn("push to notification-service failed", "notification", n.ID, "err", err)
			continue
		}
		if err := s.repo.MarkPushed(ctx, n.ID); err != nil {
			s.log.Error("push recorded as failed though it was sent", "notification", n.ID, "err", err)
			continue
		}
		sent++
	}
	return sent, nil
}
