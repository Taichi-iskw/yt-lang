package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDatabasePool creates a new PostgreSQL connection pool
func NewDatabasePool(ctx context.Context, config *Config) (*pgxpool.Pool, error) {
	dbConfig, err := config.ParseDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Create pgxpool config
	poolConfig, err := pgxpool.ParseConfig(dbConfig.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool settings
	poolConfig.MaxConns = dbConfig.MaxConns
	poolConfig.MinConns = dbConfig.MinConns
	poolConfig.MaxConnLifetime = dbConfig.MaxConnLifetime
	poolConfig.MaxConnIdleTime = dbConfig.MaxConnIdleTime

	// Create connection pool with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// CloseDatabasePool gracefully closes the database connection pool
func CloseDatabasePool(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}
