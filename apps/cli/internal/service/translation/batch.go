package translation

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
	TranscriptionSegmentID string
	SegmentIndex           int
	Text                   string
	TranslatedText         string
}

// BatchProcessor handles batching and splitting of translation segments
type BatchProcessor interface {
	CreateBatches(segments []*model.TranscriptionSegment, maxTokens int) ([]SegmentBatch, error)
	TranslateBatchWithFallback(batch SegmentBatch, plamoService PlamoService, ctx context.Context, sourceLang, targetLang string) ([]*TranslationSegment, error)
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

// TranslateBatchWithFallback implements the three-stage fallback strategy
func (bp *batchProcessor) TranslateBatchWithFallback(batch SegmentBatch, plamoService PlamoService, ctx context.Context, sourceLang, targetLang string) ([]*TranslationSegment, error) {
	// Stage 1: Try with "__" separator
	fmt.Println("First Try: Translate with __ separator")
	result, err := bp.tryTranslateWithSeparator(batch.Segments, "__", plamoService, ctx, sourceLang, targetLang)
	if err == nil {
		return result, nil
	}

	// Stage 2: Try with "<<<SEP>>>" separator
	fmt.Println("Second Try: Translate with <<<SEP>>> separator")
	result, err = bp.tryTranslateWithSeparator(batch.Segments, "<<<SEP>>>", plamoService, ctx, sourceLang, targetLang)
	if err == nil {
		return result, nil
	}

	// Stage 3: Individual translation fallback
	fmt.Println("Third Try: Translate with individual translation fallback")
	return bp.translateIndividually(batch.Segments, plamoService, ctx, sourceLang, targetLang)
}

// tryTranslateWithSeparator tries to translate segments using a specific separator
func (bp *batchProcessor) tryTranslateWithSeparator(segments []*model.TranscriptionSegment, separator string, plamoService PlamoService, ctx context.Context, sourceLang, targetLang string) ([]*TranslationSegment, error) {
	// Create batch with specified separator
	batch := SegmentBatch{
		Segments:  segments,
		Separator: separator,
	}
	bp.finalizeBatch(&batch)

	// Translate the combined text
	translatedText, err := plamoService.Translate(ctx, batch.CombinedText, sourceLang, targetLang)
	if err != nil {
		return nil, err
	}

	// Split and validate
	return bp.splitAndValidateTranslation(batch, translatedText)
}

// splitAndValidateTranslation splits translated text and validates segment count
func (bp *batchProcessor) splitAndValidateTranslation(batch SegmentBatch, translation string) ([]*TranslationSegment, error) {
	// Split translation by separator
	translatedTexts := strings.Split(translation, batch.Separator)

	// Check if the number of segments matches
	if len(translatedTexts) != len(batch.Segments) {
		return nil, errors.New("segment count mismatch: expected " + strconv.Itoa(len(batch.Segments)) + " but got " + strconv.Itoa(len(translatedTexts)))
	}

	// Create translation segments
	var results []*TranslationSegment
	for i, segment := range batch.Segments {
		result := &TranslationSegment{
			TranscriptionSegmentID: segment.ID,
			SegmentIndex:           segment.SegmentIndex,
			Text:                   segment.Text,
			TranslatedText:         strings.TrimSpace(translatedTexts[i]),
		}
		results = append(results, result)
	}

	return results, nil
}

// translateIndividually translates each segment individually as final fallback
func (bp *batchProcessor) translateIndividually(segments []*model.TranscriptionSegment, plamoService PlamoService, ctx context.Context, sourceLang, targetLang string) ([]*TranslationSegment, error) {
	var results []*TranslationSegment

	for _, segment := range segments {
		// Translate individual segment
		translatedText, err := plamoService.Translate(ctx, segment.Text, sourceLang, targetLang)
		if err != nil {
			// If individual translation fails, use the original text as fallback
			translatedText = segment.Text
		}

		result := &TranslationSegment{
			TranscriptionSegmentID: segment.ID,
			SegmentIndex:           segment.SegmentIndex,
			Text:                   segment.Text,
			TranslatedText:         strings.TrimSpace(translatedText),
		}
		results = append(results, result)
	}

	return results, nil
}
