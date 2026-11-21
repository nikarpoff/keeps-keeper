package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository stores users in Postgres.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// UserByLogin returns the user or nil if it does not exist.
func (r *Repository) UserByLogin(ctx context.Context, login string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx, `SELECT id, login, password_hash FROM users WHERE login = $1`, login).
		Scan(&u.ID, &u.Login, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to select user: %w", err)
	}
	return &u, nil
}

// CreateUser inserts a new user.
func (r *Repository) CreateUser(ctx context.Context, login, hash string) error {
	if _, err := r.pool.Exec(ctx, `INSERT INTO users (login, password_hash) VALUES ($1, $2)`, login, hash); err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

// SetPassword updates the password hash and reports whether the user exists.
func (r *Repository) SetPassword(ctx context.Context, login, hash string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE users SET password_hash = $2 WHERE login = $1`, login, hash)
	if err != nil {
		return false, fmt.Errorf("failed to update password: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// DeleteUser removes the user and reports whether it existed.
func (r *Repository) DeleteUser(ctx context.Context, login string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE login = $1`, login)
	if err != nil {
		return false, fmt.Errorf("failed to delete user: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}
