package translation

import (
	"context"
	"errors"
	"testing"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock repositories
type mockTranscriptionRepo struct {
	GetSegmentsFunc func(ctx context.Context, transcriptionID string) ([]*model.TranscriptionSegment, error)
	GetFunc         func(ctx context.Context, id string) (*model.Transcription, error)
}

func (m *mockTranscriptionRepo) GetSegments(ctx context.Context, transcriptionID string) ([]*model.TranscriptionSegment, error) {
	if m.GetSegmentsFunc != nil {
		return m.GetSegmentsFunc(ctx, transcriptionID)
	}
	return nil, nil
}

func (m *mockTranscriptionRepo) Get(ctx context.Context, id string) (*model.Transcription, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

type mockTranslationRepo struct {
	CreateFunc                func(ctx context.Context, translation *model.Translation) error
	GetFunc                   func(ctx context.Context, id int) (*model.Translation, error)
	ListByTranscriptionIDFunc func(ctx context.Context, transcriptionID int, limit, offset int) ([]*model.Translation, error)
	DeleteFunc                func(ctx context.Context, id int) error
}

func (m *mockTranslationRepo) Create(ctx context.Context, translation *model.Translation) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, translation)
	}
	return nil
}

func (m *mockTranslationRepo) Get(ctx context.Context, id int) (*model.Translation, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockTranslationRepo) ListByTranscriptionID(ctx context.Context, transcriptionID int, limit, offset int) ([]*model.Translation, error) {
	if m.ListByTranscriptionIDFunc != nil {
		return m.ListByTranscriptionIDFunc(ctx, transcriptionID, limit, offset)
	}
	return []*model.Translation{}, nil
}

func (m *mockTranslationRepo) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func TestTranslationService_CreateTranslation(t *testing.T) {
	tests := []struct {
		name            string
		transcriptionID string
		targetLang      string
		setupMocks      func(*mockTranscriptionRepo, *mockTranslationRepo, *MockCmdRunner, *mockBatchProcessor)
		wantErr         bool
		errMessage      string
	}{
		{
			name:            "successful translation with batching",
			transcriptionID: "trans-123",
			targetLang:      "ja",
			setupMocks: func(tr *mockTranscriptionRepo, tlr *mockTranslationRepo, pr *MockCmdRunner, bp *mockBatchProcessor) {
				// Setup transcription segments
				tr.GetSegmentsFunc = func(ctx context.Context, id string) ([]*model.TranscriptionSegment, error) {
					return []*model.TranscriptionSegment{
						{ID: "seg-1", TranscriptionID: "trans-123", Text: "Hello world"},
						{ID: "seg-2", TranscriptionID: "trans-123", Text: "Good morning"},
					}, nil
				}

				// Setup batch processor
				bp.CreateBatchesFunc = func(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error) {
					return []SegmentBatch{
						{
							Segments:     segments,
							CombinedText: "Hello world__Good morning",
							Separator:    "__",
						},
					}, nil
				}

				// Setup PLaMo translation
				pr.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
					return []byte("こんにちは世界__おはようございます"), nil
				}

				// Setup split translation
				bp.SplitTranslationFunc = func(batch SegmentBatch, translation string) ([]*TranslationSegment, error) {
					return []*TranslationSegment{
						{Text: "Hello world", TranslatedText: "こんにちは世界"},
						{Text: "Good morning", TranslatedText: "おはようございます"},
					}, nil
				}

				// Setup repository save
				tlr.CreateFunc = func(ctx context.Context, translation *model.Translation) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:            "no segments found",
			transcriptionID: "trans-empty",
			targetLang:      "ja",
			setupMocks: func(tr *mockTranscriptionRepo, tlr *mockTranslationRepo, pr *MockCmdRunner, bp *mockBatchProcessor) {
				tr.GetSegmentsFunc = func(ctx context.Context, id string) ([]*model.TranscriptionSegment, error) {
					return []*model.TranscriptionSegment{}, nil
				}
			},
			wantErr:    true,
			errMessage: "no segments found",
		},
		{
			name:            "PLaMo translation fails",
			transcriptionID: "trans-456",
			targetLang:      "ja",
			setupMocks: func(tr *mockTranscriptionRepo, tlr *mockTranslationRepo, pr *MockCmdRunner, bp *mockBatchProcessor) {
				tr.GetSegmentsFunc = func(ctx context.Context, id string) ([]*model.TranscriptionSegment, error) {
					return []*model.TranscriptionSegment{
						{ID: "seg-1", Text: "Hello"},
					}, nil
				}

				bp.CreateBatchesFunc = func(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error) {
					return []SegmentBatch{
						{Segments: segments, CombinedText: "Hello", Separator: "__"},
					}, nil
				}

				pr.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
					return nil, errors.New("PLaMo service unavailable")
				}

				// Fallback also fails
				bp.ProcessWithFallbackFunc = func(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error) {
					return nil, errors.New("fallback failed")
				}
			},
			wantErr:    true,
			errMessage: "PLaMo service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			transcriptionRepo := &mockTranscriptionRepo{}
			translationRepo := &mockTranslationRepo{}
			plamoService := NewPlamoService(&MockCmdRunner{})
			batchProcessor := &mockBatchProcessor{}

			// Setup mocks
			if tt.setupMocks != nil {
				cmdRunner := &MockCmdRunner{}
				plamoService = NewPlamoService(cmdRunner)
				tt.setupMocks(transcriptionRepo, translationRepo, cmdRunner, batchProcessor)
			}

			// Create service
			service := NewTranslationService(
				transcriptionRepo,
				translationRepo,
				plamoService,
				batchProcessor,
			)

			// Execute
			ctx := context.Background()
			translation, err := service.CreateTranslation(ctx, tt.transcriptionID, tt.targetLang)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
				assert.Nil(t, translation)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, translation)
				assert.Equal(t, tt.targetLang, translation.TargetLanguage)
			}
		})
	}
}

