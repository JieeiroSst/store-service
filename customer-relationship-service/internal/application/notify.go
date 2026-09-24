package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/sirupsen/logrus"
)

func syncLifecycle(ctx context.Context, o port.ContractOrchestrator, id uint) {
	if err := o.Sync(ctx, id); err != nil {
		logrus.WithError(err).WithField("contract", id).Warn("contract lifecycle not synced")
	}
}

func notify(ctx context.Context, n port.Notifier, title, format string, args ...any) {
	msg := model.Notification{Title: title, Message: fmt.Sprintf(format, args...)}
	if err := n.Notify(ctx, msg); err != nil {
		logrus.WithError(err).Warn("notification not delivered")
	}
}
