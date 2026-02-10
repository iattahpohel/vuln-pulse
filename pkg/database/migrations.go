package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// getMigrationsDir returns the absolute path to the migrations directory
func getMigrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	return filepath.Join(repoRoot, "migrations")
}

// RunMigrations applies all pending migrations
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Convert pgxpool to *sql.DB for goose
	connConfig := pool.Config().ConnConfig
	connStr := stdlib.RegisterConnConfig(connConfig)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	migrationsDir := getMigrationsDir()
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// RollbackMigrations rolls back the last migration
func RollbackMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	connConfig := pool.Config().ConnConfig
	connStr := stdlib.RegisterConnConfig(connConfig)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	migrationsDir := getMigrationsDir()
	if err := goose.Down(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}
