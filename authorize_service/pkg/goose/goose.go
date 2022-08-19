package goose

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

type Migration struct {
	db *sql.DB
}

func NewMigration(db *sql.DB) *Migration {
	return &Migration{
		db: db,
	}
}

func (m *Migration) RunMigration() error {
	err := goose.Up(m.db, "./script/migrations")
	if err != nil {
		return err
	}
	return nil
}
