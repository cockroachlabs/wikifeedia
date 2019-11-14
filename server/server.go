package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
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
	fs := http.FileServer(Assets)
	s.mux.Handle("/graphqlhttp",
		gziphandler.GzipHandler(graphql.HTTPHandler(schema)))
	s.mux.Handle("/graphql", graphql.Handler(schema))
	s.mux.Handle("/graphiql/", http.StripPrefix("/graphiql/", graphiql.Handler()))
	staticHandler := gziphandler.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fs.ServeHTTP(w, r)
	}))
	s.mux.Handle("/", staticHandler)
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("could not write response: %v", err)
		}
	})
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type ArticlesResponse struct {
	Articles []db.Article `json:"articles"`
	AsOf     string       `json:"as_of"`
}

func (s *Server) getArticles(
	ctx context.Context,
	args struct {
		Project      string
		Offset       int32
		Limit        int32
		FollowerRead *bool
		AsOf         *string
	},
) (*ArticlesResponse, error) {
	fmt.Println("h")
	start := time.Now()
	var asOf string
	if args.AsOf != nil {
		asOf = *args.AsOf
	}
	defer func() {
		log.Printf("%v?limit=%v&offset=%v&follower_read=%v&as_of=%v - %v",
			args.Project, args.Limit, args.Offset, *args.FollowerRead,
			time.Since(start), asOf)
	}()
	if !wikipedia.IsProject(args.Project) {
		return nil, fmt.Errorf("%s is not a valid project")
	}
	articles, newAsOf, err := s.db.GetArticles(ctx, args.Project, int(args.Offset), int(args.Limit),
		args.FollowerRead != nil && *args.FollowerRead, asOf)
	if err != nil {
		return nil, err
	}
	return &ArticlesResponse{
		AsOf:     newAsOf,
		Articles: articles,
	}, nil
}

// schema builds the graphql schema.
func (s *Server) schema() *graphql.Schema {
	builder := schemabuilder.NewSchema()
	obj := builder.Object("Article", db.Article{})
	obj.Key("article")
	builder.Object("ArticlesResponse", ArticlesResponse{})
	q := builder.Query()
	q.FieldFunc("articles", s.getArticles)
	mut := builder.Mutation()
	mut.FieldFunc("echo", func(args struct{ Message string }) string {
		return args.Message
	})
	return builder.MustBuild()
}
