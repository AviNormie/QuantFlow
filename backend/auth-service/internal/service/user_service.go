package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"auth-service/internal/model"
	"auth-service/internal/repository"
	"shared/jwt"
)

var (
	ErrEmailTaken      = errors.New("email already registered")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidSession  = errors.New("invalid session")
)

// UserService contains auth business logic.
type UserService struct {
	repo        *repository.UserRepository
	sessionRepo *repository.SessionRepository
}

func NewUserService(repo *repository.UserRepository, sessionRepo *repository.SessionRepository) *UserService {
	return &UserService{repo: repo, sessionRepo: sessionRepo}
}

func (s *UserService) Register(ctx context.Context, email, password string) (*model.User, error) {
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

	passwordHash, err := model.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, ErrInvalidEmail
	}
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, repository.ErrUserNotFound
	}
	return s.repo.GetUserByID(ctx, id)
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	if err := model.CheckPassword(password, user.PasswordHash); err != nil {
		return nil, repository.ErrUserNotFound
	}

	return user, nil
}

// IssueTokens creates access and refresh tokens and stores the refresh session in Redis.
func (s *UserService) IssueTokens(ctx context.Context, user *model.User) (accessToken, refreshToken string, err error) {
	accessToken, err = jwt.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", fmt.Errorf("access token: %w", err)
	}

	refreshToken, tokenID, err := jwt.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return "", "", fmt.Errorf("refresh token: %w", err)
	}

	if err := s.sessionRepo.StoreSession(ctx, user.ID, tokenID, jwt.RefreshTokenTTL()); err != nil {
		return "", "", fmt.Errorf("store session: %w", err)
	}

	return accessToken, refreshToken, nil
}

// RefreshTokens rotates refresh tokens and invalidates the previous session.
func (s *UserService) RefreshTokens(ctx context.Context, refreshToken string) (*model.User, string, string, error) {
	claims, err := jwt.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, "", "", ErrInvalidSession
	}

	exists, err := s.sessionRepo.SessionExists(ctx, claims.UserID, claims.TokenID)
	if err != nil {
		return nil, "", "", fmt.Errorf("check session: %w", err)
	}
	if !exists {
		return nil, "", "", ErrInvalidSession
	}

	if err := s.sessionRepo.DeleteSession(ctx, claims.UserID, claims.TokenID); err != nil {
		return nil, "", "", fmt.Errorf("delete session: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, "", "", err
	}

	accessToken, newRefresh, err := s.IssueTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, newRefresh, nil
}

// Logout invalidates the refresh token session in Redis.
func (s *UserService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := jwt.VerifyRefreshToken(refreshToken)
	if err != nil {
		return ErrInvalidSession
	}
	return s.sessionRepo.DeleteSession(ctx, claims.UserID, claims.TokenID)
}
