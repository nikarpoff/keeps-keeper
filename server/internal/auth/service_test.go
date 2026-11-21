package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestService(t *testing.T) {
	var (
		ctx     = context.Background()
		repo    *userRepositoryMock
		service *Service
		users   map[string]*User
		hash, _ = bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	)

	setup := func() {
		users = map[string]*User{"bob": {ID: 1, Login: "bob", PasswordHash: string(hash)}}
		repo = &userRepositoryMock{
			UserByLoginFunc: func(_ context.Context, login string) (*User, error) { return users[login], nil },
			CreateUserFunc:  func(context.Context, string, string) error { return nil },
			SetPasswordFunc: func(_ context.Context, login, _ string) (bool, error) { return users[login] != nil, nil },
			DeleteUserFunc:  func(_ context.Context, login string) (bool, error) { return users[login] != nil, nil },
		}
		service = NewService(repo, "jwt-secret", time.Hour)
	}

	t.Run("must_pass_login_and_verify", func(t *testing.T) {
		setup()
		token, err := service.Login(ctx, &LoginRequest{Login: "bob", Password: "secret"})
		require.NoError(t, err)

		login, err := service.Verify(ctx, token)
		require.NoError(t, err)
		require.Equal(t, "bob", login)
	})

	t.Run("must_fail_login_wrong_password", func(t *testing.T) {
		setup()
		_, err := service.Login(ctx, &LoginRequest{Login: "bob", Password: "wrong"})
		require.EqualError(t, err, ErrInvalidCredentials.Error())
	})

	t.Run("must_fail_login_unknown_user", func(t *testing.T) {
		setup()
		_, err := service.Login(ctx, &LoginRequest{Login: "alice", Password: "secret"})
		require.EqualError(t, err, ErrInvalidCredentials.Error())
	})

	t.Run("must_fail_login_repository_error", func(t *testing.T) {
		setup()
		repo.UserByLoginFunc = func(context.Context, string) (*User, error) { return nil, errors.New("db down") }
		_, err := service.Login(ctx, &LoginRequest{Login: "bob", Password: "secret"})
		require.EqualError(t, err, "db down")
	})

	t.Run("must_fail_verify_bad_token", func(t *testing.T) {
		setup()
		_, err := service.Verify(ctx, "garbage")
		require.EqualError(t, err, ErrInvalidToken.Error())
	})

	t.Run("must_fail_verify_foreign_secret", func(t *testing.T) {
		setup()
		token, err := NewService(repo, "other", time.Hour).Login(ctx, &LoginRequest{Login: "bob", Password: "secret"})
		require.NoError(t, err)
		_, err = service.Verify(ctx, token)
		require.EqualError(t, err, ErrInvalidToken.Error())
	})

	t.Run("must_fail_verify_deleted_user", func(t *testing.T) {
		setup()
		token, err := service.Login(ctx, &LoginRequest{Login: "bob", Password: "secret"})
		require.NoError(t, err)
		delete(users, "bob")
		_, err = service.Verify(ctx, token)
		require.EqualError(t, err, ErrInvalidToken.Error())
	})

	t.Run("must_pass_add_user", func(t *testing.T) {
		setup()
		require.NoError(t, service.AddUser(ctx, "alice", "pass"))
		calls := repo.CreateUserCalls()
		require.Len(t, calls, 1)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(calls[0].Hash), []byte("pass")))
	})

	t.Run("must_fail_add_user_empty_password", func(t *testing.T) {
		setup()
		require.EqualError(t, service.AddUser(ctx, "alice", ""), ErrEmptyCredentials.Error())
		require.Empty(t, repo.CreateUserCalls())
	})

	t.Run("must_fail_set_password_unknown_user", func(t *testing.T) {
		setup()
		require.EqualError(t, service.SetPassword(ctx, "alice", "pass"), ErrUserNotFound.Error())
	})

	t.Run("must_fail_delete_unknown_user", func(t *testing.T) {
		setup()
		require.NoError(t, service.DeleteUser(ctx, "bob"))
		require.EqualError(t, service.DeleteUser(ctx, "alice"), ErrUserNotFound.Error())
	})
}
