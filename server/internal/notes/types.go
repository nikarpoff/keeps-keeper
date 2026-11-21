// Package notes manages markdown notes and folders.
package notes

import (
	"errors"
	"time"
)

// Folder groups notes.
type Folder struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// FolderRequest is the body for creating or renaming a folder.
type FolderRequest struct {
	Name string `json:"name"`
}

// NoteSummary is a note without its body.
type NoteSummary struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	FolderID  *int64    `json:"folder_id"`
	UpdatedBy string    `json:"updated_by"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Note is a full markdown note.
type Note struct {
	NoteSummary
	Body string `json:"body"`
}

// NoteRequest is the body for creating or updating a note.
type NoteRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	FolderID *int64 `json:"folder_id"`
}

var (
	ErrNotFound   = errors.New("not found")
	ErrEmptyName  = errors.New("name must not be empty")
	ErrEmptyTitle = errors.New("title must not be empty")
)
