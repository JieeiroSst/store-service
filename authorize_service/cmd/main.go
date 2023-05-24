package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JieeiroSst/authorize-service/config"
	grpcServer "github.com/JieeiroSst/authorize-service/internal/delivery/gprc"
	http1 "github.com/JieeiroSst/authorize-service/internal/delivery/http"
	"github.com/JieeiroSst/authorize-service/internal/pb"
	"github.com/JieeiroSst/authorize-service/internal/repository"
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/JieeiroSst/authorize-service/middleware"
	"github.com/JieeiroSst/authorize-service/pkg/cache"
	"github.com/JieeiroSst/authorize-service/pkg/consul"
	"github.com/JieeiroSst/authorize-service/pkg/goose"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/JieeiroSst/authorize-service/pkg/mysql"
	"github.com/JieeiroSst/authorize-service/pkg/otp"
	"github.com/JieeiroSst/authorize-service/pkg/snowflake"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

var (
	conf   *config.Config
	dirEnv *config.Dir
	err    error
)

func main() {
	router := gin.Default()
	nodeEnv := os.Getenv("NODE_ENV")

	log.Info("nodeEnv is " + nodeEnv)

	dirEnv, err = config.ReadFileEnv(".env")
	if err != nil {
		log.Error(err.Error())
	}

	if !strings.EqualFold(nodeEnv, "") {
		consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
		conf, err = consul.ConnectConfigConsul()
		if err != nil {
			log.Error(err.Error())
		}
	} else {
		conf, err = config.ReadConf("config.yml")
		if err != nil {
			log.Error(err.Error())
		}
	}

	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.Mysql.MysqlUser,
		conf.Mysql.MysqlPassword,
		conf.Mysql.MysqlHost,
		conf.Mysql.MysqlPort,
		conf.Mysql.MysqlDbname,
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

	middleware := middleware.Newmiddleware(conf.Secret.AuthorizeKey)

	var snowflakeData = snowflake.NewSnowflake()
	var otp = otp.NewOtp(conf.Secret.JwtSecretKey)
	var cache = cache.NewCacheHelper(conf.Cache.Host)

	repository := repository.NewRepositories(mysqlOrm)
	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:       repository,
		Snowflake:   snowflakeData,
		Adapter:     adapter,
		OTP:         otp,
		CacheHelper: cache,
	})

	httpServer := http1.NewHandler(*usecase, middleware, adapter)

	httpServer.Init(router)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v",conf.Server.ServerPort),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Info(fmt.Sprintf("listen: %s\n", err))
		}
	}()

	go func() {
		s := grpc.NewServer()
		srv := &grpcServer.GRPCServer{}
		srv.NewGRPCServer(usecase)
		pb.RegisterAuthorizeServer(s, srv)
		log.Info("getway starting" + conf.Server.GRPCServer)
		l, err := net.Listen("tcp", fmt.Sprintf(":%v", conf.Server.GRPCServer))
		if err != nil {
			log.Error(err.Error())
		}
		if err := s.Serve(l); err != nil {
			log.Error(err.Error())
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(fmt.Sprintf("Server Shutdown: %v", err))
	}
	select {
	case <-ctx.Done():
		log.Info("timeout of 5 seconds.")
	}
	log.Info("Server exiting")
}
