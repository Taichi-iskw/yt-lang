package transcription

import (
	"context"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranscriptionRepository_Create(t *testing.T) {
	tests := []struct {
		name          string
		transcription *model.Transcription
		setup         func(mock pgxmock.PgxPoolIface)
		wantErr       bool
	}{
		{
			name: "successful creation",
			transcription: &model.Transcription{
				ID:               "trans-123",
				VideoID:          "video-456",
				Language:         "auto",
				Status:           "pending",
				CreatedAt:        time.Now(),
				CompletedAt:      nil,
				ErrorMessage:     nil,
				DetectedLanguage: nil,
				TotalDuration:    nil,
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO transcriptions").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("generated-uuid"))
			},
			wantErr: false,
		},
		{
			name: "database error",
			transcription: &model.Transcription{
				ID:        "trans-123",
				VideoID:   "video-456",
				Language:  "auto",
				Status:    "pending",
				CreatedAt: time.Now(),
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO transcriptions").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.setup(mock)

			repo := NewRepository(mock)
			err = repo.Create(context.Background(), tt.transcription)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify that ID was populated by database
				assert.NotEmpty(t, tt.transcription.ID)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTranscriptionRepository_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(mock pgxmock.PgxPoolIface)
		want    *model.Transcription
		wantErr bool
	}{
		{
			name: "successful get",
			id:   "trans-123",
			setup: func(mock pgxmock.PgxPoolIface) {
				now := time.Now()
				detectedLang := "en"
				duration := "00:10:30"
				rows := pgxmock.NewRows([]string{
					"id", "video_id", "language", "status", "created_at",
					"completed_at", "error_message", "detected_language", "total_duration",
				}).AddRow(
					"trans-123", "video-456", "auto", "completed", now,
					&now, nil, &detectedLang, &duration,
				)
				mock.ExpectQuery("SELECT (.+) FROM transcriptions WHERE id").
					WithArgs("trans-123").
					WillReturnRows(rows)
			},
			want: &model.Transcription{
				ID:       "trans-123",
				VideoID:  "video-456",
				Language: "auto",
				Status:   "completed",
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   "trans-nonexistent",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM transcriptions WHERE id").
					WithArgs("trans-nonexistent").
					WillReturnRows(pgxmock.NewRows([]string{"id", "video_id", "language", "status", "created_at", "completed_at", "error_message", "detected_language", "total_duration"}))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.setup(mock)

			repo := NewRepository(mock)
			result, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.want.ID, result.ID)
				assert.Equal(t, tt.want.VideoID, result.VideoID)
				assert.Equal(t, tt.want.Language, result.Language)
				assert.Equal(t, tt.want.Status, result.Status)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
