// Package authhandler serves login endpoints and the auth middleware.
package authhandler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"keeper/internal/auth"
	"keeper/internal/handlers"
)

const cookieName = "keeper_token"

//go:generate moq -out auth_service_moq_test.go . authService
type authService interface {
	Login(ctx context.Context, req *auth.LoginRequest) (string, error)
	Verify(ctx context.Context, token string) (string, error)
}

// Handler serves the auth HTTP API.
type Handler struct {
	service authService
	ttl     time.Duration
	secure  bool
}

// NewHandler creates a new Handler.
func NewHandler(service authService, ttl time.Duration, secure bool) *Handler {
	return &Handler{service: service, ttl: ttl, secure: secure}
}

// Login handles POST /api/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if !handlers.DecodeJSON(w, r, &req) {
		return
	}
	token, err := h.service.Login(r.Context(), &req)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		handlers.RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if err != nil {
		handlers.RespondInternal(w, err)
		return
	}
	h.setCookie(w, token, h.ttl)
	handlers.RespondJSON(w, http.StatusOK, &auth.UserResponse{Login: req.Login})
}

// Logout handles POST /api/logout
func (h *Handler) Logout(w http.ResponseWriter, _ *http.Request) {
	h.setCookie(w, "", -time.Second)
	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /api/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	handlers.RespondJSON(w, http.StatusOK, &auth.UserResponse{Login: auth.LoginFrom(r.Context())})
}

// Middleware rejects requests without a valid token cookie.
func (h *Handler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			handlers.RespondError(w, http.StatusUnauthorized, auth.ErrInvalidToken.Error())
			return
		}
		login, err := h.service.Verify(r.Context(), cookie.Value)
		if errors.Is(err, auth.ErrInvalidToken) {
			handlers.RespondError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if err != nil {
			handlers.RespondInternal(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.WithLogin(r.Context(), login)))
	})
}

func (h *Handler) setCookie(w http.ResponseWriter, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})
}
