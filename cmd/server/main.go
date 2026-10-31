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

	"github.com/barbodimani81/task/internal/app"
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

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: app.NewHandler(version, env),
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
