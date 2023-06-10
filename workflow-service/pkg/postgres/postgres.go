package postgres

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

// connStr := "postgres://postgres:password@localhost/DB_1?sslmode=disable"
func NewDatabase(dns string) *Database {
	db, err := sql.Open("postgres", dns)
	if err != nil {
		log.Println(err)
	}
	if err = db.Ping(); err != nil {
		log.Println(err)
	}
	return &Database{
		DB: db,
	}
}
