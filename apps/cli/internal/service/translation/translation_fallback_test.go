package translation

import (
	"context"
	"errors"
	"testing"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslationService_FallbackStrategy(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*mockTranscriptionRepo, *MockCmdRunner, *mockBatchProcessorWithFallback)
		wantErr    bool
		verify     func(*testing.T, *model.Translation)
	}{
		{
			name: "fallback from __ to <<<SEP>>> separator",
			setupMocks: func(tr *mockTranscriptionRepo, pr *MockCmdRunner, bp *mockBatchProcessorWithFallback) {
				// Setup segments
				tr.GetSegmentsFunc = func(ctx context.Context, id string) ([]*model.TranscriptionSegment, error) {
					return []*model.TranscriptionSegment{
						{ID: "1", Text: "Hello"},
						{ID: "2", Text: "World"},
					}, nil
				}
				
				// First attempt with "__" fails on split
				callCount := 0
				bp.CreateBatchesFunc = func(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error) {
					return []SegmentBatch{
						{
							Segments:     segments,
							CombinedText: "Hello__World",
							Separator:    "__",
						},
					}, nil
				}
				
				pr.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
					callCount++
					if callCount == 1 {
						// First translation removes separator
						return []byte("こんにちは世界"), nil
					}
					// Second attempt with different separator
					return []byte("こんにちは<<<SEP>>>世界"), nil
				}
				
				bp.SplitTranslationFunc = func(batch SegmentBatch, translation string) ([]*TranslationSegment, error) {
					if batch.Separator == "__" && translation == "こんにちは世界" {
						// Separator mismatch - trigger fallback
						return nil, errors.New("separator count mismatch")
					}
					// Success with <<<SEP>>>
					return []*TranslationSegment{
						{Text: "Hello", TranslatedText: "こんにちは"},
						{Text: "World", TranslatedText: "世界"},
					}, nil
				}
				
				// Fallback processor recreates batch with new separator
				bp.ProcessWithFallbackFunc = func(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error) {
					// Simulate successful fallback processing
					return []*TranslationSegment{
						{Text: "Hello", TranslatedText: "こんにちは"},
						{Text: "World", TranslatedText: "世界"},
					}, nil
				}
			},
			wantErr: false,
			verify: func(t *testing.T, translation *model.Translation) {
				assert.Contains(t, translation.Content, "こんにちは")
				assert.Contains(t, translation.Content, "世界")
			},
		},
		{
			name: "fallback to individual translation",
			setupMocks: func(tr *mockTranscriptionRepo, pr *MockCmdRunner, bp *mockBatchProcessorWithFallback) {
				tr.GetSegmentsFunc = func(ctx context.Context, id string) ([]*model.TranscriptionSegment, error) {
					return []*model.TranscriptionSegment{
						{ID: "1", Text: "Complex text"},
					}, nil
				}
				
				bp.CreateBatchesFunc = func(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error) {
					return []SegmentBatch{
						{
							Segments:     segments,
							CombinedText: "Complex text",
							Separator:    "__",
						},
					}, nil
				}
				
				// All batch attempts fail
				pr.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
					return []byte("複雑なテキスト"), nil
				}
				
				bp.SplitTranslationFunc = func(batch SegmentBatch, translation string) ([]*TranslationSegment, error) {
					return nil, errors.New("unable to split")
				}
				
				// Fallback to individual processing succeeds
				bp.ProcessWithFallbackFunc = func(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error) {
					return []*TranslationSegment{
						{Text: "Complex text", TranslatedText: "複雑なテキスト"},
					}, nil
				}
			},
			wantErr: false,
			verify: func(t *testing.T, translation *model.Translation) {
				assert.Contains(t, translation.Content, "複雑なテキスト")
			},
		},
		{
			name: "retry on transient PLaMo error",
			setupMocks: func(tr *mockTranscriptionRepo, pr *MockCmdRunner, bp *mockBatchProcessorWithFallback) {
				tr.GetSegmentsFunc = func(ctx context.Context, id string) ([]*model.TranscriptionSegment, error) {
					return []*model.TranscriptionSegment{
						{ID: "1", Text: "Test"},
					}, nil
				}
				
				bp.CreateBatchesFunc = func(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error) {
					return []SegmentBatch{
						{
							Segments:     segments,
							CombinedText: "Test",
							Separator:    "__",
						},
					}, nil
				}
				
				// First call fails, second succeeds (retry)
				callCount := 0
				pr.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
					callCount++
					if callCount == 1 {
						return nil, errors.New("PLaMo service temporarily unavailable")
					}
					return []byte("テスト"), nil
				}
				
				bp.SplitTranslationFunc = func(batch SegmentBatch, translation string) ([]*TranslationSegment, error) {
					return []*TranslationSegment{
						{Text: "Test", TranslatedText: "テスト"},
					}, nil
				}
				
				bp.ProcessWithFallbackFunc = func(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error) {
					// Retry logic built into ProcessWithFallback
					return []*TranslationSegment{
						{Text: "Test", TranslatedText: "テスト"},
					}, nil
				}
			},
			wantErr: false,
			verify: func(t *testing.T, translation *model.Translation) {
				assert.Contains(t, translation.Content, "テスト")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			transcriptionRepo := &mockTranscriptionRepo{}
			cmdRunner := &MockCmdRunner{}
			batchProcessor := &mockBatchProcessorWithFallback{}
			translationRepo := &mockTranslationRepo{
				CreateFunc: func(ctx context.Context, translation *model.Translation) error {
					return nil
				},
			}
			
			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(transcriptionRepo, cmdRunner, batchProcessor)
			}
			
			// Create service with fallback support
			service := NewTranslationServiceWithFallback(
				transcriptionRepo,
				translationRepo,
				NewPlamoService(cmdRunner),
				batchProcessor,
			)
			
			// Execute
			ctx := context.Background()
			translation, err := service.CreateTranslation(ctx, "trans-123", "ja")
			
			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, translation)
				if tt.verify != nil {
					tt.verify(t, translation)
				}
			}
		})
	}
}

// Mock batch processor with fallback support
type mockBatchProcessorWithFallback struct {
	CreateBatchesFunc        func(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error)
	SplitTranslationFunc     func(batch SegmentBatch, translation string) ([]*TranslationSegment, error)
	ProcessWithFallbackFunc  func(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error)
}

func (m *mockBatchProcessorWithFallback) CreateBatches(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error) {
	if m.CreateBatchesFunc != nil {
		return m.CreateBatchesFunc(segments, maxTokens)
	}
	return nil, nil
}

func (m *mockBatchProcessorWithFallback) SplitTranslation(batch SegmentBatch, translation string) ([]*TranslationSegment, error) {
	if m.SplitTranslationFunc != nil {
		return m.SplitTranslationFunc(batch, translation)
	}
	return nil, nil
}

func (m *mockBatchProcessorWithFallback) ProcessWithFallback(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error) {
	if m.ProcessWithFallbackFunc != nil {
		return m.ProcessWithFallbackFunc(segments)
	}
	return nil, nil
}