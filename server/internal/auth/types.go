// Package auth manages users and session tokens.
package auth

import (
	"context"
	"errors"
)

// User is a registered account.
type User struct {
	ID           int64
	Login        string
	PasswordHash string
}

// LoginRequest is the body of POST /api/login.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// UserResponse describes the current user.
type UserResponse struct {
	Login string `json:"login"`
}

var (
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmptyCredentials   = errors.New("login and password must not be empty")
)

type ctxKey struct{}

// WithLogin stores the authenticated login in the context.
func WithLogin(ctx context.Context, login string) context.Context {
	return context.WithValue(ctx, ctxKey{}, login)
}

// LoginFrom returns the authenticated login from the context.
func LoginFrom(ctx context.Context) string {
	login, _ := ctx.Value(ctxKey{}).(string)
	return login
}
