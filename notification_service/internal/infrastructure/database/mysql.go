package database

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Mysql.MysqlUser,
		cfg.Mysql.MysqlPassword,
		cfg.Mysql.MysqlHost,
		cfg.Mysql.MysqlPort,
		cfg.Mysql.MysqlDbname,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.Notification{},
		&model.UserDevice{},
		&model.UserContact{},
		&model.Campaign{},
		&model.CampaignRecipient{},
		&model.CampaignBatch{},
		&model.AuditContent{},
		&model.AuditDelivery{},
	); err != nil {
		return nil, err
	}
	if err := backfillTokenHashes(db); err != nil {
		return nil, fmt.Errorf("backfill device token hashes: %w", err)
	}

	return db, nil
}

func backfillTokenHashes(db *gorm.DB) error {
	var lastID uint = ^uint(0)
	for {
		var devices []model.UserDevice
		err := db.Where("token_hash IS NULL AND id < ?", lastID).Order("id DESC").Limit(1000).Find(&devices).Error
		if err != nil || len(devices) == 0 {
			return err
		}
		for _, d := range devices {
			lastID = d.ID
			if strings.TrimSpace(d.DeviceToken) == "" {
				if err := db.Model(&model.UserDevice{}).Where("id = ?", d.ID).Update("is_active", false).Error; err != nil {
					return err
				}
				continue
			}
			hash := model.HashToken(d.DeviceToken)
			var owners int64
			if err := db.Model(&model.UserDevice{}).Where("token_hash = ?", hash).Count(&owners).Error; err != nil {
				return err
			}
			update := map[string]any{"token_hash": hash}
			if owners > 0 {
				update = map[string]any{"is_active": false}
			}
			if err := db.Model(&model.UserDevice{}).Where("id = ?", d.ID).Updates(update).Error; err != nil {
				log.Printf("backfill device %d: %v", d.ID, err)
			}
		}
	}
}

func registerLifecycle(lc fx.Lifecycle, db *gorm.DB) {
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
	fx.Invoke(registerLifecycle),
)
