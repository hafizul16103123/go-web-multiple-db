package config

import "os"

// Config holds all runtime configuration read from environment variables.
type Config struct {
	DBDriver   string // "sql" (default) | "mongo"
	SQLDSN     string // postgres://user:pass@host:5432/dbname?sslmode=disable
	MongoURI   string // mongodb://localhost:27017
	MongoDB    string // MongoDB database name
	ServerAddr string // :8080
}

// Load reads environment variables and returns a Config with sensible defaults.
func Load() *Config {
	return &Config{
		DBDriver:   env("DB_DRIVER", "sql"),
		SQLDSN:     env("SQL_DSN", "postgres://postgres:password@localhost:5432/webapp?sslmode=disable"),
		MongoURI:   env("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:    env("MONGO_DB", "webapp"),
		ServerAddr: env("SERVER_ADDR", ":8080"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
