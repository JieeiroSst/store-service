package app

import (
	"fmt"
	"log"
	"os"

	"github.com/JIeeiroSst/post-service/config"
	"github.com/JIeeiroSst/post-service/internal/delivery/http"
	"github.com/JIeeiroSst/post-service/internal/repository"
	"github.com/JIeeiroSst/post-service/internal/usecase"
	"github.com/JIeeiroSst/post-service/pkg/mysql"
	"github.com/JIeeiroSst/post-service/pkg/snowflake"
	"github.com/gin-gonic/gin"
)

func NewApp(router *gin.Engine) {
	fmt.Println("Wellcome Server Authorize")
	conf, err := config.ReadConf("config.yml")
	if err != nil {
		log.Println(err)
	}
	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.Mysql.MysqlUser,
		conf.Mysql.MysqlPassword,
		conf.Mysql.MysqlHost,
		conf.Mysql.MysqlPort,
		conf.Mysql.MysqlDbname,
	)
	mysqlOrm := mysql.NewMysqlConn(dns)
	snowflake := snowflake.NewSnowflake()

	repository := repository.NewRepositories(mysqlOrm)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:     repository,
		Snowflake: snowflake,
	})

	http := http.NewHandler(*usecase)
	http.Init(router)

	port := os.Getenv("PORT")

	if port == "" {
		port = conf.Server.PortServer
	}
	router.Run(":" + port)
}
