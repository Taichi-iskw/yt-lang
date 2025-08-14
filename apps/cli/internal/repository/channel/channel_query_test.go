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

func TestChannelRepository_List(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		offset  int
		setup   func(mock pgxmock.PgxPoolIface)
		want    []*model.Channel
		wantErr bool
	}{
		{
			name:   "successful list with pagination",
			limit:  2,
			offset: 0,
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "url"}).
					AddRow("UC123456789", "Test Channel 1", "https://www.youtube.com/@testchannel1").
					AddRow("UC987654321", "Test Channel 2", "https://www.youtube.com/@testchannel2")
				mock.ExpectQuery("SELECT id, name, url FROM channels ORDER BY id LIMIT \\$1 OFFSET \\$2").
					WithArgs(2, 0).
					WillReturnRows(rows)
			},
			want: []*model.Channel{
				{
					ID:   "UC123456789",
					Name: "Test Channel 1",
					URL:  "https://www.youtube.com/@testchannel1",
				},
				{
					ID:   "UC987654321",
					Name: "Test Channel 2",
					URL:  "https://www.youtube.com/@testchannel2",
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
			repo := NewRepository(mock)

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
