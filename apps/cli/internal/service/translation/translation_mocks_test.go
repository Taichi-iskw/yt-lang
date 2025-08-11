package translation

import (
	"context"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// Mock repositories for testing

// mockTranscriptionRepo mocks TranscriptionRepository
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

// mockTranslationRepo mocks TranslationRepository
type mockTranslationRepo struct {
	CreateFunc                func(ctx context.Context, translation *model.Translation) error
	GetFunc                   func(ctx context.Context, id int) (*model.Translation, error)
	ListByTranscriptionIDFunc func(ctx context.Context, transcriptionID string, limit, offset int) ([]*model.Translation, error)
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

func (m *mockTranslationRepo) ListByTranscriptionID(ctx context.Context, transcriptionID string, limit, offset int) ([]*model.Translation, error) {
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

// Batch is an alias for SegmentBatch for convenience
type Batch = SegmentBatch

// mockBatchProcessor mocks BatchProcessor interface
type mockBatchProcessor struct {
	CreateBatchesFunc       func(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error)
	SplitTranslationFunc    func(batch SegmentBatch, translatedText string) ([]*TranslationSegment, error)
	ProcessWithFallbackFunc func(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error)
}

func (m *mockBatchProcessor) CreateBatches(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error) {
	if m.CreateBatchesFunc != nil {
		return m.CreateBatchesFunc(segments, maxTokens)
	}
	return []SegmentBatch{}, nil
}

func (m *mockBatchProcessor) SplitTranslation(batch SegmentBatch, translatedText string) ([]*TranslationSegment, error) {
	if m.SplitTranslationFunc != nil {
		return m.SplitTranslationFunc(batch, translatedText)
	}
	return []*TranslationSegment{}, nil
}

func (m *mockBatchProcessor) ProcessWithFallback(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error) {
	if m.ProcessWithFallbackFunc != nil {
		return m.ProcessWithFallbackFunc(segments)
	}
	// Simple mock implementation
	result := make([]*TranslationSegment, len(segments))
	for i, seg := range segments {
		result[i] = &TranslationSegment{
			SegmentIndex:   i,
			Text:           seg.Text,
			TranslatedText: "fallback: " + seg.Text,
		}
	}
	return result, nil
}
