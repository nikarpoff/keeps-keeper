package notehandler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"keeper/internal/auth"
	"keeper/internal/notes"
)

func TestHandler(t *testing.T) {
	var (
		service *noteServiceMock
		router  *mux.Router
		resp    *httptest.ResponseRecorder
	)

	setup := func() {
		service = &noteServiceMock{
			SaveNoteFunc: func(_ context.Context, id int64, req *notes.NoteRequest, login string) (*notes.Note, error) {
				return &notes.Note{NoteSummary: notes.NoteSummary{ID: 5, Title: req.Title, UpdatedBy: login}}, nil
			},
			NoteFunc:       func(context.Context, int64) (*notes.Note, error) { return nil, notes.ErrNotFound },
			DeleteNoteFunc: func(context.Context, int64) error { return errors.New("db down") },
		}
		h := NewHandler(service)
		router = mux.NewRouter()
		router.HandleFunc("/notes", h.SaveNote).Methods(http.MethodPost)
		router.HandleFunc("/notes/{id}", h.SaveNote).Methods(http.MethodPut)
		router.HandleFunc("/notes/{id}", h.GetNote).Methods(http.MethodGet)
		router.HandleFunc("/notes/{id}", h.DeleteNote).Methods(http.MethodDelete)
		resp = httptest.NewRecorder()
	}

	do := func(method, path, body string) {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(resp, req.WithContext(auth.WithLogin(req.Context(), "bob")))
	}

	t.Run("must_pass_create_note", func(t *testing.T) {
		setup()
		do(http.MethodPost, "/notes", `{"title":"Hi"}`)
		require.Equal(t, http.StatusCreated, resp.Code)
		require.Contains(t, resp.Body.String(), `"updated_by":"bob"`)
		require.Equal(t, int64(0), service.SaveNoteCalls()[0].ID)
	})

	t.Run("must_pass_update_note", func(t *testing.T) {
		setup()
		do(http.MethodPut, "/notes/5", `{"title":"Hi"}`)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Equal(t, int64(5), service.SaveNoteCalls()[0].ID)
	})

	t.Run("must_fail_invalid_body", func(t *testing.T) {
		setup()
		do(http.MethodPost, "/notes", `{`)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.JSONEq(t, `{"error":"invalid request body"}`, resp.Body.String())
	})

	t.Run("must_fail_not_found", func(t *testing.T) {
		setup()
		do(http.MethodGet, "/notes/9", "")
		require.Equal(t, http.StatusNotFound, resp.Code)
		require.JSONEq(t, `{"error":"not found"}`, resp.Body.String())
	})

	t.Run("must_fail_internal_error_impersonal", func(t *testing.T) {
		setup()
		do(http.MethodDelete, "/notes/9", "")
		require.Equal(t, http.StatusInternalServerError, resp.Code)
		require.JSONEq(t, `{"error":"internal server error"}`, resp.Body.String())
	})
}
