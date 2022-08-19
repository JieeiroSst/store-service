package app

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/JieeiroSst/authorize-service/config"
	grpcServer "github.com/JieeiroSst/authorize-service/internal/delivery/gprc"
	"github.com/JieeiroSst/authorize-service/internal/delivery/http"
	"github.com/JieeiroSst/authorize-service/internal/pb"
	"github.com/JieeiroSst/authorize-service/internal/repository"
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/JieeiroSst/authorize-service/pkg/goose"
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
		log.Println(err)
	}

	migration := goose.NewMigration(db)
	if err := migration.RunMigration(); err != nil {
		log.Println(err)
	}

	adapter, err := gormadapter.NewAdapterByDB(mysqlOrm)
	if err != nil {
		log.Println(err)
	}

	var snowflakeData = snowflake.NewSnowflake()
	var otp = otp.NewOtp(a.config.Secret.JwtSecretKey)

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
		port = a.config.Server.ServerPort
	}
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
		log.Println(err)
	}

	var snowflakeData = snowflake.NewSnowflake()
	var otp = otp.NewOtp(a.config.Secret.JwtSecretKey)
	repository := repository.NewRepositories(mysqlOrm)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:     repository,
		Snowflake: snowflakeData,
		Adapter:   adapter,
		OTP:       otp,
	})

	s := grpc.NewServer()
	srv := &grpcServer.GRPCServer{}
	srv.NewGRPCServer(usecase)
	pb.RegisterAuthorizeServer(s, srv)
	l, err := net.Listen("tcp", a.config.Server.GRPCServer)
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Serve(l); err != nil {
		log.Fatal(err)
	}
}
