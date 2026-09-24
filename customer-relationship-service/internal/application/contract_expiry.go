package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

type contractExpiry struct {
	expirer  port.ContractExpirer
	notifier port.Notifier
	now      func() time.Time
}

func NewContractExpiry(e port.ContractExpirer, n port.Notifier) port.ContractExpiryUsecase {
	return &contractExpiry{expirer: e, notifier: n, now: time.Now}
}

func (s *contractExpiry) ExpireOverdue(ctx context.Context) (int64, error) {
	n, err := s.expirer.ExpireDue(ctx, s.now())
	if err != nil {
		return 0, err
	}
	if n > 0 {
		notify(ctx, s.notifier, "Contracts expired", "%d contract(s) passed their end date and were closed", n)
	}
	return n, nil
}
