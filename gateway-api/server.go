package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/JIeeiroSst/gateway-api/config"
	"github.com/JIeeiroSst/gateway-api/graph"
	"github.com/JIeeiroSst/gateway-api/media"
	"github.com/JIeeiroSst/gateway-api/middlware"
	"github.com/JIeeiroSst/gateway-api/proxy"
	"github.com/JIeeiroSst/utils/consul"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
)

func graphqlHandler(p *proxy.Proxy) gin.HandlerFunc {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers:  &graph.Resolver{Proxy: p},
		Directives: graph.NewDirectiveRoot(p.Authorize),
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxUploadSize: 32 << 20,
		MaxMemory:     8 << 20,
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.FixedComplexityLimit(1000))
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	srv.AroundOperations(graph.LimitRootFields)
	srv.AroundRootFields(graph.RequireAuthentication)

	return func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	router := gin.Default()
	router.Use(middlware.NewMiddlware())
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.Error(context.Background(), "error %v", err)
	}
	consul := consul.NewConfigConsul(dirEnv.HostConsul,
		dirEnv.KeyConsul, dirEnv.ServiceConsul)
	var config config.Config
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		logger.Error(context.Background(), "error %v", err)
	}
	if err := json.Unmarshal(conf, &config); err != nil {
		logger.Error(context.Background(), "error %v", err)
	}
	p, err := proxy.NewProxy(config)
	if err != nil {
		logger.Error(context.Background(), "error %v", err)
		os.Exit(1)
	}
	defer p.Close()

	router.POST("/query", middlware.ClientInfo(), middlware.Authenticate(p.User), graphqlHandler(p))
	router.GET("/", playgroundHandler())
	mediaHandler := gin.WrapH(media.Handler("/media/", func(service string) string { return p.REST(service).BaseURL() }))
	router.GET("/media/*path", mediaHandler)
	router.HEAD("/media/*path", mediaHandler)
	router.GET("/chat/ws", middlware.Authenticate(p.User), gin.WrapH(media.Socket("ollama-service", "/ws", func(service string) string { return p.REST(service).BaseURL() })))
	router.GET("/rtc/ws", middlware.Authenticate(p.User), gin.WrapH(media.Signaling("webrtc-service", func(service string) string { return p.REST(service).BaseURL() })))

	httpSrv := &http.Server{
		Addr:              fmt.Sprintf(":%v", config.Server.PortServer),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Info(context.Background(), "listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(context.Background(), "Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Info(context.Background(), "Server Shutdown: %v", err)
	}
	select {
	case <-ctx.Done():
		logger.Info(context.Background(), "timeout of 5 seconds.")
	}
	logger.Info(context.Background(), "Server exiting")
}
