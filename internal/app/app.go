package app

import (
	"encoding/json"
	"log"
	"net/http"
)

func Hello() string {
	return "Hello, World!"
}

type indexResponse struct {
	Message string `json:"message"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

// NewHandler returns the HTTP routes for the service: GET / and GET /healthz.
func NewHandler(version, env string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// A write error only means the client went away; nothing to do.
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(indexResponse{
			Message: Hello(),
			Version: version,
			Env:     env,
		})
		if err != nil {
			log.Printf("writing index response: %v", err)
		}
	})

	return mux
}
