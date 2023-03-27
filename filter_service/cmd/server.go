package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/JIeeiroSst/filter-service/graph/dataloader"
	"github.com/JIeeiroSst/filter-service/graph/generated"
	"github.com/JIeeiroSst/filter-service/graph/resolver"
	"github.com/JIeeiroSst/filter-service/graph/storage"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
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
