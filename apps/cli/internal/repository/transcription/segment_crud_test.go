package transcription

import (
	"context"
	"testing"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSegmentRepository_CreateBatch(t *testing.T) {
	tests := []struct {
		name     string
		segments []*model.TranscriptionSegment
		setup    func(mock pgxmock.PgxPoolIface)
		wantErr  bool
	}{
		{
			name: "successful batch creation",
			segments: []*model.TranscriptionSegment{
				{
					ID:              "seg-1",
					TranscriptionID: "trans-123",
					SegmentIndex:    0,
					StartTime:       "00:00:00",
					EndTime:         "00:00:02.5",
					Text:            "Hello, this is a test.",
					Confidence:      floatPtr(0.95),
				},
				{
					ID:              "seg-2",
					TranscriptionID: "trans-123",
					SegmentIndex:    1,
					StartTime:       "00:00:02.5",
					EndTime:         "00:00:06",
					Text:            "We're learning Go.",
					Confidence:      floatPtr(0.92),
				},
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectCopyFrom(pgx.Identifier{"transcription_segments"},
					[]string{"transcription_id", "segment_index", "start_time", "end_time", "text", "confidence"}).
					WillReturnResult(2)
			},
			wantErr: false,
		},
		{
			name:     "empty segments",
			segments: []*model.TranscriptionSegment{},
			setup: func(mock pgxmock.PgxPoolIface) {
				// No expectation for empty segments
			},
			wantErr: false,
		},
		{
			name: "database error",
			segments: []*model.TranscriptionSegment{
				{
					ID:              "seg-1",
					TranscriptionID: "trans-123",
					SegmentIndex:    0,
					StartTime:       "00:00:00",
					EndTime:         "00:00:02.5",
					Text:            "Test segment",
					Confidence:      nil,
				},
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectCopyFrom(pgx.Identifier{"transcription_segments"},
					[]string{"transcription_id", "segment_index", "start_time", "end_time", "text", "confidence"}).
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

			repo := NewSegmentRepository(mock)
			err = repo.CreateBatch(context.Background(), tt.segments)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSegmentRepository_GetByTranscriptionID(t *testing.T) {
	tests := []struct {
		name            string
		transcriptionID string
		setup           func(mock pgxmock.PgxPoolIface)
		wantSegments    int
		wantErr         bool
	}{
		{
			name:            "successful get segments",
			transcriptionID: "trans-123",
			setup: func(mock pgxmock.PgxPoolIface) {
				conf1 := 0.95
				conf2 := 0.92
				rows := pgxmock.NewRows([]string{
					"id", "transcription_id", "segment_index", "start_time", "end_time", "text", "confidence",
				}).
					AddRow("seg-1", "trans-123", 0, "00:00:00", "00:00:02.5", "Hello, this is a test.", &conf1).
					AddRow("seg-2", "trans-123", 1, "00:00:02.5", "00:00:06", "We're learning Go.", &conf2)

				mock.ExpectQuery("SELECT (.+) FROM transcription_segments WHERE transcription_id").
					WithArgs("trans-123").
					WillReturnRows(rows)
			},
			wantSegments: 2,
			wantErr:      false,
		},
		{
			name:            "no segments found",
			transcriptionID: "trans-456",
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "transcription_id", "segment_index", "start_time", "end_time", "text", "confidence",
				})

				mock.ExpectQuery("SELECT (.+) FROM transcription_segments WHERE transcription_id").
					WithArgs("trans-456").
					WillReturnRows(rows)
			},
			wantSegments: 0,
			wantErr:      false,
		},
		{
			name:            "database error",
			transcriptionID: "trans-789",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM transcription_segments WHERE transcription_id").
					WithArgs("trans-789").
					WillReturnError(assert.AnError)
			},
			wantSegments: 0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.setup(mock)

			repo := NewSegmentRepository(mock)
			segments, err := repo.GetByTranscriptionID(context.Background(), tt.transcriptionID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, segments)
			} else {
				assert.NoError(t, err)
				assert.Len(t, segments, tt.wantSegments)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Helper function to create float64 pointer
func floatPtr(f float64) *float64 {
	return &f
}
