package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
	DataDir     string
	AppOrigins  []string // APP_ORIGINS: extra frontend origins trusted by the API CSRF guard
}

func Load() (*Config, error) {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://golearn:golearn@localhost:5433/golearn?sslmode=disable")
	dataDir := getEnv("DATA_DIR", "./data")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		DataDir:     dataDir,
		AppOrigins:  splitList(os.Getenv("APP_ORIGINS")),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitList(v string) []string {
	var out []string
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
