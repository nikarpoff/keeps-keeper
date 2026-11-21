package authhandler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"keeper/internal/auth"
)

func TestHandler(t *testing.T) {
	var (
		service *authServiceMock
		handler *Handler
		resp    *httptest.ResponseRecorder
		next    http.Handler
	)

	setup := func() {
		service = &authServiceMock{
			LoginFunc: func(_ context.Context, req *auth.LoginRequest) (string, error) {
				if req.Password != "secret" {
					return "", auth.ErrInvalidCredentials
				}
				return "token", nil
			},
			VerifyFunc: func(_ context.Context, token string) (string, error) {
				switch token {
				case "token":
					return "bob", nil
				case "broken":
					return "", errors.New("db down")
				}
				return "", auth.ErrInvalidToken
			},
		}
		handler = NewHandler(service, time.Hour, true)
		resp = httptest.NewRecorder()
		next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(auth.LoginFrom(r.Context())))
		})
	}

	login := func(body string) {
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		handler.Login(resp, req)
	}

	withToken := func(token string) {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		if token != "" {
			req.AddCookie(&http.Cookie{Name: cookieName, Value: token})
		}
		handler.Middleware(next).ServeHTTP(resp, req)
	}

	t.Run("must_pass_login_sets_cookie", func(t *testing.T) {
		setup()
		login(`{"login":"bob","password":"secret"}`)
		require.Equal(t, http.StatusOK, resp.Code)
		cookie := resp.Result().Cookies()[0]
		require.Equal(t, "token", cookie.Value)
		require.True(t, cookie.HttpOnly)
		require.True(t, cookie.Secure)
	})

	t.Run("must_fail_login_wrong_password", func(t *testing.T) {
		setup()
		login(`{"login":"bob","password":"x"}`)
		require.Equal(t, http.StatusUnauthorized, resp.Code)
		require.JSONEq(t, `{"error":"invalid login or password"}`, resp.Body.String())
	})

	t.Run("must_fail_login_content_type", func(t *testing.T) {
		setup()
		handler.Login(resp, httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{}`)))
		require.Equal(t, http.StatusUnsupportedMediaType, resp.Code)
	})

	t.Run("must_pass_middleware", func(t *testing.T) {
		setup()
		withToken("token")
		require.Equal(t, http.StatusOK, resp.Code)
		require.Equal(t, "bob", resp.Body.String())
	})

	t.Run("must_fail_middleware_without_cookie", func(t *testing.T) {
		setup()
		withToken("")
		require.Equal(t, http.StatusUnauthorized, resp.Code)
		require.Empty(t, service.VerifyCalls())
	})

	t.Run("must_fail_middleware_invalid_token", func(t *testing.T) {
		setup()
		withToken("forged")
		require.Equal(t, http.StatusUnauthorized, resp.Code)
		require.JSONEq(t, `{"error":"invalid token"}`, resp.Body.String())
	})

	t.Run("must_fail_middleware_internal_error", func(t *testing.T) {
		setup()
		withToken("broken")
		require.Equal(t, http.StatusInternalServerError, resp.Code)
		require.JSONEq(t, `{"error":"internal server error"}`, resp.Body.String())
	})
}
