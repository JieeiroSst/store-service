package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"gorm.io/gorm"
)

type userDeviceRepository struct {
	db *gorm.DB
}

func NewUserDeviceRepository(db *gorm.DB) port.UserDeviceRepository {
	return &userDeviceRepository{db: db}
}

func (r *userDeviceRepository) Create(ctx context.Context, device *model.UserDevice) error {
	if err := r.db.WithContext(ctx).Create(device).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *userDeviceRepository) GetByID(ctx context.Context, id uint) (*model.UserDevice, error) {
	var device model.UserDevice
	if err := r.db.WithContext(ctx).First(&device, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &device, nil
}

func (r *userDeviceRepository) ListActiveByUserID(ctx context.Context, userID uint) ([]model.UserDevice, error) {
	var devices []model.UserDevice
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Find(&devices).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return devices, nil
}

func (r *userDeviceRepository) Update(ctx context.Context, device *model.UserDevice) error {
	if err := r.db.WithContext(ctx).Model(&model.UserDevice{}).Where("id = ?", device.ID).Updates(device).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *userDeviceRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.UserDevice{}, "id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *userDeviceRepository) List(ctx context.Context) ([]model.UserDevice, error) {
	var devices []model.UserDevice
	if err := r.db.WithContext(ctx).Find(&devices).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return devices, nil
}
