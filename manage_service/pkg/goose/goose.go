package goose

import (
	"database/sql"
	"embed"

	"github.com/JIeeiroSst/manage-service/pkg/log"
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

func (m *Migration) RunMigration(embedMigrations embed.FS) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("mysql"); err != nil {
		log.Error(err)
		return err
	}

	err := goose.Up(m.db, "migrations", goose.WithAllowMissing())
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
