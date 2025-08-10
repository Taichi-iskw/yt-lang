//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnhancedErrorHandling tests the improved PostgreSQL error handling
func TestEnhancedErrorHandling(t *testing.T) {
	// Setup real PostgreSQL using testcontainers
	pool := setupTestDB(t)
	defer teardownTestDB(pool)

	channelRepo := NewChannelRepository(pool)
	videoRepo := NewVideoRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test data
	channel := &model.Channel{
		ID:   "UC_ENHANCED_TEST",
		Name: "Enhanced Error Test Channel",
		URL:  "https://www.youtube.com/channel/UC_ENHANCED_TEST",
	}

	video := &model.Video{
		ID:        "ENHANCED_VIDEO_001",
		ChannelID: channel.ID,
		Title:     "Enhanced Error Test Video",
		URL:       "https://www.youtube.com/watch?v=ENHANCED_VIDEO_001",
		Duration:  300,
	}

	t.Run("Channel Unique Constraint - ID Conflict", func(t *testing.T) {
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

	t.Run("Channel Unique Constraint - URL Conflict", func(t *testing.T) {
		// Try to create channel with different ID but same URL
		duplicateURLChannel := &model.Channel{
			ID:   "UC_DIFFERENT_ID",
			Name: "Different Channel",
			URL:  channel.URL, // Same URL
		}

		err := channelRepo.Create(ctx, duplicateURLChannel)
		require.Error(t, err)

		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.CodeConflict, appErr.Code)
		assert.Contains(t, appErr.Message, "channel with this URL already exists")
	})

	t.Run("Video Foreign Key Constraint Violation", func(t *testing.T) {
		// Try to create video with non-existent channel
		invalidVideo := &model.Video{
			ID:        "FK_ERROR_ENHANCED",
			ChannelID: "UC_NONEXISTENT_ENHANCED",
			Title:     "FK Error Video Enhanced",
			URL:       "https://www.youtube.com/watch?v=FK_ERROR_ENHANCED",
			Duration:  200,
		}

		err := videoRepo.Create(ctx, invalidVideo)
		require.Error(t, err)

		// Should be mapped to DEPENDENCY_ERROR
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.CodeDependency, appErr.Code)
		assert.Contains(t, appErr.Message, "referenced channel does not exist")
	})

	t.Run("Video ID Unique Constraint Violation", func(t *testing.T) {
		// Create video first time
		err := videoRepo.Create(ctx, video)
		require.NoError(t, err)

		// Try to create same video again - should get CONFLICT
		err = videoRepo.Create(ctx, video)
		require.Error(t, err)

		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.CodeConflict, appErr.Code)
		assert.Contains(t, appErr.Message, "video with this ID already exists")
	})

	t.Run("Video URL Unique Constraint Violation", func(t *testing.T) {
		// Create video with different ID but same URL
		duplicateURLVideo := &model.Video{
			ID:        "DIFFERENT_VIDEO_ID",
			ChannelID: channel.ID,
			Title:     "Different Video Title",
			URL:       video.URL, // Same URL
			Duration:  250,
		}

		err := videoRepo.Create(ctx, duplicateURLVideo)
		require.Error(t, err)

		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.CodeConflict, appErr.Code)
		assert.Contains(t, appErr.Message, "video with this URL already exists")
	})

	t.Run("Batch Insert Foreign Key Constraint", func(t *testing.T) {
		// Try batch insert with foreign key violations
		invalidVideos := []*model.Video{
			{
				ID:        "BATCH_FK_ERROR_1",
				ChannelID: "UC_NONEXISTENT_BATCH",
				Title:     "Batch FK Error 1",
				URL:       "https://www.youtube.com/watch?v=BATCH_FK_ERROR_1",
				Duration:  180,
			},
		}

		err := videoRepo.CreateBatch(ctx, invalidVideos)
		require.Error(t, err)

		// COPY FROM should also properly handle foreign key errors
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.CodeDependency, appErr.Code)
		assert.Contains(t, appErr.Message, "referenced channel does not exist")
	})

	t.Run("Mixed Batch Insert with Unique Violations", func(t *testing.T) {
		// Try batch insert with duplicate IDs
		duplicateVideos := []*model.Video{
			{
				ID:        "BATCH_DUPLICATE",
				ChannelID: channel.ID,
				Title:     "Batch Duplicate 1",
				URL:       "https://www.youtube.com/watch?v=BATCH_DUPLICATE_1",
				Duration:  180,
			},
			{
				ID:        "BATCH_DUPLICATE", // Same ID - should cause unique violation
				ChannelID: channel.ID,
				Title:     "Batch Duplicate 2",
				URL:       "https://www.youtube.com/watch?v=BATCH_DUPLICATE_2",
				Duration:  200,
			},
		}

		err := videoRepo.CreateBatch(ctx, duplicateVideos)
		require.Error(t, err)

		// Should be mapped to CONFLICT for unique violation
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.CodeConflict, appErr.Code)
		assert.Contains(t, appErr.Message, "video with this ID already exists")
	})
}
