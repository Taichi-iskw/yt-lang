package translation

import (
	"context"
	"errors"
	"testing"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

				// Setup translation with new implementation
				bp.TranslateBatchWithFallbackFunc = func(batch SegmentBatch, plamoService PlamoService, ctx context.Context, sourceLang, targetLang string) ([]*TranslationSegment, error) {
					return []*TranslationSegment{
						{Text: "Hello world", TranslatedText: "こんにちは世界"},
						{Text: "Good morning", TranslatedText: "おはようございます"},
					}, nil
				}

				// Setup repository save
				tlr.CreateBatchFunc = func(ctx context.Context, translations []*model.Translation) error {
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

				// Translation fails
				bp.TranslateBatchWithFallbackFunc = func(batch SegmentBatch, plamoService PlamoService, ctx context.Context, sourceLang, targetLang string) ([]*TranslationSegment, error) {
					return nil, errors.New("all translation strategies failed")
				}
			},
			wantErr:    true,
			errMessage: "batch translation failed",
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
