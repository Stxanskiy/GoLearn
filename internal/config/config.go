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
	// AppURL is where the site lives, used to send a student back from a payment
	// provider. It falls back to the first trusted origin, so a deployment that
	// already lists the frontend needs no second setting.
	AppURL string
}

func Load() (*Config, error) {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://golearn:golearn@localhost:5433/golearn?sslmode=disable")
	dataDir := getEnv("DATA_DIR", "./data")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	origins := splitList(os.Getenv("APP_ORIGINS"))
	appURL := strings.TrimRight(os.Getenv("APP_URL"), "/")
	if appURL == "" && len(origins) > 0 {
		appURL = strings.TrimRight(origins[0], "/")
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		DataDir:     dataDir,
		AppOrigins:  origins,
		AppURL:      appURL,
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
