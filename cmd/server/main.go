package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/barbodimani81/task/internal/app"
	"github.com/barbodimani81/task/internal/greetings"
)

// getenv returns the environment variable key, or fallback if it is unset or empty.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	port := getenv("PORT", "8080")
	version := getenv("APP_VERSION", "dev")
	env := getenv("APP_ENV", "local")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	// The pool connects lazily, so the server starts even while the database
	// is down; /readyz reports that until it comes back.
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("parsing DATABASE_URL: %v", err)
	}
	defer pool.Close()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: app.NewHandler(version, env, greetings.NewStore(pool)),
		// Without a header timeout, slow clients can hold connections open forever.
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Ctrl+C locally, SIGTERM from Docker and Kubernetes.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("listening on :%s (version=%s env=%s)", port, version, env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down")

	// Finish in-flight requests, but stay well under Kubernetes' 30s grace period.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Printf("stopped")
}
