//go:build integration

package channel

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChannelErrorHandling tests channel-specific PostgreSQL error handling
func TestChannelErrorHandling(t *testing.T) {
	// Setup real PostgreSQL using testcontainers
	pool := common.SetupTestDB(t)

	channelRepo := NewRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test data
	channel := &model.Channel{
		ID:   "UC_CHANNEL_ERROR_TEST",
		Name: "Channel Error Test",
		URL:  "https://www.youtube.com/channel/UC_CHANNEL_ERROR_TEST",
	}

	t.Run("Channel ID Unique Constraint Violation", func(t *testing.T) {
		// Create channel first time
		err := channelRepo.Create(ctx, channel)
		require.NoError(t, err)

		// Try to create same channel again - should get CONFLICT error
		err = channelRepo.Create(ctx, channel)
		require.Error(t, err)

		// Check if it's properly mapped to CONFLICT AppError
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.CodeConflict, appErr.Code)
		assert.Contains(t, appErr.Message, "channel with this ID already exists")
	})

	t.Run("Channel URL Unique Constraint Violation", func(t *testing.T) {
		// Try to create channel with different ID but same URL
		duplicateURLChannel := &model.Channel{
			ID:   "UC_DIFFERENT_CHANNEL_ID",
			Name: "Different Channel Name",
			URL:  channel.URL, // Same URL
		}

		err := channelRepo.Create(ctx, duplicateURLChannel)
		require.Error(t, err)

		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.CodeConflict, appErr.Code)
		assert.Contains(t, appErr.Message, "channel with this URL already exists")
	})
}
