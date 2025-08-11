//go:build integration

package channel

import (
	"context"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChannelRepository_Integration tests Channel Repository with real PostgreSQL
func TestChannelRepository_Integration(t *testing.T) {
	// Setup real PostgreSQL using testcontainers
	pool := common.SetupTestDB(t)

	// Create repository with real connection pool
	repo := NewRepository(pool)

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
