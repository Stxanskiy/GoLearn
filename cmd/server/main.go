package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/backendraz/golearn/internal/api"
	"github.com/backendraz/golearn/internal/apidocs"
	"github.com/backendraz/golearn/internal/billing"
	"github.com/backendraz/golearn/internal/config"
	"github.com/backendraz/golearn/internal/migrate"
	"github.com/backendraz/golearn/internal/obs"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/backendraz/golearn/internal/runner"
	"github.com/backendraz/golearn/internal/simulators"
	"github.com/backendraz/golearn/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	// Packages that have no logger of their own (the VM runner records why a boot
	// failed) write through the default one; without this they would go to stderr
	// in a different format.
	slog.SetDefault(log)

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
	quizAttemptRepo := repository.NewQuizAttemptRepo(pool)
	// Two lab backends: Firecracker micro-VMs when an FC host is configured, plain
	// containers otherwise (SANDBOX_LOCAL runs them on this machine's Docker).
	vmRunner := runner.NewVMRunner()
	sandbox := runner.NewDispatcher(runner.NewShellRunner(), vmRunner)

	// The numbers an incident actually starts from: how loaded the database pool
	// is, and how many micro-VMs exist. Both were only answerable by hand before.
	obs.Gauge("golearn_db_pool_acquired", "Connections currently checked out.",
		func() float64 { return float64(pool.Stat().AcquiredConns()) })
	obs.Gauge("golearn_db_pool_total", "Connections in the pool.",
		func() float64 { return float64(pool.Stat().TotalConns()) })
	obs.Gauge("golearn_vm_sessions", "Micro-VMs handed out to a student.",
		func() float64 { return float64(vmRunner.Stats().Sessions) })
	obs.Gauge("golearn_vm_free_slots", "Micro-VM slots still available.",
		func() float64 { return float64(vmRunner.Stats().FreeSlot) })
	// Slots are claimed before a VM boots, so used > sessions+warm means something
	// is starting — or, if it stays that way, that a slot leaked.
	obs.Gauge("golearn_vm_slots_used", "Micro-VM slots taken, including VMs still booting.",
		func() float64 { return float64(vmRunner.Stats().SlotsUsed) })
	obs.Gauge("golearn_vm_warm", "Pre-booted idle micro-VMs across all profiles.",
		func() float64 {
			var n int
			for _, c := range vmRunner.Stats().Warm {
				n += c
			}
			return float64(n)
		})

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

	// Seed the built-in simulator scenarios into the DB once so they are editable.
	simCtx, simCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := simulators.Ensure(simCtx, simRepo); err != nil {
		log.Error("ensure simulators", "error", err)
	}
	simCancel()

	r := chi.NewRouter()
	// RequestID first: everything below it, including the panic handler, can then
	// name the request a report is about.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(obs.Middleware)
	r.Use(obs.Logger(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(securityHeaders)

	// Metrics sit outside /api/v1: they are for operators, not for the API
	// contract, and must stay reachable even when the API is unhappy.
	r.Handle("/metrics", obs.Handler())

	// Liveness/readiness for k8s probes. /readyz pings the DB.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "db unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ready"))
	})

	billingRepo := repository.NewBillingRepo(pool)
	// Access is decided by the expiry, so a stale 'active' row grants nothing;
	// this only keeps the stored status honest for anything reading the table
	// directly, and frees the active-subscription index for a returning customer.
	go sweepSubscriptions(billingRepo, log)

	stores := api.Stores{
		Users:        userRepo,
		Modules:      moduleRepo,
		Lessons:      lessonRepo,
		Progress:     progressRepo,
		Submissions:  submissionRepo,
		Specs:        specRepo,
		Sims:         simRepo,
		QuizAnswers:  repository.NewQuizAnswerRepo(pool),
		QuizAttempts: quizAttemptRepo,
		Authors:      repository.NewCourseAuthorRepo(pool),
		CourseIO:     courseRepo,
		Reviews:      repository.NewReviewRepo(pool),
		Drafts:       repository.NewDraftRepo(pool),
		Sandbox:      sandbox,
		Billing:      billingRepo,
		SeedDetach:   repository.NewSeedDetachRepo(pool),
		Robokassa:    billing.RobokassaFromEnv(),
	}

	// Say which provider is live at startup. Checkout silently falling back to
	// the stub is the kind of thing that is only noticed when a student says
	// they paid and got nothing.
	switch rk := stores.Robokassa; {
	case rk == nil:
		log.Warn("billing: robokassa not configured, checkout uses the stub provider")
	case rk.Test:
		log.Info("billing: robokassa in TEST mode", "login", rk.Login)
	default:
		log.Warn("billing: robokassa in LIVE mode, real cards will be charged", "login", rk.Login)
	}

	// Without object storage uploads stay inline data URIs, so the server still runs.
	switch store, err := storage.New(storage.LoadConfig()); {
	case err == nil:
		stores.Images = store
		// The bucket is prepared again on the first upload, so a warm-up failure is not fatal.
		if err := store.Warm(context.Background()); err != nil {
			log.Warn("object storage not ready yet", "error", err)
		}
	case errors.Is(err, storage.ErrDisabled):
		log.Warn("object storage disabled: set S3_ENDPOINT to enable image uploads")
	default:
		log.Error("object storage unavailable", "error", err)
	}

	apiV1 := api.New(stores, api.Config{AllowedOrigins: cfg.AppOrigins, AppURL: cfg.AppURL, MediaURL: cfg.MediaURL}, log)
	r.Mount("/api/v1", apiV1.Routes())

	// The contract is the only description of this service now that it renders
	// no pages, so it ships with the code it describes.
	docs := apidocs.New(os.Getenv("OPENAPI_SPEC"))
	r.Get("/openapi.yaml", docs.Spec)
	r.Get("/docs", docs.UI)

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

// securityHeaders sets baseline hardening headers on every response. The app relies
// on inline scripts/styles + a couple of CDNs (Google Fonts, unpkg icons) and the
// WebSocket terminal, so the CSP allows those explicitly rather than being maximally
// strict. HSTS is only advertised when the request arrived over HTTPS.
// sweepSubscriptions marks lapsed subscriptions, hourly and on boot.
//
// Hourly rather than at every request: nothing reads the status to decide
// access, so the only cost of being an hour late is a report being an hour
// stale.
func sweepSubscriptions(repo *repository.BillingRepo, log *slog.Logger) {
	run := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		n, err := repo.SweepExpired(ctx)
		if err != nil {
			log.Error("sweep expired subscriptions", "error", err)
			return
		}
		if n > 0 {
			log.Info("subscriptions expired", "count", n)
		}
	}
	run()
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for range t.C {
		run()
	}
}

func securityHeaders(next http.Handler) http.Handler {
	// A JSON API needs to load nothing at all. The permissive policy this
	// replaces existed for the server-rendered pages, which are gone; /docs
	// sets its own, narrower exception for the viewer it loads, and the lab
	// preview sets one too — it is the single response meant to be framed, and
	// frame-ancestors here silently made it unviewable. Tightening this policy
	// means checking those two handlers still override what they need.
	const csp = "default-src 'none'; frame-ancestors 'none'; base-uri 'none'"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", csp)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}
