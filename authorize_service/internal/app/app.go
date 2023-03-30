package app

import (
	"fmt"
	"net"
	"os"

	"github.com/JieeiroSst/authorize-service/config"
	grpcServer "github.com/JieeiroSst/authorize-service/internal/delivery/gprc"
	"github.com/JieeiroSst/authorize-service/internal/delivery/http"
	"github.com/JieeiroSst/authorize-service/internal/pb"
	"github.com/JieeiroSst/authorize-service/internal/repository"
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/JieeiroSst/authorize-service/middleware"
	"github.com/JieeiroSst/authorize-service/pkg/cache"
	"github.com/JieeiroSst/authorize-service/pkg/goose"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/JieeiroSst/authorize-service/pkg/mysql"
	"github.com/JieeiroSst/authorize-service/pkg/otp"
	"github.com/JieeiroSst/authorize-service/pkg/snowflake"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type App struct {
	config *config.Config
}

func NewApp(config *config.Config) *App {
	return &App{
		config: config,
	}
}

func (a *App) NewServerApp(router *gin.Engine) {
	fmt.Println("Wellcome Server Authorize")
	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		a.config.Mysql.MysqlUser,
		a.config.Mysql.MysqlPassword,
		a.config.Mysql.MysqlHost,
		a.config.Mysql.MysqlPort,
		a.config.Mysql.MysqlDbname,
	)
	mysqlOrm := mysql.NewMysqlConn(dns)

	db, err := mysqlOrm.DB()
	if err != nil {
		log.Error(err.Error())
	}

	migration := goose.NewMigration(db)
	if err := migration.RunMigration(); err != nil {
		log.Error(err.Error())
	}

	adapter, err := gormadapter.NewAdapterByDB(mysqlOrm)
	if err != nil {
		log.Error(err.Error())
	}

	middleware := middleware.Newmiddleware(a.config.Secret.AuthorizeKey)

	var snowflakeData = snowflake.NewSnowflake()
	var otp = otp.NewOtp(a.config.Secret.JwtSecretKey)
	var cache = cache.NewCacheHelper(a.config.Cache.Host)

	repository := repository.NewRepositories(mysqlOrm)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:       repository,
		Snowflake:   snowflakeData,
		Adapter:     adapter,
		OTP:         otp,
		CacheHelper: cache,
	})

	http := http.NewHandler(*usecase, middleware, adapter)

	http.Init(router)

	port := os.Getenv("PORT")

	if port == "" {
		port = a.config.Server.ServerPort
	}
	log.Info("server starting" + a.config.Server.ServerPort)
	router.Run(":" + port)
}

func (a *App) NewGRPCServer() {
	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		a.config.Mysql.MysqlUser,
		a.config.Mysql.MysqlPassword,
		a.config.Mysql.MysqlHost,
		a.config.Mysql.MysqlPort,
		a.config.Mysql.MysqlDbname,
	)
	mysqlOrm := mysql.NewMysqlConn(dns)

	adapter, err := gormadapter.NewAdapterByDB(mysqlOrm)
	if err != nil {
		log.Error(err.Error())
	}

	var snowflakeData = snowflake.NewSnowflake()
	var otp = otp.NewOtp(a.config.Secret.JwtSecretKey)
	var cache = cache.NewCacheHelper(a.config.Cache.Host)
	repository := repository.NewRepositories(mysqlOrm)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:       repository,
		Snowflake:   snowflakeData,
		Adapter:     adapter,
		OTP:         otp,
		CacheHelper: cache,
	})

	s := grpc.NewServer()
	srv := &grpcServer.GRPCServer{}
	srv.NewGRPCServer(usecase)
	pb.RegisterAuthorizeServer(s, srv)
	log.Info("getway starting" + a.config.Server.GRPCServer)
	l, err := net.Listen("tcp", fmt.Sprintf(":%v", a.config.Server.GRPCServer))
	if err != nil {
		log.Error(err.Error())
	}
	if err := s.Serve(l); err != nil {
		log.Error(err.Error())
	}
}
