package app

import (
	"fmt"
	"log"
	"os"

	"github.com/JieeiroSst/authorize-service/config"
	"github.com/JieeiroSst/authorize-service/internal/delivery/http"
	"github.com/JieeiroSst/authorize-service/internal/repository"
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/JieeiroSst/authorize-service/pkg/mysql"
	"github.com/JieeiroSst/authorize-service/pkg/otp"
	"github.com/JieeiroSst/authorize-service/pkg/snowflake"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
)

type App struct {
}

func NewApp(router *gin.Engine) {
	fmt.Println("Wellcome Server Authorize")
	conf, err := config.ReadConf("config.yml")
	if err != nil {
		log.Println(err)
	}
	//"test:test@(localhost:3306)/test"
	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.Mysql.MysqlUser,
		conf.Mysql.MysqlPassword,
		conf.Mysql.MysqlHost,
		conf.Mysql.MysqlPort,
		conf.Mysql.MysqlDbname,
	)
	mysqlOrm, err := mysql.NewMysqlConn(dns)
	if err != nil {
		log.Println(err)
	}
	adapter, err := gormadapter.NewAdapterByDB(mysqlOrm)
	if err != nil {
		log.Println(err)
	}

	var snowflakeData = snowflake.NewSnowflake()
	var otp = otp.NewOtp(conf.Secret.JwtSecretKey)

	repository := repository.NewRepositories(mysqlOrm)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:     repository,
		Snowflake: snowflakeData,
		Adapter:   adapter,
		OTP:       otp,
	})

	http := http.NewHandler(*usecase, adapter)

	http.Init(router)

	port := os.Getenv("PORT")

	if port == "" {
		port = conf.Server.ServerPort
	}
	router.Run(":" + port)
}
