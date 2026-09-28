package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"github.com/google/uuid"
)

type channelService struct {
	channels port.ChannelRepository
	jobs     port.JobRepository
}

func NewChannelService(channels port.ChannelRepository, jobs port.JobRepository) port.ChannelUsecase {
	return &channelService{channels: channels, jobs: jobs}
}

func validateChannel(c *model.Channel) error {
	if c.Name == "" {
		return invalid("channel name is required")
	}
	if !c.Type.Valid() {
		return invalid("channel type must be email|sms|firebase")
	}
	return nil
}

func (s *channelService) Create(ctx context.Context, c *model.Channel) (*model.Channel, error) {
	if err := validateChannel(c); err != nil {
		return nil, err
	}
	now := time.Now()
	c.ID = uuid.NewString()
	c.CreatedAt, c.UpdatedAt = now, now
	if err := s.channels.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *channelService) Get(ctx context.Context, id string) (*model.Channel, error) {
	return s.channels.Get(ctx, id)
}

func (s *channelService) List(ctx context.Context, f port.ChannelFilter) ([]*model.Channel, error) {
	return s.channels.List(ctx, f)
}

func (s *channelService) Update(ctx context.Context, id string, apply func(*model.Channel) error) (*model.Channel, error) {
	c, err := s.channels.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	createdAt := c.CreatedAt
	if err := apply(c); err != nil {
		return nil, invalid("%v", err)
	}
	c.ID, c.CreatedAt, c.UpdatedAt = id, createdAt, time.Now()
	if err := validateChannel(c); err != nil {
		return nil, err
	}
	if err := s.channels.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *channelService) Delete(ctx context.Context, id string) error {
	n, err := s.jobs.CountByChannel(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("%w: channel is used by %d job(s)", port.ErrConflict, n)
	}
	return s.channels.Delete(ctx, id)
}
