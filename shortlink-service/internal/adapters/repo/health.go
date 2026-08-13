package repo

import (
	"context"

	"gorm.io/gorm"
)

type DBPing struct{ db *gorm.DB }

func NewDBPing(db *gorm.DB) *DBPing { return &DBPing{db: db} }

func (p *DBPing) Ping(ctx context.Context) error {
	return p.db.WithContext(ctx).Exec("SELECT 1").Error
}
