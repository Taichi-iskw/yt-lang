package channel

import (
	"context"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		channel *model.Channel
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "successful creation",
			channel: &model.Channel{
				ID:   "UC123456789",
				Name: "Test Channel",
				URL:  "https://www.youtube.com/@testchannel",
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO channels").
					WithArgs("UC123456789", "Test Channel", "https://www.youtube.com/@testchannel").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name: "database error",
			channel: &model.Channel{
				ID:   "UC123456789",
				Name: "Test Channel",
				URL:  "https://www.youtube.com/@testchannel",
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO channels").
					WithArgs("UC123456789", "Test Channel", "https://www.youtube.com/@testchannel").
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

			err = repo.Create(ctx, tt.channel)

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

func TestChannelRepository_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(mock pgxmock.PgxPoolIface)
		want    *model.Channel
		wantErr bool
	}{
		{
			name: "channel found",
			id:   "UC123456789",
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "url"}).
					AddRow("UC123456789", "Test Channel", "https://www.youtube.com/@testchannel")
				mock.ExpectQuery("SELECT id, name, url FROM channels WHERE id = \\$1").
					WithArgs("UC123456789").
					WillReturnRows(rows)
			},
			want: &model.Channel{
				ID:   "UC123456789",
				Name: "Test Channel",
				URL:  "https://www.youtube.com/@testchannel",
			},
			wantErr: false,
		},
		{
			name: "channel not found",
			id:   "UCnotfound",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, name, url FROM channels WHERE id = \\$1").
					WithArgs("UCnotfound").
					WillReturnRows(pgxmock.NewRows([]string{"id", "name", "url"}))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "database error",
			id:   "UC123456789",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, name, url FROM channels WHERE id = \\$1").
					WithArgs("UC123456789").
					WillReturnError(assert.AnError)
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

func TestChannelRepository_GetByURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		setup   func(mock pgxmock.PgxPoolIface)
		want    *model.Channel
		wantErr bool
	}{
		{
			name: "channel found by URL",
			url:  "https://www.youtube.com/@testchannel",
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "url"}).
					AddRow("UC123456789", "Test Channel", "https://www.youtube.com/@testchannel")
				mock.ExpectQuery("SELECT id, name, url FROM channels WHERE url = \\$1").
					WithArgs("https://www.youtube.com/@testchannel").
					WillReturnRows(rows)
			},
			want: &model.Channel{
				ID:   "UC123456789",
				Name: "Test Channel",
				URL:  "https://www.youtube.com/@testchannel",
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

			got, err := repo.GetByURL(ctx, tt.url)

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

func TestChannelRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		channel *model.Channel
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "successful update",
			channel: &model.Channel{
				ID:   "UC123456789",
				Name: "Updated Channel",
				URL:  "https://www.youtube.com/@updatedchannel",
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE channels SET name = \\$2, url = \\$3 WHERE id = \\$1").
					WithArgs("UC123456789", "Updated Channel", "https://www.youtube.com/@updatedchannel").
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

			err = repo.Update(ctx, tt.channel)

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

func TestChannelRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   "UC123456789",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM channels WHERE id = \\$1").
					WithArgs("UC123456789").
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
