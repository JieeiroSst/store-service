package store

import (
	"context"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/model"
)

type JobStore interface {
	CreateHistory(ctx context.Context, h *model.NotifyHistory) error
	UpdateHistoryStatus(ctx context.Context, id string, status model.NotifyStatus, errMsg string) error
	UpdateJobRun(ctx context.Context, id string, next *time.Time, failed bool) error
}
