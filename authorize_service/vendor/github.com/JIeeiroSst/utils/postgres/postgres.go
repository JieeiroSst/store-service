package postgres

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
	PgDriver           string
}

var (
	instance *PostgresConnect
	once     sync.Once
)

type PostgresConnect struct {
	db *gorm.DB
}

func GetPostgresConnInstance(dns string) *PostgresConnect {
	once.Do(func() {
		db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
		if err != nil {
			panic(err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			panic(err)
		}

		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(30 * time.Minute)

		instance = &PostgresConnect{db: db}
	})
	return instance
}

func NewPostgresConn(postgres PostgresConfig) *gorm.DB {
	dns := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		postgres.PostgresqlHost, postgres.PostgresqlUser, postgres.PostgresqlPassword, postgres.PostgresqlDbname, postgres.PostgresqlPort)
	return GetPostgresConnInstance(dns).db
}
