package notes

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const noteColumns = `id, title, folder_id, updated_by, updated_at, body`

// Repository stores notes and folders in Postgres.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Folders returns all folders ordered by name.
func (r *Repository) Folders(ctx context.Context) ([]Folder, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM folders ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to select folders: %w", err)
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Folder])
}

// SaveFolder inserts a folder when id is 0, otherwise renames it. Returns nil if missing.
func (r *Repository) SaveFolder(ctx context.Context, id int64, name string) (*Folder, error) {
	query, args := `INSERT INTO folders (name) VALUES ($1) RETURNING id, name`, []any{name}
	if id != 0 {
		query, args = `UPDATE folders SET name = $1 WHERE id = $2 RETURNING id, name`, append(args, id)
	}
	var f Folder
	err := r.pool.QueryRow(ctx, query, args...).Scan(&f.ID, &f.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to save folder: %w", err)
	}
	return &f, nil
}

// DeleteFolder removes a folder and reports whether it existed.
func (r *Repository) DeleteFolder(ctx context.Context, id int64) (bool, error) {
	return r.delete(ctx, `DELETE FROM folders WHERE id = $1`, id)
}

// Notes returns all notes without bodies, most recently updated first.
func (r *Repository) Notes(ctx context.Context) ([]NoteSummary, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, title, folder_id, updated_by, updated_at FROM notes ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to select notes: %w", err)
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[NoteSummary])
}

// Note returns the note or nil if it does not exist.
func (r *Repository) Note(ctx context.Context, id int64) (*Note, error) {
	return r.scanNote(r.pool.QueryRow(ctx, `SELECT `+noteColumns+` FROM notes WHERE id = $1`, id))
}

// SaveNote inserts a note when id is 0, otherwise updates it. Returns nil if missing.
func (r *Repository) SaveNote(ctx context.Context, id int64, req *NoteRequest, login string) (*Note, error) {
	query := `INSERT INTO notes (title, body, folder_id, updated_by) VALUES ($1, $2, $3, $4) RETURNING ` + noteColumns
	args := []any{req.Title, req.Body, req.FolderID, login}
	if id != 0 {
		query = `UPDATE notes SET title = $1, body = $2, folder_id = $3, updated_by = $4, updated_at = now()
			WHERE id = $5 RETURNING ` + noteColumns
		args = append(args, id)
	}
	return r.scanNote(r.pool.QueryRow(ctx, query, args...))
}

// DeleteNote removes a note and reports whether it existed.
func (r *Repository) DeleteNote(ctx context.Context, id int64) (bool, error) {
	return r.delete(ctx, `DELETE FROM notes WHERE id = $1`, id)
}

func (r *Repository) scanNote(row pgx.Row) (*Note, error) {
	var n Note
	err := row.Scan(&n.ID, &n.Title, &n.FolderID, &n.UpdatedBy, &n.UpdatedAt, &n.Body)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan note: %w", err)
	}
	return &n, nil
}

func (r *Repository) delete(ctx context.Context, query string, id int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to delete: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}
