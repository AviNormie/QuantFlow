package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"auth-service/internal/models"
	"auth-service/internal/repository"
)

var (
	ErrEmailTaken     = errors.New("email already registered")
	ErrInvalidEmail   = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
)

// UserService contains auth business logic. It does not talk to the database directly.
type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register creates a user with a bcrypt-hashed password.
func (s *UserService) Register(ctx context.Context, email, password string) (*models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, ErrInvalidPassword
	}

	_, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}

	passwordHash, err := models.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// GetUserByEmail loads a user by email.
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, ErrInvalidEmail
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID loads a user by id.
func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, repository.ErrUserNotFound
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticate verifies email and password. Returns the user on success.
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	if err := models.CheckPassword(password, user.PasswordHash); err != nil {
		return nil, repository.ErrUserNotFound
	}

	return user, nil
}
