package server

import (
	"net/http"

	"github.com/cockroachlabs/wikifeedia/db"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/graphiql"
	"github.com/samsarahq/thunder/graphql/introspection"
	"github.com/samsarahq/thunder/graphql/schemabuilder"
)

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
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// schema builds the graphql schema.
func (s *Server) schema() *graphql.Schema {
	builder := schemabuilder.NewSchema()
	obj := builder.Object("Article", db.Article{})
	obj.Key("article")
	q := builder.Query()
	q.FieldFunc("articles", s.db.GetAllArticles)
	mut := builder.Mutation()
	mut.FieldFunc("echo", func(args struct{ Message string }) string {
		return args.Message
	})
	return builder.MustBuild()
}
