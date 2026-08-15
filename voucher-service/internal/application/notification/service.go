package notification

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Service struct {
	repo     NotificationRepository
	notifier NotifierRegistry
}

func NewService(repo NotificationRepository, notifier NotifierRegistry) NotificationService {
	return &Service{repo: repo, notifier: notifier}
}

func (s *Service) Send(ctx context.Context, in SendInput) error {
	id := shared.NewVoucherID().String()
	if err := s.repo.Create(ctx, id, in.RecipientType, in.RecipientID, in.Channel, in.TemplateCode, in.Payload); err != nil {
		return err
	}

	n, err := s.notifier.Resolve(in.Channel)
	if err != nil {
		_ = s.repo.MarkFailed(ctx, id, err.Error())
		return err
	}

	if err := n.Send(ctx, in.RecipientID, in.TemplateCode, in.Payload); err != nil {
		_ = s.repo.MarkFailed(ctx, id, err.Error())
		return err
	}

	return s.repo.MarkSent(ctx, id)
}
