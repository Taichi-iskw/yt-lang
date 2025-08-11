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
	CreateFunc func(ctx context.Context, translation *model.Translation) error
}

func (m *mockTranslationRepo) Create(ctx context.Context, translation *model.Translation) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, translation)
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
	CreateBatchesFunc     func(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error)
	SplitTranslationFunc  func(batch SegmentBatch, translation string) ([]*TranslationSegment, error)
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