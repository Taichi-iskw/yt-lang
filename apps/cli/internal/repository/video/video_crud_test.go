package video

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
				Duration:  212.0,
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO videos").
					WithArgs("dQw4w9WgXcQ", "UC123456789", "Never Gonna Give You Up", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 212.0).
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
				Duration:  212.0,
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO videos").
					WithArgs("dQw4w9WgXcQ", "UC123456789", "Never Gonna Give You Up", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 212.0).
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
			repo := NewRepository(mock)

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
				Duration:  212.0,
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
			repo := NewRepository(mock)

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
				Duration:  220.0,
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE videos SET channel_id = \\$2, title = \\$3, url = \\$4, duration = \\$5 WHERE id = \\$1").
					WithArgs("dQw4w9WgXcQ", "UC123456789", "Updated Title", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", 220.0).
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
			repo := NewRepository(mock)

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
			repo := NewRepository(mock)

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
