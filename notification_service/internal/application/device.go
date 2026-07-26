package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

type userDeviceService struct {
	repo port.UserDeviceRepository
}

func NewUserDeviceService(repo port.UserDeviceRepository) port.UserDeviceUsecase {
	return &userDeviceService{repo: repo}
}

func (s *userDeviceService) RegisterDevice(ctx context.Context, device *model.UserDevice) (*model.UserDevice, error) {
	device.IsActive = true
	device.LastUsedAt = time.Now()
	device.CreatedAt = time.Now()

	if err := s.repo.Create(ctx, device); err != nil {
		return nil, err
	}
	return device, nil
}

func (s *userDeviceService) GetDevice(ctx context.Context, id uint) (*model.UserDevice, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userDeviceService) ListDevices(ctx context.Context) ([]model.UserDevice, error) {
	return s.repo.List(ctx)
}

func (s *userDeviceService) UpdateDevice(ctx context.Context, device *model.UserDevice) (*model.UserDevice, error) {
	if _, err := s.repo.GetByID(ctx, device.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, device); err != nil {
		return nil, err
	}
	return device, nil
}

func (s *userDeviceService) DeleteDevice(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
