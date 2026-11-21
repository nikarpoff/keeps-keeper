// Package httpserver wires routes and runs the HTTP server.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"keeper/internal/env"
	"keeper/internal/logger"
)

const idPath = "/{id:[1-9][0-9]*}"

//go:generate moq -out auth_handler_moq_test.go . authHandler
type authHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Me(w http.ResponseWriter, r *http.Request)
	Middleware(next http.Handler) http.Handler
}

//go:generate moq -out note_handler_moq_test.go . noteHandler
type noteHandler interface {
	ListFolders(w http.ResponseWriter, r *http.Request)
	SaveFolder(w http.ResponseWriter, r *http.Request)
	DeleteFolder(w http.ResponseWriter, r *http.Request)
	ListNotes(w http.ResponseWriter, r *http.Request)
	GetNote(w http.ResponseWriter, r *http.Request)
	SaveNote(w http.ResponseWriter, r *http.Request)
	DeleteNote(w http.ResponseWriter, r *http.Request)
}

// Server is the HTTP transport of the service.
type Server struct {
	vars   *env.Vars
	server *http.Server
}

// NewServer registers the routes and configures the HTTP server.
func NewServer(vars *env.Vars, auth authHandler, notes noteHandler) *Server {
	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	router.HandleFunc("/api/login", auth.Login).Methods(http.MethodPost)
	router.HandleFunc("/api/logout", auth.Logout).Methods(http.MethodPost)

	api := router.PathPrefix("/api").Subrouter()
	api.Use(auth.Middleware)
	api.HandleFunc("/me", auth.Me).Methods(http.MethodGet)

	api.HandleFunc("/folders", notes.ListFolders).Methods(http.MethodGet)
	api.HandleFunc("/folders", notes.SaveFolder).Methods(http.MethodPost)
	api.HandleFunc("/folders"+idPath, notes.SaveFolder).Methods(http.MethodPut)
	api.HandleFunc("/folders"+idPath, notes.DeleteFolder).Methods(http.MethodDelete)

	api.HandleFunc("/notes", notes.ListNotes).Methods(http.MethodGet)
	api.HandleFunc("/notes", notes.SaveNote).Methods(http.MethodPost)
	api.HandleFunc("/notes"+idPath, notes.GetNote).Methods(http.MethodGet)
	api.HandleFunc("/notes"+idPath, notes.SaveNote).Methods(http.MethodPut)
	api.HandleFunc("/notes"+idPath, notes.DeleteNote).Methods(http.MethodDelete)

	router.PathPrefix("/").Handler(http.FileServer(http.Dir(vars.AppStaticDir)))

	return &Server{
		vars: vars,
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", vars.AppServerPort),
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

// Run starts the HTTP server and blocks until it is stopped.
func (s *Server) Run() error {
	logger.Info("Server listening on port %d", s.vars.AppServerPort)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error(err)
		return err
	}
	return nil
}

// Stop gracefully shuts the HTTP server down within APP_SERVER_SHUTDOWN_TIMEOUT.
func (s *Server) Stop() {
	logger.Info("stopping HTTP server...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.vars.AppServerShutdownTimeout)*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error(fmt.Errorf("failed to shutdown HTTP server: %w", err))
	}
}
