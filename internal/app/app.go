package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/barbodimani81/task/internal/greetings"
)

// Greeter looks up greeting messages. greetings.Store implements it.
type Greeter interface {
	Message(ctx context.Context, language string) (string, error)
	Ping(ctx context.Context) error
}

type indexResponse struct {
	Message string `json:"message"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// NewHandler returns the HTTP routes for the service:
// GET / (greeting), GET /healthz (liveness) and GET /readyz (readiness).
func NewHandler(version, env string, greeter Greeter) http.Handler {
	mux := http.NewServeMux()

	// Liveness: the process is up. It must not check the database, or a
	// database outage would make Kubernetes restart every Pod at once.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// A write error only means the client went away; nothing to do.
		_, _ = w.Write([]byte("ok"))
	})

	// Readiness: this Pod can serve requests, which needs the database.
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := greeter.Ping(ctx); err != nil {
			log.Printf("readiness: database ping: %v", err)
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		lang := r.URL.Query().Get("lang")
		if lang == "" {
			lang = "en"
		}

		msg, err := greeter.Message(r.Context(), lang)
		switch {
		case errors.Is(err, greetings.ErrNotFound):
			writeJSON(w, http.StatusNotFound, errorResponse{
				Error: fmt.Sprintf("no greeting for language %q", lang),
			})
			return
		case err != nil:
			// Log the details, but never send database errors to the client.
			log.Printf("looking up greeting %q: %v", lang, err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
			return
		}

		writeJSON(w, http.StatusOK, indexResponse{Message: msg, Version: version, Env: env})
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writing response: %v", err)
	}
}
