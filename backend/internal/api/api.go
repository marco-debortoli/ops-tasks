// Package api exposes the store over a JSON HTTP API and serves the built frontend.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"opstasks/internal/store"
)

// Clock decides what "today" is: the app's zone, not the browser's or the server's.
type Clock struct {
	Loc *time.Location
}

func (c Clock) Today() string { return time.Now().In(c.Loc).Format(time.DateOnly) }

type Server struct {
	store     *store.Store
	clock     Clock
	staticDir string
}

func New(s *store.Store, clock Clock, staticDir string) http.Handler {
	srv := &Server{store: s, clock: clock, staticDir: staticDir}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handle(srv.health))
	mux.HandleFunc("GET /api/dashboard", handle(srv.dashboard))
	mux.HandleFunc("GET /api/history", handle(srv.history))

	mux.HandleFunc("GET /api/categories", handle(srv.listCategories))
	mux.HandleFunc("POST /api/categories", handle(srv.createCategory))
	mux.HandleFunc("PATCH /api/categories/{id}", handle(srv.updateCategory))
	mux.HandleFunc("DELETE /api/categories/{id}", handle(srv.deleteCategory))

	mux.HandleFunc("POST /api/tasks", handle(srv.createTask))
	mux.HandleFunc("GET /api/tasks/{id}", handle(srv.getTask))
	mux.HandleFunc("PATCH /api/tasks/{id}", handle(srv.updateTask))
	mux.HandleFunc("DELETE /api/tasks/{id}", handle(srv.deleteTask))
	mux.HandleFunc("POST /api/tasks/{id}/complete", handle(srv.completeTask))
	mux.HandleFunc("POST /api/tasks/{id}/reopen", handle(srv.reopenTask))

	mux.HandleFunc("POST /api/tasks/{id}/subtasks", handle(srv.createSubtask))
	mux.HandleFunc("PUT /api/tasks/{id}/subtasks/order", handle(srv.reorderSubtasks))
	mux.HandleFunc("PATCH /api/subtasks/{id}", handle(srv.updateSubtask))
	mux.HandleFunc("DELETE /api/subtasks/{id}", handle(srv.deleteSubtask))

	mux.HandleFunc("GET /api/projects", handle(srv.listProjects))
	mux.HandleFunc("POST /api/projects", handle(srv.createProject))
	mux.HandleFunc("GET /api/projects/{id}", handle(srv.getProject))
	mux.HandleFunc("PATCH /api/projects/{id}", handle(srv.updateProject))
	mux.HandleFunc("DELETE /api/projects/{id}", handle(srv.deleteProject))

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, errorBody{"no such endpoint"})
	})
	mux.HandleFunc("/", srv.spa)
	return logRequests(mux)
}

type errorBody struct {
	Error string `json:"error"`
}

// handlerFunc returns the response body, or an error mapped to a status code.
// A nil body with no error means 204 No Content.
type handlerFunc func(r *http.Request) (any, error)

type badRequest struct{ msg string }

func (e badRequest) Error() string { return e.msg }

func handle(fn handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := fn(r)
		var ve *store.ValidationError
		var br badRequest
		switch {
		case err == nil && body == nil:
			w.WriteHeader(http.StatusNoContent)
		case err == nil:
			status := http.StatusOK
			if r.Method == http.MethodPost {
				status = http.StatusCreated
			}
			writeJSON(w, status, body)
		case errors.Is(err, store.ErrNotFound):
			writeJSON(w, http.StatusNotFound, errorBody{"not found"})
		case errors.As(err, &ve):
			writeJSON(w, http.StatusBadRequest, errorBody{ve.Msg})
		case errors.As(err, &br):
			writeJSON(w, http.StatusBadRequest, errorBody{br.msg})
		default:
			slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
			writeJSON(w, http.StatusInternalServerError, errorBody{"internal error"})
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func decode(r *http.Request, v any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return badRequest{"request body is empty"}
		}
		return badRequest{fmt.Sprintf("invalid JSON: %v", err)}
	}
	return nil
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, store.ErrNotFound
	}
	return id, nil
}

// spa serves the built frontend, falling back to index.html for client-side routes.
func (s *Server) spa(w http.ResponseWriter, r *http.Request) {
	if s.staticDir == "" {
		http.Error(w, "frontend not built", http.StatusNotFound)
		return
	}
	p := filepath.Join(s.staticDir, filepath.FromSlash(filepath.Clean("/"+r.URL.Path)))
	if info, err := os.Stat(p); err != nil || info.IsDir() {
		p = filepath.Join(s.staticDir, "index.html")
	}
	if strings.HasPrefix(r.URL.Path, "/_app/immutable/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	http.ServeFile(w, r, p)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("api", "method", r.Method, "path", r.URL.Path, "status", rec.status, "dur", time.Since(start).Round(time.Millisecond))
	})
}
