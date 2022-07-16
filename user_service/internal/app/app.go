package app

import (
	"fmt"

	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/internal/delivery/http"
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/internal/usecase"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/user-service/pkg/hash"
	"github.com/JIeeiroSst/user-service/pkg/mysql"
	"github.com/JIeeiroSst/user-service/pkg/snowflake"
	"github.com/JIeeiroSst/user-service/pkg/token"
	"github.com/gin-gonic/gin"
)

func NewApp(router *gin.Engine) {
	conf, err := config.ReadConf("config.yml")
	if err != nil {

	}

	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.Mysql.MysqlUser,
		conf.Mysql.MysqlPassword,
		conf.Mysql.MysqlHost,
		conf.Mysql.MysqlPort,
		conf.Mysql.MysqlDbname,
	)
	mysqlOrm := mysql.NewMysqlConn(dns)
	mysqlOrm.AutoMigrate(&model.Users{})

	snowflake := snowflake.NewSnowflake()
	hash := hash.NewHash()
	token := token.NewToken(conf)

	repository := repository.NewRepositories(mysqlOrm)

	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:     repository,
		Snowflake: snowflake,
		Hash:      hash,
		Token:     token,
	})

	http := http.NewHandler(*usecase)

	http.Init(router)

	router.Run(conf.Server.PortServer)
}
