package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"layout/server/backends/memory"
	"layout/server/graphql/generated"
	"layout/server/graphql/resolvers"
)

func startServer() {
	args := memory.Credentials{}
	backend, err := memory.Backend(&args)
	if err != nil {
		fmt.Printf("Error creating testing backend: %v", err)
		os.Exit(1)
	}

	topResolver := resolvers.Resolver{Backend: backend}
	config := generated.Config{Resolvers: &topResolver}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(config))

	// Ingest additional test data in a go-routine.
	port := flags.playgroundPort
	if flags.addTestData {
		go ingestData(port)
	}

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%d/ for GraphQL playground", port)
	// blocking until termination
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
