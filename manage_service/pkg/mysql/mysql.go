package mysql

import (
	"fmt"
	"sync"

	"github.com/JIeeiroSst/manage-service/pkg/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	instance *MysqlConnect
	once     sync.Once
)

type MysqlConnect struct {
	db *gorm.DB
}

func GetMysqlConnInstance(dns string) *MysqlConnect {
	once.Do(func() {
		db, err := gorm.Open(mysql.Open(dns), &gorm.Config{})
		if err != nil {
			log.Error(err)
			return
		}
		stmt := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %v;", "db")
		if rs := db.Exec(stmt); rs.Error != nil {
			log.Error(err)
			return
		}

		instance = &MysqlConnect{db: db}
	})
	return instance
}

func NewMysqlConn(dns string) *gorm.DB {
	return GetMysqlConnInstance(dns).db
}
