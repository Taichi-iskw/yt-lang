package translation

import (
	"errors"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// SegmentBatch represents a batch of segments to be translated together
type SegmentBatch struct {
	Segments     []*model.TranscriptionSegment
	CombinedText string
	Separator    string
}

// TranslationSegment represents a translated segment
type TranslationSegment struct {
	ID              string
	TranscriptionID string
	SegmentIndex    int
	Text            string
	TranslatedText  string
}

// BatchProcessor handles batching and splitting of translation segments
type BatchProcessor interface {
	CreateBatches(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error)
	SplitTranslation(batch SegmentBatch, translation string) ([]*TranslationSegment, error)
	ProcessWithFallback(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error)
}

// batchProcessor implements BatchProcessor
type batchProcessor struct {
	separators []string
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor() BatchProcessor {
	return &batchProcessor{
		separators: []string{"__", "<<<SEP>>>"},
	}
}

// estimateTokenCount estimates token count for text based on language
func estimateTokenCount(text string, language string) int {
	switch language {
	case "en":
		return len(text) / 4 // English: ~4 chars ≈ 1 token
	case "ja":
		return len([]rune(text)) / 2 // Japanese: ~2 chars ≈ 1 token
	default:
		return len([]rune(text)) / 2 // Conservative estimate
	}
}

// CreateBatches creates batches of segments that fit within token limits
func (bp *batchProcessor) CreateBatches(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error) {
	if maxTokens <= 0 {
		return nil, errors.New("maxTokens must be positive")
	}

	if len(segments) == 0 {
		return []SegmentBatch{}, nil
	}

	var batches []SegmentBatch
	separator := bp.separators[0] // Use first separator "__"

	// Simple implementation: try to fit segments in batches
	currentBatch := SegmentBatch{
		Segments:  []*model.TranscriptionSegment{},
		Separator: separator,
	}
	currentTokens := 0

	for _, segment := range segments {
		segmentTokens := estimateTokenCount(segment.Text, "en") // Default to English

		// If adding this segment would exceed limit, start new batch
		if currentTokens+segmentTokens > maxTokens && len(currentBatch.Segments) > 0 {
			// Finalize current batch
			bp.finalizeBatch(&currentBatch)
			batches = append(batches, currentBatch)

			// Start new batch
			currentBatch = SegmentBatch{
				Segments:  []*model.TranscriptionSegment{segment},
				Separator: separator,
			}
			currentTokens = segmentTokens
		} else {
			currentBatch.Segments = append(currentBatch.Segments, segment)
			currentTokens += segmentTokens
		}
	}

	// Add final batch if it has segments
	if len(currentBatch.Segments) > 0 {
		bp.finalizeBatch(&currentBatch)
		batches = append(batches, currentBatch)
	}

	return batches, nil
}

// finalizeBatch sets the CombinedText for a batch
func (bp *batchProcessor) finalizeBatch(batch *SegmentBatch) {
	var texts []string
	for _, segment := range batch.Segments {
		texts = append(texts, segment.Text)
	}
	batch.CombinedText = strings.Join(texts, batch.Separator)
}

// SplitTranslation splits translated text back into individual segments
func (bp *batchProcessor) SplitTranslation(batch SegmentBatch, translation string) ([]*TranslationSegment, error) {
	// Check separator count
	originalCount := strings.Count(batch.CombinedText, batch.Separator)
	translatedCount := strings.Count(translation, batch.Separator)

	if originalCount != translatedCount {
		return nil, errors.New("separator count mismatch")
	}

	// Split translation by separator
	translatedTexts := strings.Split(translation, batch.Separator)

	// Create translation segments
	var results []*TranslationSegment
	for i, segment := range batch.Segments {
		if i >= len(translatedTexts) {
			return nil, errors.New("insufficient translated texts")
		}

		result := &TranslationSegment{
			ID:              segment.ID,
			TranscriptionID: segment.TranscriptionID,
			SegmentIndex:    segment.SegmentIndex,
			Text:            segment.Text,
			TranslatedText:  strings.TrimSpace(translatedTexts[i]),
		}
		results = append(results, result)
	}

	return results, nil
}

// ProcessWithFallback processes segments with fallback strategy
func (bp *batchProcessor) ProcessWithFallback(segments []*model.TranscriptionSegment) ([]*TranslationSegment, error) {
	// Fallback strategy: try different separators, then individual translation
	// This is a placeholder implementation
	var results []*TranslationSegment
	
	for _, segment := range segments {
		result := &TranslationSegment{
			ID:              segment.ID,
			TranscriptionID: segment.TranscriptionID,
			SegmentIndex:    segment.SegmentIndex,
			Text:            segment.Text,
			TranslatedText:  "fallback_" + segment.Text, // Placeholder translation
		}
		results = append(results, result)
	}
	
	return results, nil
}
