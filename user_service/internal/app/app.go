package app

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/JIeeiroSst/user-service/config"
	grpcServer "github.com/JIeeiroSst/user-service/internal/delivery/grpc"
	"github.com/JIeeiroSst/user-service/internal/delivery/http"
	"github.com/JIeeiroSst/user-service/internal/pb"
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/internal/usecase"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/JIeeiroSst/user-service/pkg/hash"
	"github.com/JIeeiroSst/user-service/pkg/postgres"
	"github.com/JIeeiroSst/user-service/pkg/snowflake"
	"github.com/JIeeiroSst/user-service/pkg/token"
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

func (a *App) NewUserApp(router *gin.Engine) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		a.config.Postgres.PostgresqlHost, a.config.Postgres.PostgresqlUser, a.config.Postgres.PostgresqlPassword,
		a.config.Postgres.PostgresqlDbname, a.config.Postgres.PostgresqlPort)

	postgresConn := postgres.NewPostgresConn(dsn)
	postgresConn.AutoMigrate(&model.Users{}, &model.Role{})

	snowflake := snowflake.NewSnowflake()
	hash := hash.NewHash()
	token := token.NewToken(a.config)

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
		port = a.config.Server.PortServer
	}
	router.Run(":" + port)
}

func (a *App) NewServerGrpc() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		a.config.Postgres.PostgresqlHost, a.config.Postgres.PostgresqlUser, a.config.Postgres.PostgresqlPassword,
		a.config.Postgres.PostgresqlDbname, a.config.Postgres.PostgresqlPort)

	postgresConn := postgres.NewPostgresConn(dsn)
	postgresConn.AutoMigrate(&model.Users{}, &model.Role{})

	snowflake := snowflake.NewSnowflake()
	hash := hash.NewHash()
	token := token.NewToken(a.config)

	repository := repository.NewRepositories(postgresConn)

	usecase := usecase.NewUsecase(usecase.Dependency{
		Repos:     repository,
		Snowflake: snowflake,
		Hash:      hash,
		Token:     token,
	})

	s := grpc.NewServer()
	srv := &grpcServer.GRPCServer{}
	srv.NewGRPCServer(*usecase)
	pb.RegisterAuthenticationServer(s, srv)
	l, err := net.Listen("tcp", a.config.Server.PortServerGrpc)
	if err != nil {
		log.Println(err)
	}
	if err := s.Serve(l); err != nil {
		log.Println(err)
	}

}
