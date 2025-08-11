package translation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslationRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		translation *model.Translation
		wantErr     bool
	}{
		{
			name: "successful creation",
			translation: &model.Translation{
				TranscriptionID: "1",
				TargetLanguage:  "ja",
				Content:         "こんにちは",
				Source:          "plamo",
			},
			wantErr: false,
		},
		{
			name: "duplicate translation returns error",
			translation: &model.Translation{
				TranscriptionID: "1",
				TargetLanguage:  "ja",
				Content:         "こんにちは",
				Source:          "plamo",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewTranslationRepository(mock)

			if tt.wantErr {
				// Expect constraint violation error
				mock.ExpectQuery("INSERT INTO translations").
					WithArgs(tt.translation.TranscriptionID, tt.translation.TargetLanguage,
						tt.translation.Content, tt.translation.Source).
					WillReturnError(errors.New("constraint violation"))
			} else {
				// Expect successful insert with returning ID and created_at
				rows := mock.NewRows([]string{"id", "created_at"}).
					AddRow(1, time.Now())
				mock.ExpectQuery("INSERT INTO translations").
					WithArgs(tt.translation.TranscriptionID, tt.translation.TargetLanguage,
						tt.translation.Content, tt.translation.Source).
					WillReturnRows(rows)
			}

			ctx := context.Background()
			err = repo.Create(ctx, tt.translation)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, tt.translation.ID)
				assert.NotZero(t, tt.translation.CreatedAt)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTranslationRepository_Get(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		setupMock   func(pgxmock.PgxPoolIface)
		want        *model.Translation
		wantErr     bool
		expectedErr string
	}{
		{
			name: "successful get",
			id:   1,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := mock.NewRows([]string{"id", "transcription_id", "target_language", "content", "source", "created_at"}).
					AddRow(1, "123", "ja", "こんにちは世界", "plamo", time.Now())
				mock.ExpectQuery("SELECT (.+) FROM translations WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: &model.Translation{
				ID:              1,
				TranscriptionID: "123",
				TargetLanguage:  "ja",
				Content:         "こんにちは世界",
				Source:          "plamo",
			},
			wantErr: false,
		},
		{
			name: "translation not found",
			id:   999,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM translations WHERE id = \\$1").
					WithArgs(999).
					WillReturnError(errors.New("no rows in result set"))
			},
			want:        nil,
			wantErr:     true,
			expectedErr: "no rows in result set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewTranslationRepository(mock)
			tt.setupMock(mock)

			ctx := context.Background()
			result, err := repo.Get(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.want.ID, result.ID)
				assert.Equal(t, tt.want.TranscriptionID, result.TranscriptionID)
				assert.Equal(t, tt.want.TargetLanguage, result.TargetLanguage)
				assert.Equal(t, tt.want.Content, result.Content)
				assert.Equal(t, tt.want.Source, result.Source)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTranslationRepository_GetByTranscriptionIDAndLanguage(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTranslationRepository(mock)

	transcriptionID := "1"
	targetLanguage := "ja"

	// Setup mock expectation
	rows := mock.NewRows([]string{"id", "transcription_id", "target_language", "content", "source", "created_at"}).
		AddRow(1, transcriptionID, targetLanguage, "こんにちは", "plamo", time.Now())
	mock.ExpectQuery("SELECT (.+) FROM translations WHERE transcription_id = \\$1 AND target_language = \\$2").
		WithArgs(transcriptionID, targetLanguage).
		WillReturnRows(rows)

	ctx := context.Background()
	translation, err := repo.GetByTranscriptionIDAndLanguage(ctx, transcriptionID, targetLanguage)

	require.NoError(t, err)
	require.NotNil(t, translation)
	assert.Equal(t, transcriptionID, translation.TranscriptionID)
	assert.Equal(t, targetLanguage, translation.TargetLanguage)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTranslationRepository_ListByTranscriptionID(t *testing.T) {
	tests := []struct {
		name            string
		transcriptionID string
		limit           int
		offset          int
		setupMock       func(pgxmock.PgxPoolIface)
		expectedCount   int
		wantErr         bool
		expectedErr     string
	}{
		{
			name:            "successful list with pagination",
			transcriptionID: "123",
			limit:           10,
			offset:          0,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := mock.NewRows([]string{"id", "transcription_id", "target_language", "content", "source", "created_at"}).
					AddRow(1, "123", "ja", "こんにちは", "plamo", time.Now()).
					AddRow(2, "123", "en", "hello", "plamo", time.Now())
				mock.ExpectQuery("SELECT (.+) FROM translations WHERE transcription_id = \\$1 ORDER BY created_at DESC LIMIT \\$2 OFFSET \\$3").
					WithArgs("123", 10, 0).
					WillReturnRows(rows)
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:            "empty result",
			transcriptionID: "999",
			limit:           10,
			offset:          0,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := mock.NewRows([]string{"id", "transcription_id", "target_language", "content", "source", "created_at"})
				mock.ExpectQuery("SELECT (.+) FROM translations WHERE transcription_id = \\$1 ORDER BY created_at DESC LIMIT \\$2 OFFSET \\$3").
					WithArgs("999", 10, 0).
					WillReturnRows(rows)
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name:            "database error",
			transcriptionID: "123",
			limit:           10,
			offset:          0,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM translations WHERE transcription_id = \\$1 ORDER BY created_at DESC LIMIT \\$2 OFFSET \\$3").
					WithArgs("123", 10, 0).
					WillReturnError(errors.New("database connection failed"))
			},
			expectedCount: 0,
			wantErr:       true,
			expectedErr:   "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewTranslationRepository(mock)
			tt.setupMock(mock)

			ctx := context.Background()
			result, err := repo.ListByTranscriptionID(ctx, tt.transcriptionID, tt.limit, tt.offset)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)

				// Verify transcription ID matches for non-empty results
				for _, translation := range result {
					assert.Equal(t, tt.transcriptionID, translation.TranscriptionID)
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTranslationRepository_Delete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTranslationRepository(mock)

	translationID := 1

	// Setup mock expectation for delete
	mock.ExpectExec("DELETE FROM translations WHERE id = \\$1").
		WithArgs(translationID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	ctx := context.Background()
	err = repo.Delete(ctx, translationID)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
