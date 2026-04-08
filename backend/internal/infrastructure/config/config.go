package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := envOr("DB_HOST", "localhost")
		portDB := envOr("DB_PORT", "5432")
		user := envOr("DB_USER", "postgres")
		pass := envOr("DB_PASSWORD", "postgres")
		name := envOr("DB_NAME", "smart_inventory")
		sslmode := envOr("DB_SSLMODE", "disable")
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, portDB, name, sslmode)
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
	}
}

func envOr(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
