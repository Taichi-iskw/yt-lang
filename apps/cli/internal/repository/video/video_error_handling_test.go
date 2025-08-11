//go:build integration

package video

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/channel"
	"github.com/Taichi-iskw/yt-lang/internal/repository/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVideoErrorHandling tests video-specific PostgreSQL error handling
func TestVideoErrorHandling(t *testing.T) {
	// Setup real PostgreSQL using testcontainers
	pool := common.SetupTestDB(t)

	channelRepo := channel.NewRepository(pool)
	videoRepo := NewRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Setup test channel first
	testChannel := &model.Channel{
		ID:   "UC_VIDEO_ERROR_TEST",
		Name: "Video Error Test Channel",
		URL:  "https://www.youtube.com/channel/UC_VIDEO_ERROR_TEST",
	}
	err := channelRepo.Create(ctx, testChannel)
	require.NoError(t, err)

	// Test data
	video := &model.Video{
		ID:        "VIDEO_ERROR_TEST_001",
		ChannelID: testChannel.ID,
		Title:     "Video Error Test",
		URL:       "https://www.youtube.com/watch?v=VIDEO_ERROR_TEST_001",
		Duration:  300,
	}

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
			ChannelID: testChannel.ID,
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

	t.Run("Video Foreign Key Constraint Violation", func(t *testing.T) {
		// Try to create video with non-existent channel
		invalidVideo := &model.Video{
			ID:        "FK_ERROR_VIDEO",
			ChannelID: "UC_NONEXISTENT_CHANNEL",
			Title:     "FK Error Video",
			URL:       "https://www.youtube.com/watch?v=FK_ERROR_VIDEO",
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

	t.Run("UpsertBatch with Mixed Errors", func(t *testing.T) {
		// Try UpsertBatch with foreign key violations
		invalidVideos := []*model.Video{
			{
				ID:        "UPSERT_FK_ERROR_1",
				ChannelID: "UC_NONEXISTENT_UPSERT",
				Title:     "Upsert FK Error 1",
				URL:       "https://www.youtube.com/watch?v=UPSERT_FK_ERROR_1",
				Duration:  180,
			},
		}

		err := videoRepo.UpsertBatch(ctx, invalidVideos)
		require.Error(t, err)

		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.CodeDependency, appErr.Code)
		assert.Contains(t, appErr.Message, "referenced channel does not exist")
	})
}