// Mock batch processor
type mockBatchProcessor struct {
	CreateBatchesFunc       func(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error)
	SplitTranslationFunc    func(batch SegmentBatch, translation string) ([]*TranslationSegment, error)
	ProcessWithFallbackFunc func(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error)
}

func (m *mockBatchProcessor) CreateBatches(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error) {
	if m.CreateBatchesFunc != nil {
		return m.CreateBatchesFunc(segments, maxTokens)
	}
	return nil, nil
}

func (m *mockBatchProcessor) SplitTranslation(batch SegmentBatch, translation string) ([]*TranslationSegment, error) {
	if m.SplitTranslationFunc != nil {
		return m.SplitTranslationFunc(batch, translation)
	}
	return nil, nil
}

func (m *mockBatchProcessor) ProcessWithFallback(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error) {
	if m.ProcessWithFallbackFunc != nil {
		return m.ProcessWithFallbackFunc(segments)
	}
	return nil, nil
}

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
						ID:              123,
						TranscriptionID: 456,
						TargetLanguage:  "ja",
						Content:         "こんにちは世界",
						Source:          "plamo",
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
			// Create mocks
			translationRepo := &mockTranslationRepo{}
			if tt.setupMocks != nil {
				tt.setupMocks(translationRepo)
			}

			// Create service
			service := NewTranslationService(
				&mockTranscriptionRepo{},
				translationRepo,
				NewPlamoService(&MockCmdRunner{}),
				&mockBatchProcessor{},
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
				// Segments are nil for now since we're not implementing segment retrieval yet
				assert.Nil(t, segments)
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
				tr.ListByTranscriptionIDFunc = func(ctx context.Context, transcriptionID int, limit, offset int) ([]*model.Translation, error) {
					return []*model.Translation{
						{ID: 1, TranscriptionID: 123, TargetLanguage: "ja", Content: "こんにちは", Source: "plamo"},
						{ID: 2, TranscriptionID: 123, TargetLanguage: "en", Content: "hello", Source: "plamo"},
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
				tr.ListByTranscriptionIDFunc = func(ctx context.Context, transcriptionID int, limit, offset int) ([]*model.Translation, error) {
					return []*model.Translation{}, nil
				}
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name:            "invalid transcription ID",
			transcriptionID: "invalid",
			limit:           10,
			offset:          0,
			setupMocks: func(tr *mockTranslationRepo) {
				// No setup needed - error should occur before repository call
			},
			expectedCount: 0,
			wantErr:       true,
			expectedErr:   "invalid transcription ID",
		},
		{
			name:            "repository error",
			transcriptionID: "123",
			limit:           10,
			offset:          0,
			setupMocks: func(tr *mockTranslationRepo) {
				tr.ListByTranscriptionIDFunc = func(ctx context.Context, transcriptionID int, limit, offset int) ([]*model.Translation, error) {
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
			// Create mocks
			translationRepo := &mockTranslationRepo{}
			if tt.setupMocks != nil {
				tt.setupMocks(translationRepo)
			}

			// Create service
			service := NewTranslationService(
				&mockTranscriptionRepo{},
				translationRepo,
				NewPlamoService(&MockCmdRunner{}),
				&mockBatchProcessor{},
			)

			// Execute
			ctx := context.Background()
			result, err := service.ListTranslations(ctx, tt.transcriptionID, tt.limit, tt.offset)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
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
					return errors.New("deletion failed")
				}
			},
			wantErr:     true,
			expectedErr: "deletion failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			translationRepo := &mockTranslationRepo{}
			if tt.setupMocks != nil {
				tt.setupMocks(translationRepo)
			}

			// Create service
			service := NewTranslationService(
				&mockTranscriptionRepo{},
				translationRepo,
				NewPlamoService(&MockCmdRunner{}),
				&mockBatchProcessor{},
			)

			// Execute
			ctx := context.Background()
			err := service.DeleteTranslation(ctx, tt.id)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
