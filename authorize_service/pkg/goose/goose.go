package goose

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

type Migration struct {
	db *sql.DB
}

var embedMigrations embed.FS

func NewMigration(db *sql.DB) *Migration {
	return &Migration{
		db: db,
	}
}

func (m *Migration) RunMigration() error {
	goose.SetBaseFS(embedMigrations)

    if err := goose.SetDialect("mysql"); err != nil {
        panic(err)
    }

	err := goose.Up(m.db, "script/migrations")
	if err != nil {
		return err
	}
	return nil
}
