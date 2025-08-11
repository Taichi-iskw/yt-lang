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
				TranscriptionID: 1,
				TargetLanguage:  "ja",
				Content:         "こんにちは",
				Source:          "plamo",
			},
			wantErr: false,
		},
		{
			name: "duplicate translation returns error",
			translation: &model.Translation{
				TranscriptionID: 1,
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

func TestTranslationRepository_GetByTranscriptionIDAndLanguage(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTranslationRepository(mock)

	transcriptionID := 1
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