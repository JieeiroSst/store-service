package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/JIeeiroSst/gateway-api/catalog"
	"github.com/JIeeiroSst/gateway-api/config"
	"github.com/JIeeiroSst/gateway-api/graph"
	"github.com/JIeeiroSst/gateway-api/middlware"
	"github.com/JIeeiroSst/utils/consul"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
)

func graphqlHandler(c *catalog.Catalog) gin.HandlerFunc {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Catalog: c}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("Store Service API Catalog", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	router := gin.Default()

	router.Use(middlware.NewMiddlware())

	var conf config.Config
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		logger.Error(context.Background(), "error %v", err)
	} else {
		raw, err := consul.NewConfigConsul(dirEnv.HostConsul,
			dirEnv.KeyConsul, dirEnv.ServiceConsul).ConnectConfigConsul()
		if err != nil {
			logger.Error(context.Background(), "error %v", err)
		} else if err := json.Unmarshal(raw, &conf); err != nil {
			logger.Error(context.Background(), "error %v", err)
		}
	}
	conf.ApplyDefaults()

	cat, err := catalog.Load()
	if err != nil {
		logger.Error(context.Background(), "load catalog: %v", err)
		os.Exit(1)
	}

	gql := graphqlHandler(cat)
	router.POST("/query", gql)
	router.GET("/query", gql)
	router.GET("/", playgroundHandler())
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "synced_at": cat.SyncedAt})
	})
	router.GET("/schema.graphql", func(c *gin.Context) {
		sdl, _ := cat.SDL("")
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(sdl))
	})
	router.GET("/schema/:service", func(c *gin.Context) {
		sdl, ok := cat.SDL(strings.TrimSuffix(c.Param("service"), ".graphqls"))
		if !ok {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(sdl))
	})

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", conf.Server.PortServer),
		Handler: router,
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
