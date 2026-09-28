package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *userDeviceRepository) Register(ctx context.Context, device *model.UserDevice, opts port.RegisterOptions) (*model.UserDevice, error) {
	var (
		saved *model.UserDevice
		err   error
	)
	for attempt := 0; attempt < 2; attempt++ {
		candidate := *device
		if saved, err = r.register(ctx, &candidate, opts); err == nil {
			return saved, nil
		}
	}
	return nil, err
}

func (r *userDeviceRepository) register(ctx context.Context, device *model.UserDevice, opts port.RegisterOptions) (*model.UserDevice, error) {
	var saved model.UserDevice
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.UserDevice
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token_hash = ?", *device.TokenHash).First(&existing).Error
		switch {
		case err == nil:
			if err := tx.Model(&model.UserDevice{}).Where("id = ?", existing.ID).Updates(map[string]any{
				"user_id":      device.UserID,
				"device_token": device.DeviceToken,
				"device_id":    device.DeviceID,
				"device_type":  device.DeviceType,
				"is_active":    true,
				"last_used_at": device.LastUsedAt,
			}).Error; err != nil {
				return err
			}
			device.ID, device.CreatedAt = existing.ID, existing.CreatedAt
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := tx.Create(device).Error; err != nil {
				return err
			}
		default:
			return err
		}

		if opts.ReplaceSameInstall {
			if err := tx.Model(&model.UserDevice{}).
				Where("device_id = ? AND id <> ? AND is_active = ?", device.DeviceID, device.ID, true).
				Update("is_active", false).Error; err != nil {
				return err
			}
		}
		if opts.SingleDevicePerUser {
			if err := tx.Model(&model.UserDevice{}).
				Where("user_id = ? AND id <> ? AND is_active = ?", device.UserID, device.ID, true).
				Update("is_active", false).Error; err != nil {
				return err
			}
		}
		return tx.First(&saved, device.ID).Error
	})
	if err != nil {
		return nil, common.ErrDBFailed
	}
	return &saved, nil
}

func (r *userDeviceRepository) DeactivateByTokenHash(ctx context.Context, tokenHash string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.UserDevice{}).Where("token_hash = ? AND is_active = ?", tokenHash, true).Update("is_active", false)
	if res.Error != nil {
		return 0, common.ErrDBFailed
	}
	return res.RowsAffected, nil
}

func (r *userDeviceRepository) DeactivateByUserID(ctx context.Context, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.UserDevice{}).Where("user_id = ? AND is_active = ?", userID, true).Update("is_active", false)
	if res.Error != nil {
		return 0, common.ErrDBFailed
	}
	return res.RowsAffected, nil
}

func (r *userDeviceRepository) DeactivateByIDs(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Model(&model.UserDevice{}).Where("id IN ?", ids).Update("is_active", false).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *userDeviceRepository) DeactivateStale(ctx context.Context, lastUsedBefore time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.UserDevice{}).
		Where("is_active = ? AND last_used_at < ?", true, lastUsedBefore).
		Update("is_active", false)
	if res.Error != nil {
		return 0, common.ErrDBFailed
	}
	return res.RowsAffected, nil
}

func (r *userDeviceRepository) DeactivateByDeviceID(ctx context.Context, userID uint, deviceID string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.UserDevice{}).
		Where("user_id = ? AND device_id = ? AND is_active = ?", userID, deviceID, true).
		Update("is_active", false)
	if res.Error != nil {
		return 0, common.ErrDBFailed
	}
	return res.RowsAffected, nil
}

func (r *userDeviceRepository) ListByUserID(ctx context.Context, userID uint) ([]model.UserDevice, error) {
	var devices []model.UserDevice
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&devices).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return devices, nil
}
