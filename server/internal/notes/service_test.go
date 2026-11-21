package notes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	var (
		ctx     = context.Background()
		repo    *noteRepositoryMock
		service *Service
	)

	setup := func() {
		repo = &noteRepositoryMock{
			SaveFolderFunc: func(_ context.Context, id int64, name string) (*Folder, error) {
				if id > 10 {
					return nil, nil
				}
				return &Folder{ID: 1, Name: name}, nil
			},
			DeleteFolderFunc: func(_ context.Context, id int64) (bool, error) { return id == 1, nil },
			NoteFunc:         func(_ context.Context, id int64) (*Note, error) { return nil, nil },
			SaveNoteFunc: func(_ context.Context, id int64, req *NoteRequest, login string) (*Note, error) {
				return &Note{NoteSummary: NoteSummary{ID: 1, Title: req.Title, UpdatedBy: login}}, nil
			},
			DeleteNoteFunc: func(context.Context, int64) (bool, error) { return false, nil },
		}
		service = NewService(repo)
	}

	t.Run("must_pass_save_folder_trimmed", func(t *testing.T) {
		setup()
		folder, err := service.SaveFolder(ctx, 0, &FolderRequest{Name: "  work "})
		require.NoError(t, err)
		require.Equal(t, "work", folder.Name)
	})

	t.Run("must_fail_save_folder_empty_name", func(t *testing.T) {
		setup()
		_, err := service.SaveFolder(ctx, 0, &FolderRequest{Name: "   "})
		require.EqualError(t, err, ErrEmptyName.Error())
		require.Empty(t, repo.SaveFolderCalls())
	})

	t.Run("must_fail_rename_missing_folder", func(t *testing.T) {
		setup()
		_, err := service.SaveFolder(ctx, 42, &FolderRequest{Name: "x"})
		require.EqualError(t, err, ErrNotFound.Error())
	})

	t.Run("must_fail_delete_missing_folder", func(t *testing.T) {
		setup()
		require.NoError(t, service.DeleteFolder(ctx, 1))
		require.EqualError(t, service.DeleteFolder(ctx, 2), ErrNotFound.Error())
	})

	t.Run("must_pass_save_note", func(t *testing.T) {
		setup()
		note, err := service.SaveNote(ctx, 0, &NoteRequest{Title: " Hi ", Body: "# Hi"}, "bob")
		require.NoError(t, err)
		require.Equal(t, "Hi", note.Title)
		require.Equal(t, "bob", note.UpdatedBy)
	})

	t.Run("must_fail_save_note_empty_title", func(t *testing.T) {
		setup()
		_, err := service.SaveNote(ctx, 0, &NoteRequest{Title: " "}, "bob")
		require.EqualError(t, err, ErrEmptyTitle.Error())
		require.Empty(t, repo.SaveNoteCalls())
	})

	t.Run("must_fail_missing_note", func(t *testing.T) {
		setup()
		_, err := service.Note(ctx, 7)
		require.EqualError(t, err, ErrNotFound.Error())
		require.EqualError(t, service.DeleteNote(ctx, 7), ErrNotFound.Error())
	})
}
