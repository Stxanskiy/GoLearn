package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/backendraz/golearn/internal/config"
	"github.com/backendraz/golearn/internal/handler"
	"github.com/backendraz/golearn/internal/migrate"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	_ = godotenv.Load() // load .env if exists

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		log.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Schema is brought up to date on every boot: a fresh volume or a stale one
	// missing late columns no longer needs a manual psql step.
	migCtx, migCancel := context.WithTimeout(context.Background(), 60*time.Second)
	ran, err := migrate.Up(migCtx, pool, os.Getenv("MIGRATIONS_DIR"))
	migCancel()
	if err != nil {
		log.Error("apply migrations", "error", err)
		os.Exit(1)
	}
	if len(ran) > 0 {
		log.Info("migrations applied", "count", len(ran), "versions", ran)
	}

	moduleRepo := repository.NewModuleRepo(pool)
	lessonRepo := repository.NewLessonRepo(pool)
	progressRepo := repository.NewProgressRepo(pool)
	submissionRepo := repository.NewSubmissionRepo(pool)
	userRepo := repository.NewUserRepo(pool)
	specRepo := repository.NewSpecRepo(pool)
	courseRepo := repository.NewCourseRepo(pool, moduleRepo, lessonRepo)
	simRepo := repository.NewSimRepo(pool)

	// A freshly migrated database has no accounts and self-registration is off
	// by default, so seed the first admin instead of locking the owner out.
	bootCtx, bootCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if email, err := userRepo.BootstrapAdmin(bootCtx); err != nil {
		log.Error("bootstrap admin", "error", err)
	} else if email != "" {
		log.Info("created first admin account", "email", email,
			"password", "ADMIN_PASSWORD env (default golearn123) — смени после входа")
	}
	bootCancel()

	h := handler.New(moduleRepo, lessonRepo, progressRepo, submissionRepo, userRepo, specRepo, courseRepo, simRepo, log)

	// Seed the built-in simulator scenarios into the DB once so they are editable.
	simCtx, simCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := h.EnsureSimulators(simCtx); err != nil {
		log.Error("ensure simulators", "error", err)
	}
	simCancel()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// Serve static files. Assets are linked with a ?v=<version> query that
	// changes when they change, so the browser refetches after a redesign;
	// must-revalidate keeps a cached copy honest even without the query.
	fs := http.FileServer(http.Dir("internal/static"))
	staticHandler := http.StripPrefix("/static/", fs)
	r.Handle("/static/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=300, must-revalidate")
		staticHandler.ServeHTTP(w, req)
	}))

	h.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
		// No Read/WriteTimeout: those deadlines persist onto hijacked WebSocket
		// connections (the interactive terminal) and would drop them after ~15s.
		// ReadHeaderTimeout still guards against slow-header (slowloris) attacks.
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("server starting", "port", cfg.Port, "url", "http://localhost:"+cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "error", err)
	}
	log.Info("server stopped")
}
