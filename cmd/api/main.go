package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"                          // side-effect import: registers "postgres" driver with database/sql
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

// repos bundles both repositories so mustInitRepos can return them together as one value.
type repos struct {
	post   repository.PostRepository   // interface — concrete type is chosen at startup based on DB_DRIVER
	author repository.AuthorRepository // interface — same concrete backend as post
}

func main() {
	cfg := config.Load() // reads DB_DRIVER, DSN, SERVER_ADDR from env; falls back to sensible defaults

	r := mustInitRepos(cfg) // connects to DB and returns concrete repo implementations behind the interfaces

	// services only see the repository interfaces, not the concrete DB type
	authorSvc := service.NewAuthorService(r.author)
	postSvc := service.NewPostService(r.post, r.author) // post service also needs author repo to validate author existence on create

	mux := http.NewServeMux() // stdlib request router; Go 1.22+ supports "METHOD /path" patterns
	handler.NewAuthorHandler(authorSvc).RegisterRoutes(mux) // mounts GET/POST/PUT/DELETE /api/v1/authors/*
	handler.NewPostHandler(postSvc).RegisterRoutes(mux)     // mounts GET/POST/PUT/DELETE /api/v1/posts/*

	log.Printf("server listening on %s  [db=%s]", cfg.ServerAddr, cfg.DBDriver)
	if err := http.ListenAndServe(cfg.ServerAddr, mux); err != nil { // blocks until the process is killed
		log.Fatalf("server error: %v", err)
	}
}

// mustInitRepos selects and initialises the correct repositories based on DB_DRIVER.
// It exits the process immediately (log.Fatalf) if the database cannot be reached —
// fail-fast at startup is safer than serving requests against a broken DB connection.
func mustInitRepos(cfg *config.Config) repos {
	switch cfg.DBDriver {
	case "sql":
		db, err := sql.Open("postgres", cfg.SQLDSN) // validates DSN format only; no network call yet
		if err != nil {
			log.Fatalf("sql: open failed: %v", err)
		}
		if err := db.Ping(); err != nil { // first real network connection — confirms DB is reachable
			log.Fatalf("sql: ping failed — is PostgreSQL running?\n  DSN: %s\n  error: %v", cfg.SQLDSN, err)
		}
		// authors table must be created before posts because posts.author_id references it (FK)
		authorRepo, err := sqldb.NewAuthorRepository(db) // runs CREATE TABLE IF NOT EXISTS authors
		if err != nil {
			log.Fatalf("sql: authors migration failed: %v", err)
		}
		postRepo, err := sqldb.NewPostRepository(db) // runs CREATE TABLE IF NOT EXISTS posts
		if err != nil {
			log.Fatalf("sql: posts migration failed: %v", err)
		}
		log.Println("connected to PostgreSQL")
		return repos{post: postRepo, author: authorRepo}

	case "mongo":
		client, err := mongo.Connect(mongoopts.Client().ApplyURI(cfg.MongoURI)) // parses URI; actual connection is lazy
		if err != nil {
			log.Fatalf("mongo: connect failed: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // ping must complete within 10s
		defer cancel()                                                            // releases timer resources when mustInitRepos returns
		if err := client.Ping(ctx, nil); err != nil {                            // forces a real connection to verify the server is up
			log.Fatalf("mongo: ping failed — is MongoDB running?\n  URI: %s\n  error: %v", cfg.MongoURI, err)
		}
		db := client.Database(cfg.MongoDB) // selects the database; created automatically on first write if it doesn't exist
		log.Println("connected to MongoDB")
		return repos{
			post:   mongodb.NewPostRepository(db),
			author: mongodb.NewAuthorRepository(db),
		}

	default: // DB_DRIVER is empty, unset, or unrecognised — safe fallback for local dev without a DB
		log.Printf("unknown DB_DRIVER %q — falling back to in-memory store", cfg.DBDriver)
		return repos{
			post:   memory.NewPostRepository(),                 // plain map; all data is lost when the process exits
			author: memory.NewAuthorRepository(), // plain map; all data is lost when the process exits
		}
	}
}
