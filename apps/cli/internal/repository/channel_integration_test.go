//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestChannelRepository_Integration tests Channel Repository with real PostgreSQL
func TestChannelRepository_Integration(t *testing.T) {
	// Setup real PostgreSQL using testcontainers
	pool := setupTestDB(t)
	defer teardownTestDB(pool)

	// Create repository with real connection pool
	repo := NewChannelRepository(pool)

	// Test data
	channel := &model.Channel{
		ID:   "UC123456789",
		Name: "Test Channel",
		URL:  "https://www.youtube.com/channel/UC123456789",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Create and GetByID", func(t *testing.T) {
		// Create channel
		err := repo.Create(ctx, channel)
		require.NoError(t, err)

		// Retrieve channel
		retrieved, err := repo.GetByID(ctx, channel.ID)
		require.NoError(t, err)
		assert.Equal(t, channel.ID, retrieved.ID)
		assert.Equal(t, channel.Name, retrieved.Name)
		assert.Equal(t, channel.URL, retrieved.URL)
	})

	t.Run("GetByURL", func(t *testing.T) {
		retrieved, err := repo.GetByURL(ctx, channel.URL)
		require.NoError(t, err)
		assert.Equal(t, channel.ID, retrieved.ID)
		assert.Equal(t, channel.Name, retrieved.Name)
		assert.Equal(t, channel.URL, retrieved.URL)
	})

	t.Run("Update", func(t *testing.T) {
		// Update channel name
		channel.Name = "Updated Channel Name"
		err := repo.Update(ctx, channel)
		require.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetByID(ctx, channel.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Channel Name", retrieved.Name)
	})

	t.Run("List with pagination", func(t *testing.T) {
		channels, err := repo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, channels)
		assert.Equal(t, channel.ID, channels[0].ID)
	})

	t.Run("Delete", func(t *testing.T) {
		// Delete channel
		err := repo.Delete(ctx, channel.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(ctx, channel.ID)
		assert.Error(t, err) // Should return NOT_FOUND error
	})
}

// setupTestDB creates a real PostgreSQL database for testing
func setupTestDB(t *testing.T) Pool {
	ctx := context.Background()

	// Define PostgreSQL container request
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	// Start PostgreSQL container
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get host and port
	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)
	
	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Build connection string
	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	// Run migrations using real migration files
	err = runMigrations(connStr)
	require.NoError(t, err)

	// Create connection pool
	config, err := pgxpool.ParseConfig(connStr)
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)

	// Store container for cleanup
	t.Cleanup(func() {
		if pool != nil {
			pool.Close()
		}
		if postgresContainer != nil {
			postgresContainer.Terminate(ctx)
		}
	})

	return pool
}


// teardownTestDB cleans up the test database
func teardownTestDB(pool Pool) {
	// Cleanup is handled by t.Cleanup() in setupTestDB
}