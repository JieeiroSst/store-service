package main

import (
	"encoding/json"
	"fmt"

	"github.com/JIeeiroSst/kitchen-service/config"
	"github.com/JIeeiroSst/kitchen-service/internal/model"
	"github.com/JIeeiroSst/kitchen-service/pkg/postgres"
	"github.com/JieeiroSst/logger"
)

func main() {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.ConfigZap().Error(err.Error())
	}
	consul := logger.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	var config config.Config
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.ConfigZap().Error(err.Error())
	}

	if err := json.Unmarshal(conf, &config); err != nil {
		logger.ConfigZap().Error(err.Error())
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.Postgres.PostgresqlHost, config.Postgres.PostgresqlUser, config.Postgres.PostgresqlPassword,
		config.Postgres.PostgresqlDbname, config.Postgres.PostgresqlPort)

	postgresConn := postgres.NewPostgresConn(dsn)
	postgresConn.AutoMigrate(&model.Food{}, &model.Category{}, &model.Kitchen{})
}
