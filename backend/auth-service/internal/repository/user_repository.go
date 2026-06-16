package repository

import (
	"context"
	"errors"
	"fmt"

	"auth-service/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

// UserRepository handles persistence for users.
type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// CreateUser inserts a user. Password must already be hashed; set user.Email and user.PasswordHash.
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// GetUserByEmail returns a user by email or ErrUserNotFound.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user, err := r.scanUser(ctx, query, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID returns a user by id or ErrUserNotFound.
func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	const query = `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user, err := r.scanUser(ctx, query, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) scanUser(ctx context.Context, query string, arg any) (*models.User, error) {
	var user models.User

	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	return &user, nil
}
