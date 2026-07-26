package database

import (
	"fmt"

	"github.com/JieeiroSst/banking-service/config"
	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Mysql.MysqlUser,
		cfg.Mysql.MysqlPassword,
		cfg.Mysql.MysqlHost,
		cfg.Mysql.MysqlPort,
		cfg.Mysql.MysqlDbname,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.Person{},
		&model.Branch{},
		&model.Customer{},
		&model.Employee{},
		&model.Account{},
		&model.Loan{},
		&model.LoanPayment{},
		&model.Transaction{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
)
