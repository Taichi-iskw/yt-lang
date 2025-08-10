package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVideoRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		video   *model.Video
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "successful creation",
			video: &model.Video{
				ID:        "dQw4w9WgXcQ",
				ChannelID: "UC123456789",
				Title:     "Never Gonna Give You Up",
				URL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				Duration:  212,
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO videos").
					WithArgs("dQw4w9WgXcQ", "UC123456789", "Never Gonna Give You Up", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 212).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name: "database error",
			video: &model.Video{
				ID:        "dQw4w9WgXcQ",
				ChannelID: "UC123456789",
				Title:     "Never Gonna Give You Up",
				URL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				Duration:  212,
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO videos").
					WithArgs("dQw4w9WgXcQ", "UC123456789", "Never Gonna Give You Up", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 212).
					WillReturnError(assert.AnError)
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

			err = repo.Create(ctx, tt.video)

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

func TestVideoRepository_CreateBatch(t *testing.T) {
	tests := []struct {
		name    string
		videos  []*model.Video
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "successful batch creation with multiple inserts",
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
				// Expect multiple INSERT calls for batch operation
				mock.ExpectExec("INSERT INTO videos").
					WithArgs("dQw4w9WgXcQ", "UC123456789", "Never Gonna Give You Up", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 212).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectExec("INSERT INTO videos").
					WithArgs("oHg5SJYRHA0", "UC123456789", "Never Gonna Let You Down", "https://www.youtube.com/watch?v=oHg5SJYRHA0", 233).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
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
			name: "database error in batch",
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
				mock.ExpectExec("INSERT INTO videos").
					WithArgs("dQw4w9WgXcQ", "UC123456789", "Never Gonna Give You Up", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 212).
					WillReturnError(assert.AnError)
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

func TestVideoRepository_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(mock pgxmock.PgxPoolIface)
		want    *model.Video
		wantErr bool
	}{
		{
			name: "video found",
			id:   "dQw4w9WgXcQ",
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "channel_id", "title", "url", "duration"}).
					AddRow("dQw4w9WgXcQ", "UC123456789", "Never Gonna Give You Up", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 212)
				mock.ExpectQuery("SELECT id, channel_id, title, url, duration FROM videos WHERE id = \\$1").
					WithArgs("dQw4w9WgXcQ").
					WillReturnRows(rows)
			},
			want: &model.Video{
				ID:        "dQw4w9WgXcQ",
				ChannelID: "UC123456789",
				Title:     "Never Gonna Give You Up",
				URL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				Duration:  212,
			},
			wantErr: false,
		},
		{
			name: "video not found",
			id:   "notfound",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, channel_id, title, url, duration FROM videos WHERE id = \\$1").
					WithArgs("notfound").
					WillReturnRows(pgxmock.NewRows([]string{"id", "channel_id", "title", "url", "duration"}))
			},
			want:    nil,
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

			got, err := repo.GetByID(ctx, tt.id)

			// Verify result
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			// Verify all expectations were met
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "pgxmock expectations were not met")
		})
	}
}

func TestVideoRepository_GetByChannelID(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		limit     int
		offset    int
		setup     func(mock pgxmock.PgxPoolIface)
		want      []*model.Video
		wantErr   bool
	}{
		{
			name:      "videos found for channel",
			channelID: "UC123456789",
			limit:     2,
			offset:    0,
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "channel_id", "title", "url", "duration"}).
					AddRow("dQw4w9WgXcQ", "UC123456789", "Never Gonna Give You Up", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 212).
					AddRow("oHg5SJYRHA0", "UC123456789", "Never Gonna Let You Down", "https://www.youtube.com/watch?v=oHg5SJYRHA0", 233)
				mock.ExpectQuery("SELECT id, channel_id, title, url, duration FROM videos WHERE channel_id = \\$1 ORDER BY id LIMIT \\$2 OFFSET \\$3").
					WithArgs("UC123456789", 2, 0).
					WillReturnRows(rows)
			},
			want: []*model.Video{
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
			wantErr: false,
		},
		{
			name:      "no videos found for channel",
			channelID: "UCnotfound",
			limit:     10,
			offset:    0,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, channel_id, title, url, duration FROM videos WHERE channel_id = \\$1 ORDER BY id LIMIT \\$2 OFFSET \\$3").
					WithArgs("UCnotfound", 10, 0).
					WillReturnRows(pgxmock.NewRows([]string{"id", "channel_id", "title", "url", "duration"}))
			},
			want:    []*model.Video{},
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

			got, err := repo.GetByChannelID(ctx, tt.channelID, tt.limit, tt.offset)

			// Verify result
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			// Verify all expectations were met
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "pgxmock expectations were not met")
		})
	}
}

func TestVideoRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		video   *model.Video
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "successful update",
			video: &model.Video{
				ID:        "dQw4w9WgXcQ",
				ChannelID: "UC123456789",
				Title:     "Updated Title",
				URL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				Duration:  220,
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE videos SET channel_id = \\$2, title = \\$3, url = \\$4, duration = \\$5 WHERE id = \\$1").
					WithArgs("dQw4w9WgXcQ", "UC123456789", "Updated Title", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 220).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
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

			err = repo.Update(ctx, tt.video)

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

func TestVideoRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   "dQw4w9WgXcQ",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM videos WHERE id = \\$1").
					WithArgs("dQw4w9WgXcQ").
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
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

			err = repo.Delete(ctx, tt.id)

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

func TestVideoRepository_List(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		offset  int
		setup   func(mock pgxmock.PgxPoolIface)
		want    []*model.Video
		wantErr bool
	}{
		{
			name:   "successful list with pagination",
			limit:  2,
			offset: 0,
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "channel_id", "title", "url", "duration"}).
					AddRow("dQw4w9WgXcQ", "UC123456789", "Never Gonna Give You Up", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 212).
					AddRow("oHg5SJYRHA0", "UC123456789", "Never Gonna Let You Down", "https://www.youtube.com/watch?v=oHg5SJYRHA0", 233)
				mock.ExpectQuery("SELECT id, channel_id, title, url, duration FROM videos ORDER BY id LIMIT \\$1 OFFSET \\$2").
					WithArgs(2, 0).
					WillReturnRows(rows)
			},
			want: []*model.Video{
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

			got, err := repo.List(ctx, tt.limit, tt.offset)

			// Verify result
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			// Verify all expectations were met
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "pgxmock expectations were not met")
		})
	}
}
