package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVideoRepository_CreateBatch(t *testing.T) {
	tests := []struct {
		name    string
		videos  []*model.Video
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "successful batch creation with COPY FROM",
			videos: []*model.Video{
				{
					ID:        "dQw4w9WgXcQ",
					ChannelID: "UC123456789",
					Title:     "Never Gonna Give You Up",
					URL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
					Duration:  212,
				},
				{
					ID:        "oHg5SJYRHA0",
					ChannelID: "UC123456789",
					Title:     "Never Gonna Let You Down",
					URL:       "https://www.youtube.com/watch?v=oHg5SJYRHA0",
					Duration:  233,
				},
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				// Expect CopyFrom call for bulk insert
				mock.ExpectCopyFrom(
					[]string{"videos"}, // table identifier
					[]string{"id", "channel_id", "title", "url", "duration"}, // columns
				).WillReturnResult(2) // 2 rows inserted
			},
			wantErr: false,
		},
		{
			name:   "empty batch",
			videos: []*model.Video{},
			setup: func(mock pgxmock.PgxPoolIface) {
				// No expectations for empty batch
			},
			wantErr: false,
		},
		{
			name: "database error in COPY FROM",
			videos: []*model.Video{
				{
					ID:        "dQw4w9WgXcQ",
					ChannelID: "UC123456789",
					Title:     "Never Gonna Give You Up",
					URL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
					Duration:  212,
				},
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				// Expect CopyFrom call that fails
				mock.ExpectCopyFrom(
					[]string{"videos"}, // table identifier
					[]string{"id", "channel_id", "title", "url", "duration"}, // columns
				).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup pgxmock
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			// Setup expectations
			tt.setup(mock)

			// Create repository
			repo := NewVideoRepository(mock)

			// Execute test
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = repo.CreateBatch(ctx, tt.videos)

			// Verify result
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "pgxmock expectations were not met")
		})
	}
}

func TestVideoRepository_UpsertBatch(t *testing.T) {
	tests := []struct {
		name    string
		videos  []*model.Video
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "successful upsert with no existing videos",
			videos: []*model.Video{
				{
					ID:        "video1",
					ChannelID: "UC123456789",
					Title:     "Video 1",
					URL:       "https://www.youtube.com/watch?v=video1",
					Duration:  300,
				},
				{
					ID:        "video2",
					ChannelID: "UC123456789",
					Title:     "Video 2",
					URL:       "https://www.youtube.com/watch?v=video2",
					Duration:  150,
				},
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				// First query: get existing video IDs for the channel
				mock.ExpectQuery("SELECT id FROM videos WHERE channel_id = \\$1").
					WithArgs("UC123456789").
					WillReturnRows(pgxmock.NewRows([]string{"id"})) // No existing videos

				// Second: COPY FROM for all videos (none filtered out)
				mock.ExpectCopyFrom(pgx.Identifier{"videos"}, []string{"id", "channel_id", "title", "url", "duration"}).
					WillReturnResult(2)
			},
			wantErr: false,
		},
		{
			name: "successful upsert filtering existing videos",
			videos: []*model.Video{
				{
					ID:        "video1", // existing
					ChannelID: "UC123456789",
					Title:     "Video 1",
					URL:       "https://www.youtube.com/watch?v=video1",
					Duration:  300,
				},
				{
					ID:        "video3", // new
					ChannelID: "UC123456789",
					Title:     "Video 3",
					URL:       "https://www.youtube.com/watch?v=video3",
					Duration:  200,
				},
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				// First query: get existing video IDs
				mock.ExpectQuery("SELECT id FROM videos WHERE channel_id = \\$1").
					WithArgs("UC123456789").
					WillReturnRows(pgxmock.NewRows([]string{"id"}).
						AddRow("video1")) // video1 already exists

				// Second: COPY FROM only video3 (video1 filtered out)
				mock.ExpectCopyFrom(pgx.Identifier{"videos"}, []string{"id", "channel_id", "title", "url", "duration"}).
					WillReturnResult(1)
			},
			wantErr: false,
		},
		{
			name: "all videos already exist - no COPY FROM",
			videos: []*model.Video{
				{
					ID:        "video1",
					ChannelID: "UC123456789",
					Title:     "Video 1",
					URL:       "https://www.youtube.com/watch?v=video1",
					Duration:  300,
				},
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				// First query: all videos exist
				mock.ExpectQuery("SELECT id FROM videos WHERE channel_id = \\$1").
					WithArgs("UC123456789").
					WillReturnRows(pgxmock.NewRows([]string{"id"}).
						AddRow("video1"))
				// No COPY FROM expected since all videos filtered out
			},
			wantErr: false,
		},
		{
			name:   "empty videos list",
			videos: []*model.Video{},
			setup: func(mock pgxmock.PgxPoolIface) {
				// No expectations - should return early
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup pgxmock
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			// Setup expectations
			tt.setup(mock)

			// Create repository
			repo := NewVideoRepository(mock)

			// Execute test
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = repo.UpsertBatch(ctx, tt.videos)

			// Verify result
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "pgxmock expectations were not met")
		})
	}
}
