package server

import (
	"fmt"
	"log"
	"os"

	"github.com/JIeeiroSst/point-service/application"
	"github.com/JIeeiroSst/point-service/config"
	middleware "github.com/JIeeiroSst/point-service/interfaces"
	"github.com/JIeeiroSst/point-service/pkg/consul"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files" // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
)

type ServerImpl struct {
	router *gin.Engine
}

func InitServer() Server {
	serverImpl := &ServerImpl{}
	router := gin.Default()

	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {

	}

	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {

	}

	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.Mysql.MysqlUser,
		conf.Mysql.MysqlPassword,
		conf.Mysql.MysqlHost,
		conf.Mysql.MysqlPort,
		conf.Mysql.MysqlDbname,
	)

	router.Use(middleware.CORSMiddleware())
	swaggerDocs()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	application.InitRewardDiscountRouter(router, dns)
	application.InitConvertedRewardPointRouter(router, dns)
	application.InitRewardPointRouter(router, dns)

	serverImpl.router = router
	return serverImpl
}

func swaggerDocs() {
	docs.SwaggerInfo.Title = os.Getenv("SWAGGER_TITLE")
	docs.SwaggerInfo.Description = os.Getenv("SWAGGER_DESCRIPTION")
	docs.SwaggerInfo.Version = os.Getenv("SWAGGER_VERSION")
	docs.SwaggerInfo.Host = os.Getenv("SWAGGER_HOST")
	docs.SwaggerInfo.BasePath = os.Getenv("SWAGGER_BASEPATH")
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}

func (api *ServerImpl) RunServer() {
	appPort := os.Getenv("PORT")
	if appPort == "" {
		appPort = os.Getenv("LOCAL_PORT") //localhost
	}
	log.Fatal(api.router.Run(":" + appPort))
}
