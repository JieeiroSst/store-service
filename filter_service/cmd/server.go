package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/JIeeiroSst/filter-service/adapter"
	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/config"
	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/resolver"
	"github.com/JIeeiroSst/filter-service/pkg/consul"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	if !strings.EqualFold(os.Getenv("production"), "") {
		dirEnv, err := config.ReadFileEnv(".env")
		if err != nil {
			log.Println(err)
		} else {
			conf, err := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul).ConnectConfigConsul()
			if err != nil {
				log.Println(err)
			} else if conf.Server.ServerPort != "" {
				port = conf.Server.ServerPort
			}
		}
	}

	clients := adapter.NewClients(&http.Client{Timeout: 15 * time.Second})
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver.NewResolver(clients)}))
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("filter-service", "/query"))
	mux.Handle("/query", rest.ForwardHeaders(srv))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })

	log.Printf("GraphQL gateway for %d services on :%s", len(clients.Services), port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
