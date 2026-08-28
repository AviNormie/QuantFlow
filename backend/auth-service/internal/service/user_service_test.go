package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"auth-service/internal/model"
	"auth-service/internal/repository"
	"auth-service/internal/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserService(t *testing.T) (*service.UserService, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(redisClient)
	return service.NewUserService(userRepo, sessionRepo), db
}

func TestRegisterAndAuthenticate(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	svc, _ := setupUserService(t)
	ctx := context.Background()

	user, err := svc.Register(ctx, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = svc.Register(ctx, "test@example.com", "password123")
	if err == nil {
		t.Fatal("expected duplicate email error")
	}

	authUser, err := svc.Authenticate(ctx, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if authUser.ID != user.ID {
		t.Fatal("user id mismatch")
	}
}

func TestRefreshAndLogout(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	svc, _ := setupUserService(t)
	ctx := context.Background()

	user, err := svc.Register(ctx, "refresh@example.com", "password123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, refresh, err := svc.IssueTokens(ctx, user)
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	_, access, newRefresh, err := svc.RefreshTokens(ctx, refresh)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if access == "" || newRefresh == "" {
		t.Fatal("expected new tokens")
	}

	if err := svc.Logout(ctx, newRefresh); err != nil {
		t.Fatalf("logout: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	_, _, _, err = svc.RefreshTokens(ctx, newRefresh)
	if err == nil {
		t.Fatal("expected invalid session after logout")
	}
}
