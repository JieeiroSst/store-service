package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
)

type viewerUsecase struct {
	counters port.ViewerCounter
	cfg      *config.Config
}

func NewViewerUsecase(counters port.ViewerCounter, cfg *config.Config) port.ViewerUsecase {
	return &viewerUsecase{counters: counters, cfg: cfg}
}

func (u *viewerUsecase) Heartbeat(ctx context.Context, roomID, sessionID string) error {
	if err := u.counters.Heartbeat(ctx, roomID, sessionID, u.cfg.Viewer.HeartbeatWindowDuration()); err != nil {
		return fmt.Errorf("viewer heartbeat: %w", err)
	}
	return nil
}

func (u *viewerUsecase) GetViewerCount(ctx context.Context, roomID string) (int64, error) {
	return u.counters.Get(ctx, roomID, u.cfg.Viewer.HeartbeatWindowDuration())
}
