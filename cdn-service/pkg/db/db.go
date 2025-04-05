package db

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

type DBInstance struct {
	DB *sql.DB
}

var (
	instance *DBInstance
	once     sync.Once
)

// connStr := "host=localhost port=5432 user=postgres password=yourpassword dbname=yourdb sslmode=disable"
func GetInstance(connectionString string) *DBInstance {
	once.Do(func() {
		db, err := sql.Open("postgres", connectionString)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		if err = db.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}

		instance = &DBInstance{DB: db}
		log.Println("Database connection established successfully")
	})

	return instance
}

func (d *DBInstance) Close() {
	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
			return
		}
		log.Println("Database connection closed successfully")
	}
}
