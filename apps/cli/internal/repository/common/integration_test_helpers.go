//go:build integration

package repository

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// runMigrations executes database migrations using real migration files
func runMigrations(databaseURL string) error {
	// Get the directory of this source file
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)

	// Navigate from apps/cli/internal/repository to migrations
	migrationsPath := filepath.Join(currentDir, "..", "..", "..", "..", "migrations")
	migrationsPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path to migrations: %w", err)
	}
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	// Create migrate instance
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
