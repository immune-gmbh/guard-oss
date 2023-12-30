package debugv1

import (
	"context"
	"net/http"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/immune-gmbh/guard/apisrv/v2/internal/graphql"
)

type WithDatabase struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, opts ...interface{}) (http.Handler, error) {
	var pool *pgxpool.Pool

	for _, opt := range opts {
		switch opt := opt.(type) {
		case WithDatabase:
			pool = opt.Pool

		default:
		}
	}

	// gql handler
	schema := graphql.NewExecutableSchema(graphql.Config{
		Resolvers:  &Resolver{pool},
		Directives: graphql.DirectiveRoot{},
		Complexity: graphql.ComplexityRoot{},
	})
	srv := gqlhandler.NewDefaultServer(schema)
	srv.Use(extension.FixedComplexityLimit(300))

	// router
	router := mux.NewRouter()
	router.Methods("POST").
		Path("/debugv1/graphql").
		Handler(srv)
	router.Methods("GET").
		Path("/debugv1").
		Handler(playground.Handler("api-gateway", "/debugv1/graphql"))

	return router, nil
}
