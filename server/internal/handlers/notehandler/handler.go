// Package notehandler serves the notes and folders HTTP API.
package notehandler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"keeper/internal/auth"
	"keeper/internal/handlers"
	"keeper/internal/notes"
)

//go:generate moq -out note_service_moq_test.go . noteService
type noteService interface {
	Folders(ctx context.Context) ([]notes.Folder, error)
	SaveFolder(ctx context.Context, id int64, req *notes.FolderRequest) (*notes.Folder, error)
	DeleteFolder(ctx context.Context, id int64) error
	Notes(ctx context.Context) ([]notes.NoteSummary, error)
	Note(ctx context.Context, id int64) (*notes.Note, error)
	SaveNote(ctx context.Context, id int64, req *notes.NoteRequest, login string) (*notes.Note, error)
	DeleteNote(ctx context.Context, id int64) error
}

// Handler serves the notes HTTP API.
type Handler struct {
	service noteService
}

// NewHandler creates a new Handler.
func NewHandler(service noteService) *Handler {
	return &Handler{service: service}
}

// ListFolders handles GET /api/folders
func (h *Handler) ListFolders(w http.ResponseWriter, r *http.Request) {
	folders, err := h.service.Folders(r.Context())
	respond(w, http.StatusOK, folders, err)
}

// SaveFolder handles POST /api/folders and PUT /api/folders/{id}
func (h *Handler) SaveFolder(w http.ResponseWriter, r *http.Request) {
	var req notes.FolderRequest
	if !handlers.DecodeJSON(w, r, &req) {
		return
	}
	id := pathID(r)
	folder, err := h.service.SaveFolder(r.Context(), id, &req)
	respond(w, savedCode(id), folder, err)
}

// DeleteFolder handles DELETE /api/folders/{id}
func (h *Handler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusNoContent, nil, h.service.DeleteFolder(r.Context(), pathID(r)))
}

// ListNotes handles GET /api/notes
func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.Notes(r.Context())
	respond(w, http.StatusOK, list, err)
}

// GetNote handles GET /api/notes/{id}
func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) {
	note, err := h.service.Note(r.Context(), pathID(r))
	respond(w, http.StatusOK, note, err)
}

// SaveNote handles POST /api/notes and PUT /api/notes/{id}
func (h *Handler) SaveNote(w http.ResponseWriter, r *http.Request) {
	var req notes.NoteRequest
	if !handlers.DecodeJSON(w, r, &req) {
		return
	}
	id := pathID(r)
	note, err := h.service.SaveNote(r.Context(), id, &req, auth.LoginFrom(r.Context()))
	respond(w, savedCode(id), note, err)
}

// DeleteNote handles DELETE /api/notes/{id}
func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusNoContent, nil, h.service.DeleteNote(r.Context(), pathID(r)))
}

func respond(w http.ResponseWriter, code int, payload any, err error) {
	switch {
	case errors.Is(err, notes.ErrNotFound):
		handlers.RespondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, notes.ErrEmptyName), errors.Is(err, notes.ErrEmptyTitle):
		handlers.RespondError(w, http.StatusUnprocessableEntity, err.Error())
	case err != nil:
		handlers.RespondInternal(w, err)
	case code == http.StatusNoContent:
		w.WriteHeader(code)
	default:
		handlers.RespondJSON(w, code, payload)
	}
}

func pathID(r *http.Request) int64 {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	return id
}

func savedCode(id int64) int {
	if id == 0 {
		return http.StatusCreated
	}
	return http.StatusOK
}
