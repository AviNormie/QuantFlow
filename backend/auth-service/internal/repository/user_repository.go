package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

// UserRepository handles user persistence through GORM.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	return &user, nil
}

// SessionRepository stores refresh-token sessions in Redis.
type SessionRepository struct {
	client *redis.Client
}

func NewSessionRepository(client *redis.Client) *SessionRepository {
	return &SessionRepository{client: client}
}

func sessionKey(userID, tokenID string) string {
	return fmt.Sprintf("session:%s:%s", userID, tokenID)
}

func (r *SessionRepository) StoreSession(ctx context.Context, userID, tokenID string, ttl time.Duration) error {
	return r.client.Set(ctx, sessionKey(userID, tokenID), "1", ttl).Err()
}

func (r *SessionRepository) DeleteSession(ctx context.Context, userID, tokenID string) error {
	return r.client.Del(ctx, sessionKey(userID, tokenID)).Err()
}

func (r *SessionRepository) SessionExists(ctx context.Context, userID, tokenID string) (bool, error) {
	count, err := r.client.Exists(ctx, sessionKey(userID, tokenID)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
