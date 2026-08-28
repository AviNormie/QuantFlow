package database

import (
	"context"
	"fmt"
	"os"

	"auth-service/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the shared GORM database handle for auth-service.
var DB *gorm.DB

// Connect opens GORM using DATABASE_URL and runs migrations.
func Connect(ctx context.Context) (*gorm.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if err := db.WithContext(ctx).AutoMigrate(&model.User{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	DB = db
	return db, nil
}

// Close shuts down the database pool.
func Close() {
	if DB == nil {
		return
	}
	if sqlDB, err := DB.DB(); err == nil {
		sqlDB.Close()
	}
	DB = nil
}
