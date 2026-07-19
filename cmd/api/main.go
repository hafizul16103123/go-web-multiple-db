package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mongoopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"webapp/internal/config"
	"webapp/internal/handler"
	"webapp/internal/repository"
	"webapp/internal/repository/memory"
	"webapp/internal/repository/mongodb"
	"webapp/internal/repository/sqldb"
	"webapp/internal/service"
)

// repos bundles both repositories so mustInitRepo can return them together.
type repos struct {
	post   repository.PostRepository
	author repository.AuthorRepository
}

func main() {
	cfg := config.Load()

	r := mustInitRepos(cfg)

	authorSvc := service.NewAuthorService(r.author)
	postSvc := service.NewPostService(r.post, r.author)

	mux := http.NewServeMux()
	handler.NewAuthorHandler(authorSvc).RegisterRoutes(mux)
	handler.NewPostHandler(postSvc).RegisterRoutes(mux)

	log.Printf("server listening on %s  [db=%s]", cfg.ServerAddr, cfg.DBDriver)
	if err := http.ListenAndServe(cfg.ServerAddr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// mustInitRepos selects and initialises the correct repositories based on DB_DRIVER.
// It exits the process immediately if the database cannot be reached.
func mustInitRepos(cfg *config.Config) repos {
	switch cfg.DBDriver {
	case "sql":
		db, err := sql.Open("postgres", cfg.SQLDSN)
		if err != nil {
			log.Fatalf("sql: open failed: %v", err)
		}
		if err := db.Ping(); err != nil {
			log.Fatalf("sql: ping failed — is PostgreSQL running?\n  DSN: %s\n  error: %v", cfg.SQLDSN, err)
		}
		// Authors table must exist before posts (posts.author_id references it).
		authorRepo, err := sqldb.NewAuthorRepository(db)
		if err != nil {
			log.Fatalf("sql: authors migration failed: %v", err)
		}
		postRepo, err := sqldb.New(db)
		if err != nil {
			log.Fatalf("sql: posts migration failed: %v", err)
		}
		log.Println("connected to PostgreSQL")
		return repos{post: postRepo, author: authorRepo}

	case "mongo":
		client, err := mongo.Connect(mongoopts.Client().ApplyURI(cfg.MongoURI))
		if err != nil {
			log.Fatalf("mongo: connect failed: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Ping(ctx, nil); err != nil {
			log.Fatalf("mongo: ping failed — is MongoDB running?\n  URI: %s\n  error: %v", cfg.MongoURI, err)
		}
		db := client.Database(cfg.MongoDB)
		log.Println("connected to MongoDB")
		return repos{
			post:   mongodb.New(db),
			author: mongodb.NewAuthorRepository(db),
		}

	default:
		log.Printf("unknown DB_DRIVER %q — falling back to in-memory store", cfg.DBDriver)
		return repos{
			post:   memory.New(),
			author: memory.NewAuthorRepository(),
		}
	}
}
