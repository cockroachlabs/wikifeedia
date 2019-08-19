package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cockroachlabs/wikifeedia/db"
	"github.com/cockroachlabs/wikifeedia/wikipedia"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/graphiql"
	"github.com/samsarahq/thunder/graphql/introspection"
	"github.com/samsarahq/thunder/graphql/schemabuilder"
)

//go:generate go run github.com/shurcooL/vfsgen/cmd/vfsgendev -source="github.com/cockroachlabs/wikifeedia/server".Assets

// Server is an http.Handler for a graphql server for this application.
type Server struct {
	db  *db.DB
	mux http.ServeMux
}

// New creates a new Server.
func New(conn *db.DB) *Server {
	s := &Server{
		db: conn,
	}
	schema := s.schema()

	introspection.AddIntrospectionToSchema(schema)

	s.mux.Handle("/graphqlhttp", graphql.HTTPHandler(schema))
	s.mux.Handle("/graphql", graphql.Handler(schema))
	s.mux.Handle("/graphiql/", http.StripPrefix("/graphiql/", graphiql.Handler()))
	s.mux.Handle("/", http.FileServer(Assets))
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("could not write response: %v", err)
		}
	})
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	s.mux.ServeHTTP(w, r)
}

func (s *Server) getAllArticles(
	ctx context.Context, args struct {
		Project string `json:"project"`
	},
) ([]db.Article, error) {
	if !wikipedia.IsProject(args.Project) {
		return nil, fmt.Errorf("%s is not a valid project")
	}
	return s.db.GetAllArticles(ctx, args.Project)
}

// schema builds the graphql schema.
func (s *Server) schema() *graphql.Schema {
	builder := schemabuilder.NewSchema()
	obj := builder.Object("Article", db.Article{})
	obj.Key("article")
	q := builder.Query()
	q.FieldFunc("articles", s.getAllArticles)
	mut := builder.Mutation()
	mut.FieldFunc("echo", func(args struct{ Message string }) string {
		return args.Message
	})
	return builder.MustBuild()
}
