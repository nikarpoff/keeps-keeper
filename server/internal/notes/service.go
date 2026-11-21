package notes

import (
	"context"
	"strings"
)

//go:generate moq -out note_repository_moq_test.go . noteRepository
type noteRepository interface {
	Folders(ctx context.Context) ([]Folder, error)
	SaveFolder(ctx context.Context, id int64, name string) (*Folder, error)
	DeleteFolder(ctx context.Context, id int64) (bool, error)
	Notes(ctx context.Context) ([]NoteSummary, error)
	Note(ctx context.Context, id int64) (*Note, error)
	SaveNote(ctx context.Context, id int64, req *NoteRequest, login string) (*Note, error)
	DeleteNote(ctx context.Context, id int64) (bool, error)
}

// Service validates requests and maps missing records to ErrNotFound.
type Service struct {
	repo noteRepository
}

// NewService creates a new Service.
func NewService(repo noteRepository) *Service {
	return &Service{repo: repo}
}

// Folders returns all folders.
func (s *Service) Folders(ctx context.Context) ([]Folder, error) {
	return s.repo.Folders(ctx)
}

// SaveFolder creates (id 0) or renames a folder.
func (s *Service) SaveFolder(ctx context.Context, id int64, req *FolderRequest) (*Folder, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrEmptyName
	}
	return found(s.repo.SaveFolder(ctx, id, name))
}

// DeleteFolder removes a folder, its notes stay without a folder.
func (s *Service) DeleteFolder(ctx context.Context, id int64) error {
	return deleted(s.repo.DeleteFolder(ctx, id))
}

// Notes returns all notes without bodies.
func (s *Service) Notes(ctx context.Context) ([]NoteSummary, error) {
	return s.repo.Notes(ctx)
}

// Note returns a single note.
func (s *Service) Note(ctx context.Context, id int64) (*Note, error) {
	return found(s.repo.Note(ctx, id))
}

// SaveNote creates (id 0) or updates a note on behalf of login.
func (s *Service) SaveNote(ctx context.Context, id int64, req *NoteRequest, login string) (*Note, error) {
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return nil, ErrEmptyTitle
	}
	return found(s.repo.SaveNote(ctx, id, req, login))
}

// DeleteNote removes a note.
func (s *Service) DeleteNote(ctx context.Context, id int64) error {
	return deleted(s.repo.DeleteNote(ctx, id))
}

func found[T any](v *T, err error) (*T, error) {
	if err == nil && v == nil {
		return nil, ErrNotFound
	}
	return v, err
}

func deleted(ok bool, err error) error {
	if err == nil && !ok {
		return ErrNotFound
	}
	return err
}
