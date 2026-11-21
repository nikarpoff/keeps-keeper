package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"keeper/internal/env"
)

func TestNewServer(t *testing.T) {
	var (
		server *Server
		auth   *authHandlerMock
		notes  *noteHandlerMock
		resp   *httptest.ResponseRecorder
		ok     = func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	)

	setup := func() {
		auth = &authHandlerMock{
			LoginFunc: ok,
			MiddlewareFunc: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("X-Test-Auth") == "" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		}
		notes = &noteHandlerMock{GetNoteFunc: ok}
		server = NewServer(&env.Vars{AppServerPort: 80, AppStaticDir: t.TempDir()}, auth, notes)
		resp = httptest.NewRecorder()
	}

	serve := func(method, path string, authed bool) {
		req := httptest.NewRequest(method, path, nil)
		if authed {
			req.Header.Set("X-Test-Auth", "1")
		}
		server.server.Handler.ServeHTTP(resp, req)
	}

	t.Run("must_pass_health", func(t *testing.T) {
		setup()
		serve(http.MethodGet, "/health", false)
		require.Equal(t, "OK", resp.Body.String())
	})

	t.Run("must_pass_login_without_auth", func(t *testing.T) {
		setup()
		serve(http.MethodPost, "/api/login", false)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Len(t, auth.LoginCalls(), 1)
	})

	t.Run("must_pass_get_note", func(t *testing.T) {
		setup()
		serve(http.MethodGet, "/api/notes/12", true)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Len(t, notes.GetNoteCalls(), 1)
	})

	t.Run("must_fail_api_without_auth", func(t *testing.T) {
		setup()
		serve(http.MethodGet, "/api/notes/12", false)
		require.Equal(t, http.StatusUnauthorized, resp.Code)
		require.Empty(t, notes.GetNoteCalls())
	})

	t.Run("must_fail_zero_id", func(t *testing.T) {
		setup()
		serve(http.MethodGet, "/api/notes/0", true)
		require.Equal(t, http.StatusNotFound, resp.Code)
		require.Empty(t, notes.GetNoteCalls())
	})
}
