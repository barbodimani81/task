package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/barbodimani81/task/internal/greetings"
)

// fakeGreeter is an in-memory Greeter, so these tests need no database.
type fakeGreeter struct {
	messages map[string]string
	err      error // returned by every call when set
}

func (f fakeGreeter) Message(ctx context.Context, language string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	msg, ok := f.messages[language]
	if !ok {
		return "", greetings.ErrNotFound
	}
	return msg, nil
}

func (f fakeGreeter) Ping(ctx context.Context) error {
	return f.err
}

var testGreeter = fakeGreeter{messages: map[string]string{
	"en": "Hello, World!",
	"fa": "سلام، دنیا!",
}}

var dbDown = fakeGreeter{err: errors.New(`relation "greetings" does not exist`)}

// serve sends one request to the handler in memory, with no network or port.
func serve(g Greeter, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	NewHandler("v1", "test", g).ServeHTTP(rec, req)
	return rec
}

func TestHealthzIgnoresDatabase(t *testing.T) {
	rec := serve(dbDown, http.MethodGet, "/healthz")

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Errorf("body = %q, want %q", got, "ok")
	}
}

func TestReadyz(t *testing.T) {
	if rec := serve(testGreeter, http.MethodGet, "/readyz"); rec.Code != http.StatusOK {
		t.Errorf("database up: status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec := serve(dbDown, http.MethodGet, "/readyz"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("database down: status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestIndex(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/", "Hello, World!"},
		{"/?lang=en", "Hello, World!"},
		{"/?lang=fa", "سلام، دنیا!"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := serve(testGreeter, http.MethodGet, tt.path)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/json")
			}

			var got indexResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decoding body: %v", err)
			}
			want := indexResponse{Message: tt.want, Version: "v1", Env: "test"}
			if got != want {
				t.Errorf("body = %+v, want %+v", got, want)
			}
		})
	}
}

func TestIndexUnknownLanguageIs404(t *testing.T) {
	rec := serve(testGreeter, http.MethodGet, "/?lang=xx")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestIndexDatabaseErrorHidesDetails(t *testing.T) {
	rec := serve(dbDown, http.MethodGet, "/")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var got errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if got.Error != "internal error" {
		t.Errorf("error = %q, want %q", got.Error, "internal error")
	}
}

func TestUnknownPathIs404(t *testing.T) {
	rec := serve(testGreeter, http.MethodGet, "/nope")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestPostHealthzIs405(t *testing.T) {
	rec := serve(testGreeter, http.MethodPost, "/healthz")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
