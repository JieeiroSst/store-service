package repository

import (
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = fx.Options(
	fx.Provide(NewNotificationRepository),
	fx.Provide(func(db *gorm.DB) *userDeviceRepository { return &userDeviceRepository{db: db} }),
	fx.Provide(func(r *userDeviceRepository) port.UserDeviceRepository { return r }),
	fx.Provide(func(r *userDeviceRepository) port.DeviceDeactivator { return r }),
	fx.Provide(NewContactRepository),
	fx.Provide(NewAuditRepository),
	fx.Provide(NewCampaignRepository),
	fx.Provide(NewCampaignBatchRepository),
	fx.Provide(NewCampaignTargetSource),
)
