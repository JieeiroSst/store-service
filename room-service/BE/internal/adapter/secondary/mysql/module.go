package mysql

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JIeeiroSst/room-service/config"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	"go.uber.org/fx"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Module = fx.Options(
	fx.Provide(
		NewDB,
		fx.Annotate(NewRoomRepository, fx.As(new(port.RoomRepository))),
		fx.Annotate(NewMessageRepository, fx.As(new(port.MessageRepository))),
	),
)

func NewDB(lc fx.Lifecycle, cfg *config.Config) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	for attempt := 1; attempt <= 15; attempt++ {
		db, err = gorm.Open(gormmysql.Open(cfg.MySQLDSN()), &gorm.Config{
			TranslateError: true,
			Logger: logger.New(log.New(os.Stderr, "\r\n", log.LstdFlags), logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
			}),
		})
		if err == nil {
			break
		}
		log.Printf("mysql not ready (attempt %d): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("mysql: %w", err)
	}

	if err := db.AutoMigrate(&roomRow{}, &memberRow{}, &messageRow{}); err != nil {
		return nil, fmt.Errorf("mysql migrate: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})
	return db, nil
}
