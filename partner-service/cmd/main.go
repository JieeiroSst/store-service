package main

import (
	"fmt"

	"github.com/JIeeiroSst/partner-service/internal/adapters/cache"
	"github.com/JIeeiroSst/partner-service/internal/adapters/repository"
	"github.com/JIeeiroSst/partner-service/internal/config"
	"github.com/JIeeiroSst/partner-service/internal/consul"
	"github.com/JIeeiroSst/partner-service/internal/core/services"
	"github.com/JIeeiroSst/partner-service/internal/logger"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	partnershipService         *services.PartnershipService
	partnershipsPartnerService *services.PartnershipsPartnerService
	partnerService             *services.PartnerService
	projectService             *services.ProjectService
)

func main() {
	logger.SetupLogger()
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.Log.Error(err)
	}

	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.Log.Error(err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Mysql.MysqlHost, conf.Mysql.MysqlPort, conf.Mysql.MysqlUser, conf.Mysql.MysqlPassword, conf.Mysql.MysqlDbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	redisCache, err := cache.NewRedisCache(conf.Cache.Host, "")
	if err != nil {
		panic(err)
	}

	store := repository.NewDB(db, redisCache)

	partnershipService = services.NewPartnershipService(store)
	partnershipsPartnerService = services.NewPartnershipsPartnerService(store)
	partnerService = services.NewPartnerService(store)
	projectService = services.NewProjectService(store)

	InitRoutes(conf)
}

func InitRoutes(conf *config.Config) {
	router := gin.Default()

	pprof.Register(router)

	err := router.Run(conf.Server.ServerPort)
	if err != nil {
		logger.Log.Error(err)
	}
}
