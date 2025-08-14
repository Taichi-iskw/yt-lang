package translation

import (
	"context"
	"errors"
	"testing"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslationService_GetTranslation(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setupMocks func(*mockTranslationRepo)
		wantErr    bool
		expectNil  bool
		expectedID int
	}{
		{
			name: "successful get translation",
			id:   "123",
			setupMocks: func(tr *mockTranslationRepo) {
				tr.GetFunc = func(ctx context.Context, id int) (*model.Translation, error) {
					return &model.Translation{
						ID:                     123,
						TranscriptionSegmentID: "456",
						TargetLanguage:         "ja",
						TranslatedText:         "こんにちは世界",
						Source:                 "plamo",
					}, nil
				}
			},
			wantErr:    false,
			expectNil:  false,
			expectedID: 123,
		},
		{
			name: "translation not found",
			id:   "999",
			setupMocks: func(tr *mockTranslationRepo) {
				tr.GetFunc = func(ctx context.Context, id int) (*model.Translation, error) {
					return nil, errors.New("translation not found")
				}
			},
			wantErr:   true,
			expectNil: true,
		},
		{
			name: "invalid id format",
			id:   "invalid",
			setupMocks: func(tr *mockTranslationRepo) {
				// No setup needed - error should occur before repository call
			},
			wantErr:   true,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockTranscriptionRepo := &mockTranscriptionRepo{}
			mockTranslationRepo := &mockTranslationRepo{}
			mockCmdRunner := &MockCmdRunner{}
			mockBatchProcessor := &mockBatchProcessor{}

			tt.setupMocks(mockTranslationRepo)

			// Create service
			plamoService := NewPlamoService(mockCmdRunner)
			service := NewTranslationService(
				mockTranscriptionRepo,
				mockTranslationRepo,
				plamoService,
				mockBatchProcessor,
			)

			// Execute
			ctx := context.Background()
			translation, segments, err := service.GetTranslation(ctx, tt.id)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, translation)
				assert.Nil(t, segments)
			} else {
				assert.NotNil(t, translation)
				assert.Equal(t, tt.expectedID, translation.ID)
				// Now we implement segment retrieval, so segments should be available
				assert.NotNil(t, segments)
			}
		})
	}
}

func TestTranslationService_ListTranslations(t *testing.T) {
	tests := []struct {
		name            string
		transcriptionID string
		limit           int
		offset          int
		setupMocks      func(*mockTranslationRepo)
		expectedCount   int
		wantErr         bool
		expectedErr     string
	}{
		{
			name:            "successful list translations",
			transcriptionID: "123",
			limit:           10,
			offset:          0,
			setupMocks: func(tr *mockTranslationRepo) {
				tr.ListByTranscriptionIDFunc = func(ctx context.Context, transcriptionID string, limit, offset int) ([]*model.Translation, error) {
					return []*model.Translation{
						{ID: 1, TranscriptionSegmentID: "seg-123", TargetLanguage: "ja", TranslatedText: "こんにちは", Source: "plamo"},
						{ID: 2, TranscriptionSegmentID: "seg-124", TargetLanguage: "en", TranslatedText: "hello", Source: "plamo"},
					}, nil
				}
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:            "empty list",
			transcriptionID: "999",
			limit:           10,
			offset:          0,
			setupMocks: func(tr *mockTranslationRepo) {
				tr.ListByTranscriptionIDFunc = func(ctx context.Context, transcriptionID string, limit, offset int) ([]*model.Translation, error) {
					return []*model.Translation{}, nil
				}
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name:            "repository error",
			transcriptionID: "123",
			limit:           10,
			offset:          0,
			setupMocks: func(tr *mockTranslationRepo) {
				tr.ListByTranscriptionIDFunc = func(ctx context.Context, transcriptionID string, limit, offset int) ([]*model.Translation, error) {
					return nil, errors.New("database error")
				}
			},
			expectedCount: 0,
			wantErr:       true,
			expectedErr:   "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockTranscriptionRepo := &mockTranscriptionRepo{}
			mockTranslationRepo := &mockTranslationRepo{}
			mockCmdRunner := &MockCmdRunner{}
			mockBatchProcessor := &mockBatchProcessor{}

			tt.setupMocks(mockTranslationRepo)

			// Create service
			plamoService := NewPlamoService(mockCmdRunner)
			service := NewTranslationService(
				mockTranscriptionRepo,
				mockTranslationRepo,
				plamoService,
				mockBatchProcessor,
			)

			// Execute
			ctx := context.Background()
			translations, err := service.ListTranslations(ctx, tt.transcriptionID, tt.limit, tt.offset)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Len(t, translations, tt.expectedCount)
			}
		})
	}
}

func TestTranslationService_DeleteTranslation(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		setupMocks  func(*mockTranslationRepo)
		wantErr     bool
		expectedErr string
	}{
		{
			name: "successful delete",
			id:   "123",
			setupMocks: func(tr *mockTranslationRepo) {
				tr.DeleteFunc = func(ctx context.Context, id int) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "invalid id format",
			id:   "invalid",
			setupMocks: func(tr *mockTranslationRepo) {
				// No setup needed - error should occur before repository call
			},
			wantErr:     true,
			expectedErr: "invalid translation ID",
		},
		{
			name: "repository error",
			id:   "123",
			setupMocks: func(tr *mockTranslationRepo) {
				tr.DeleteFunc = func(ctx context.Context, id int) error {
					return errors.New("database error")
				}
			},
			wantErr:     true,
			expectedErr: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockTranscriptionRepo := &mockTranscriptionRepo{}
			mockTranslationRepo := &mockTranslationRepo{}
			mockCmdRunner := &MockCmdRunner{}
			mockBatchProcessor := &mockBatchProcessor{}

			tt.setupMocks(mockTranslationRepo)

			// Create service
			plamoService := NewPlamoService(mockCmdRunner)
			service := NewTranslationService(
				mockTranscriptionRepo,
				mockTranslationRepo,
				plamoService,
				mockBatchProcessor,
			)

			// Execute
			ctx := context.Background()
			err := service.DeleteTranslation(ctx, tt.id)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
