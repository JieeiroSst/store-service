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
	if err := goose.SetDialect("mysql"); err != nil {
		return err 
	}

	if err := goose.Up(m.db, "./script/migrations"); err != nil {
		return err 
	}
	return nil 
}