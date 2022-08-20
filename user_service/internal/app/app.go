package app

import (
	"fmt"
	"log"
	"os"

	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/internal/delivery/http"
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/internal/usecase"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/user-service/pkg/hash"
	"github.com/JIeeiroSst/user-service/pkg/postgres"
	"github.com/JIeeiroSst/user-service/pkg/snowflake"
	"github.com/JIeeiroSst/user-service/pkg/token"
	"github.com/gin-gonic/gin"
)

func NewApp(router *gin.Engine) {
	dir := "config.yml"
	conf, err := config.ReadConf(dir)
	if err != nil {
		log.Fatal(err)
	}

	dsn:=fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		conf.Postgres.PostgresqlHost,conf.Postgres.PostgresqlUser,conf.Postgres.PostgresqlPassword,
		conf.Postgres.PostgresqlDbname,conf.Postgres.PostgresqlPort)

	postgresConn:= postgres.NewPostgresConn(dsn)
	postgresConn.AutoMigrate(&model.Users{}, &model.Role{})

	snowflake := snowflake.NewSnowflake()
	hash := hash.NewHash()
	token := token.NewToken(conf)

	repository := repository.NewRepositories(postgresConn)

	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:     repository,
		Snowflake: snowflake,
		Hash:      hash,
		Token:     token,
	})

	http := http.NewHandler(*usecase)

	http.Init(router)

	port := os.Getenv("PORT")

	if port == "" {
		port = conf.Server.PortServer
	}
	router.Run(":" + port)
}
