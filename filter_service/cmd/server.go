package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/JIeeiroSst/filter-service/config"
	"github.com/JIeeiroSst/filter-service/graph/dataloader"
	"github.com/JIeeiroSst/filter-service/graph/generated"
	"github.com/JIeeiroSst/filter-service/graph/resolver"
	"github.com/JIeeiroSst/filter-service/graph/storage"
	"github.com/JIeeiroSst/filter-service/pkg/consul"
)

var (
	conf        *config.Config
	dirEnv      *config.Dir
	err         error
	defaultPort = "8080"
	port        string
)

func main() {

	nodeEnv := os.Getenv("production")

	dirEnv, err = config.ReadFileEnv(".env")
	if err != nil {
		log.Println(err)
	}

	if !strings.EqualFold(nodeEnv, "") {
		consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
		conf, err = consul.ConnectConfigConsul()
		if err != nil {
			log.Println(err)
		}
		port = conf.Server.ServerPort
	} else {
		port = defaultPort
	}

	db := storage.NewMemoryStorage()

	loader := dataloader.NewDataLoader(db)

	graphResolver := resolver.NewResolver(db)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: graphResolver}))

	dataloaderSrv := dataloader.Middleware(loader, srv)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", dataloaderSrv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
