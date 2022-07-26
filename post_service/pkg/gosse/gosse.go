package gosse

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func RunMigration(db *sql.DB) {
	if err := goose.SetDialect("mysql"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "./script/migrations"); err != nil {
		panic(err)
	}
}
