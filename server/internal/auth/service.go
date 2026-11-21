package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

//go:generate moq -out user_repository_moq_test.go . userRepository
type userRepository interface {
	UserByLogin(ctx context.Context, login string) (*User, error)
	CreateUser(ctx context.Context, login, hash string) error
	SetPassword(ctx context.Context, login, hash string) (bool, error)
	DeleteUser(ctx context.Context, login string) (bool, error)
}

// Service authenticates users and issues JWT tokens.
type Service struct {
	repo   userRepository
	secret []byte
	ttl    time.Duration
}

// NewService creates a new Service.
func NewService(repo userRepository, secret string, ttl time.Duration) *Service {
	return &Service{repo: repo, secret: []byte(secret), ttl: ttl}
}

// Login checks credentials and returns a signed token.
func (s *Service) Login(ctx context.Context, req *LoginRequest) (string, error) {
	u, err := s.repo.UserByLogin(ctx, req.Login)
	if err != nil {
		return "", err
	}
	if u == nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		return "", ErrInvalidCredentials
	}
	claims := jwt.RegisteredClaims{
		Subject:   u.Login,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.ttl)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return token, nil
}

// Verify validates the token and returns the login of an existing user.
func (s *Service) Verify(ctx context.Context, token string) (string, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (any, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return "", ErrInvalidToken
	}
	u, err := s.repo.UserByLogin(ctx, claims.Subject)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", ErrInvalidToken
	}
	return u.Login, nil
}

// AddUser creates a user with the given password.
func (s *Service) AddUser(ctx context.Context, login, password string) error {
	hash, err := hashPassword(login, password)
	if err != nil {
		return err
	}
	return s.repo.CreateUser(ctx, login, hash)
}

// SetPassword changes the password of an existing user.
func (s *Service) SetPassword(ctx context.Context, login, password string) error {
	hash, err := hashPassword(login, password)
	if err != nil {
		return err
	}
	ok, err := s.repo.SetPassword(ctx, login, hash)
	if err == nil && !ok {
		return ErrUserNotFound
	}
	return err
}

// DeleteUser removes an existing user.
func (s *Service) DeleteUser(ctx context.Context, login string) error {
	ok, err := s.repo.DeleteUser(ctx, login)
	if err == nil && !ok {
		return ErrUserNotFound
	}
	return err
}

func hashPassword(login, password string) (string, error) {
	if login == "" || password == "" {
		return "", ErrEmptyCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}
