package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/metrics"
)

type nodeSchedulerUsecase struct {
	assigner *nodeAssigner
	runner   port.TranscodeRunner
	cfg      *config.Config
}

func NewNodeSchedulerUsecase(nodes port.NodeRegistry, runner port.TranscodeRunner, cfg *config.Config) port.NodeSchedulerUsecase {
	return &nodeSchedulerUsecase{assigner: newNodeAssigner(nodes), runner: runner, cfg: cfg}
}

func (s *nodeSchedulerUsecase) Heartbeat(ctx context.Context) error {
	active := len(s.runner.ActiveStreamKeys())
	max := s.cfg.Node.MaxStreams
	capacity := max - active
	if capacity < 0 {
		capacity = 0
	}

	node := model.TranscodeNode{
		ID:            s.cfg.Node.ID,
		Addr:          s.cfg.Node.RTMPAddr,
		HTTPAddr:      s.cfg.Node.HTTPAddr,
		ActiveStreams: active,
		MaxStreams:    max,
		Capacity:      capacity,
		LoadPerCore:   float64(active) / float64(max),
	}

	metrics.ActiveStreams.Set(float64(active))
	metrics.NodeCapacity.Set(float64(capacity))

	return s.assigner.nodes.Heartbeat(ctx, node, s.cfg.Transcode.HeartbeatTTLDuration())
}

func (s *nodeSchedulerUsecase) AssignNode(ctx context.Context, streamKey string) (*model.TranscodeNode, error) {
	return s.assigner.assign(ctx, streamKey)
}

func (s *nodeSchedulerUsecase) ReleaseNode(ctx context.Context, streamKey string) error {
	return s.assigner.release(ctx, streamKey)
}

func (s *nodeSchedulerUsecase) CheckStale(ctx context.Context) error {
	stale := s.runner.StaleSince(s.cfg.Transcode.WatchdogStaleAfterDuration())
	for _, streamKey := range stale {
		if err := s.runner.Restart(ctx, streamKey); err != nil {
			return fmt.Errorf("restart %s: %w", streamKey, err)
		}
	}
	return nil
}
