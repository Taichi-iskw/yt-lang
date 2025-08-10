//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgreSQLErrorHandling tests specific PostgreSQL error scenarios
func TestPostgreSQLErrorHandling(t *testing.T) {
	// Setup real PostgreSQL using testcontainers  
	pool := setupTestDB(t)
	defer teardownTestDB(pool)

	channelRepo := NewChannelRepository(pool)
	videoRepo := NewVideoRepository(pool)
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test data
	channel := &model.Channel{
		ID:   "UC_ERROR_TEST",
		Name: "Error Test Channel",
		URL:  "https://www.youtube.com/channel/UC_ERROR_TEST",
	}

	video := &model.Video{
		ID:        "ERROR_VIDEO_001", 
		ChannelID: channel.ID,
		Title:     "Error Test Video",
		URL:       "https://www.youtube.com/watch?v=ERROR_VIDEO_001",
		Duration:  300,
	}

	t.Run("Channel Unique Constraint Violation", func(t *testing.T) {
		// Create channel first time - should succeed
		err := channelRepo.Create(ctx, channel)
		require.NoError(t, err)

		// Try to create same channel again - should get UNIQUE violation
		err = channelRepo.Create(ctx, channel)
		require.Error(t, err)
		
		// Check if it's a PostgreSQL error
		var pgErr *pgconn.PgError
		if assert.ErrorAs(t, err, &pgErr) {
			t.Logf("PostgreSQL Error Code: %s", pgErr.Code)
			t.Logf("PostgreSQL Error Message: %s", pgErr.Message)
			t.Logf("PostgreSQL Error Detail: %s", pgErr.Detail)
			t.Logf("PostgreSQL Constraint Name: %s", pgErr.ConstraintName)
			
			// Should be UNIQUE_VIOLATION (23505)
			assert.Equal(t, "23505", pgErr.Code)
		}
	})

	t.Run("Video Foreign Key Constraint Violation", func(t *testing.T) {
		// Try to create video with non-existent channel
		invalidVideo := &model.Video{
			ID:        "FK_ERROR_VIDEO",
			ChannelID: "UC_NONEXISTENT",  // This channel doesn't exist
			Title:     "FK Error Video",
			URL:       "https://www.youtube.com/watch?v=FK_ERROR_VIDEO",
			Duration:  200,
		}

		err := videoRepo.Create(ctx, invalidVideo)
		require.Error(t, err)

		// Check if it's a PostgreSQL foreign key error
		var pgErr *pgconn.PgError
		if assert.ErrorAs(t, err, &pgErr) {
			t.Logf("PostgreSQL Error Code: %s", pgErr.Code)
			t.Logf("PostgreSQL Error Message: %s", pgErr.Message)
			t.Logf("PostgreSQL Error Detail: %s", pgErr.Detail)
			t.Logf("PostgreSQL Constraint Name: %s", pgErr.ConstraintName)
			
			// Should be FOREIGN_KEY_VIOLATION (23503)
			assert.Equal(t, "23503", pgErr.Code)
		}
	})

	t.Run("Video Unique Constraint Violation", func(t *testing.T) {
		// Create video first time - should succeed
		err := videoRepo.Create(ctx, video)
		require.NoError(t, err)

		// Try to create same video again - should get UNIQUE violation
		err = videoRepo.Create(ctx, video)
		require.Error(t, err)

		var pgErr *pgconn.PgError
		if assert.ErrorAs(t, err, &pgErr) {
			t.Logf("PostgreSQL Error Code: %s", pgErr.Code)
			t.Logf("PostgreSQL Error Message: %s", pgErr.Message)
			t.Logf("PostgreSQL Constraint Name: %s", pgErr.ConstraintName)
			
			// Should be UNIQUE_VIOLATION (23505)
			assert.Equal(t, "23505", pgErr.Code)
		}
	})

	t.Run("Video URL Unique Constraint Violation", func(t *testing.T) {
		// Create video with different ID but same URL
		duplicateURLVideo := &model.Video{
			ID:        "DIFFERENT_ID",
			ChannelID: channel.ID,
			Title:     "Different Video",
			URL:       video.URL, // Same URL as previous video
			Duration:  250,
		}

		err := videoRepo.Create(ctx, duplicateURLVideo)
		require.Error(t, err)

		var pgErr *pgconn.PgError
		if assert.ErrorAs(t, err, &pgErr) {
			t.Logf("PostgreSQL Error Code: %s", pgErr.Code)
			t.Logf("PostgreSQL Error Message: %s", pgErr.Message)
			t.Logf("PostgreSQL Constraint Name: %s", pgErr.ConstraintName)
			
			// Should be UNIQUE_VIOLATION (23505) on URL
			assert.Equal(t, "23505", pgErr.Code)
			assert.Contains(t, pgErr.ConstraintName, "url") // URL constraint
		}
	})

	t.Run("BatchInsert Error Handling", func(t *testing.T) {
		// Try batch insert with foreign key violation
		invalidVideos := []*model.Video{
			{
				ID:        "BATCH_ERROR_1", 
				ChannelID: "UC_NONEXISTENT_BATCH",
				Title:     "Batch Error 1",
				URL:       "https://www.youtube.com/watch?v=BATCH_ERROR_1",
				Duration:  180,
			},
			{
				ID:        "BATCH_ERROR_2",
				ChannelID: "UC_NONEXISTENT_BATCH", 
				Title:     "Batch Error 2",
				URL:       "https://www.youtube.com/watch?v=BATCH_ERROR_2",
				Duration:  200,
			},
		}

		err := videoRepo.CreateBatch(ctx, invalidVideos)
		require.Error(t, err)

		var pgErr *pgconn.PgError
		if assert.ErrorAs(t, err, &pgErr) {
			t.Logf("Batch Insert Error Code: %s", pgErr.Code)
			t.Logf("Batch Insert Error Message: %s", pgErr.Message)
			
			// COPY FROM should also catch foreign key violations
			assert.Equal(t, "23503", pgErr.Code)
		}
	})
}