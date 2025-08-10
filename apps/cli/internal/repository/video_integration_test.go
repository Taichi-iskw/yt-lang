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

// TestVideoRepository_Integration tests Video Repository with real PostgreSQL
func TestVideoRepository_Integration(t *testing.T) {
	// Setup real PostgreSQL using testcontainers
	pool := setupTestDBForVideo(t)
	defer teardownTestDB(pool)

	// Create repository with real connection pool
	repo := NewVideoRepository(pool)

	// Test data - first create a channel
	channel := &model.Channel{
		ID:   "UC123456789",
		Name: "Test Channel",
		URL:  "https://www.youtube.com/channel/UC123456789",
	}

	channelRepo := NewChannelRepository(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := channelRepo.Create(ctx, channel)
	require.NoError(t, err)

	// Test video data
	video := &model.Video{
		ID:        "dQw4w9WgXcQ",
		ChannelID: channel.ID,
		Title:     "Never Gonna Give You Up",
		URL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		Duration:  212,
	}

	t.Run("Create and GetByID", func(t *testing.T) {
		// Create video
		err := repo.Create(ctx, video)
		require.NoError(t, err)

		// Retrieve video
		retrieved, err := repo.GetByID(ctx, video.ID)
		require.NoError(t, err)
		assert.Equal(t, video.ID, retrieved.ID)
		assert.Equal(t, video.ChannelID, retrieved.ChannelID)
		assert.Equal(t, video.Title, retrieved.Title)
		assert.Equal(t, video.URL, retrieved.URL)
		assert.Equal(t, video.Duration, retrieved.Duration)
	})

	t.Run("GetByChannelID", func(t *testing.T) {
		videos, err := repo.GetByChannelID(ctx, channel.ID, 10, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, videos)
		assert.Equal(t, video.ID, videos[0].ID)
	})

	t.Run("CreateBatch with COPY FROM", func(t *testing.T) {
		// Test batch insert with COPY FROM protocol
		batchVideos := []*model.Video{
			{
				ID:        "oHg5SJYRHA0",
				ChannelID: channel.ID,
				Title:     "Batch Video 1",
				URL:       "https://www.youtube.com/watch?v=oHg5SJYRHA0",
				Duration:  233,
			},
			{
				ID:        "iik25wqIuFo",
				ChannelID: channel.ID,
				Title:     "Batch Video 2",
				URL:       "https://www.youtube.com/watch?v=iik25wqIuFo",
				Duration:  185,
			},
		}

		// Use COPY FROM for bulk insert
		err := repo.CreateBatch(ctx, batchVideos)
		require.NoError(t, err)

		// Verify batch insert worked
		allVideos, err := repo.GetByChannelID(ctx, channel.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, allVideos, 3) // original + 2 batch videos
	})

	t.Run("Update", func(t *testing.T) {
		// Update video title
		video.Title = "Updated Video Title"
		err := repo.Update(ctx, video)
		require.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetByID(ctx, video.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Video Title", retrieved.Title)
	})

	t.Run("List with pagination", func(t *testing.T) {
		videos, err := repo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, videos)
		assert.GreaterOrEqual(t, len(videos), 3) // At least 3 videos created
	})

	t.Run("Delete", func(t *testing.T) {
		// Delete video
		err := repo.Delete(ctx, video.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(ctx, video.ID)
		assert.Error(t, err) // Should return NOT_FOUND error
	})
}

// setupTestDBForVideo creates a real PostgreSQL database for video testing (separate from channel test)
func setupTestDBForVideo(t *testing.T) Pool {
	ctx := context.Background()

	// Define PostgreSQL container request
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testvideodb",
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
	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testvideodb?sslmode=disable", host, port.Port())

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

