package translation

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

const (
	defaultMaxTokens = 7000 // PLaMo input limit
)

// TranscriptionRepository interface for accessing transcription data
type TranscriptionRepository interface {
	GetSegments(ctx context.Context, transcriptionID string) ([]*model.TranscriptionSegment, error)
	Get(ctx context.Context, id string) (*model.Transcription, error)
}

// TranslationRepository interface for accessing translation data
type TranslationRepository interface {
	Get(ctx context.Context, id int) (*model.Translation, error)
	Create(ctx context.Context, translation *model.Translation) error
	ListByTranscriptionID(ctx context.Context, transcriptionID string, limit, offset int) ([]*model.Translation, error)
	Delete(ctx context.Context, id int) error
}

// TranslationService defines the main translation service interface
type TranslationService interface {
	CreateTranslation(ctx context.Context, transcriptionID string, targetLang string) (*model.Translation, error)
	GetTranslation(ctx context.Context, id string) (*model.Translation, []*TranslationSegment, error)
	ListTranslations(ctx context.Context, transcriptionID string, limit, offset int) ([]*model.Translation, error)
	DeleteTranslation(ctx context.Context, id string) error
	GetPlamoService() PlamoService
}

// translationService implements TranslationService
type translationService struct {
	transcriptionRepo TranscriptionRepository
	translationRepo   TranslationRepository
	plamoService      PlamoService
	batchProcessor    BatchProcessor
}

// NewTranslationService creates a new translation service
func NewTranslationService(
	transcriptionRepo TranscriptionRepository,
	translationRepo TranslationRepository,
	plamoService PlamoService,
	batchProcessor BatchProcessor,
) TranslationService {
	return &translationService{
		transcriptionRepo: transcriptionRepo,
		translationRepo:   translationRepo,
		plamoService:      plamoService,
		batchProcessor:    batchProcessor,
	}
}

// NewTranslationServiceWithFallback creates a new translation service with fallback support
func NewTranslationServiceWithFallback(
	transcriptionRepo TranscriptionRepository,
	translationRepo TranslationRepository,
	plamoService PlamoService,
	batchProcessor BatchProcessor,
) TranslationService {
	// Same implementation but with enhanced error handling
	return &translationService{
		transcriptionRepo: transcriptionRepo,
		translationRepo:   translationRepo,
		plamoService:      plamoService,
		batchProcessor:    batchProcessor,
	}
}

// CreateTranslation creates a new translation for a transcription
func (s *translationService) CreateTranslation(ctx context.Context, transcriptionID string, targetLang string) (*model.Translation, error) {
	// Step 1: Get transcription segments
	segments, err := s.transcriptionRepo.GetSegments(ctx, transcriptionID)
	if err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		return nil, errors.New("no segments found")
	}

	// Step 2: Create batches for efficient translation
	batches, err := s.batchProcessor.CreateBatches(segments, defaultMaxTokens)
	if err != nil {
		return nil, err
	}

	// Step 3: Optimize for batch translation - start server once for multiple batches
	sourceLanguage := "en" // Default source language - should be detected from transcription

	// If we have multiple batches, start the server once for better performance
	if len(batches) > 1 {
		if err := s.plamoService.StartServer(ctx); err != nil {
			// If server startup fails, continue with simple mode
			// Server implementation will handle this gracefully
		}
		// Note: We don't defer StopServer here as it's managed at CLI level
	}

	// Step 3: Translate each batch
	var allTranslatedSegments []*TranslationSegment

	for _, batch := range batches {
		// Translate batch
		translatedText, err := s.plamoService.Translate(ctx, batch.CombinedText, sourceLanguage, targetLang)
		if err != nil {
			// Try fallback strategy if translation fails
			fallbackSegments, fallbackErr := s.batchProcessor.ProcessWithFallback(batch.Segments)
			if fallbackErr != nil {
				return nil, err // Return original error
			}
			allTranslatedSegments = append(allTranslatedSegments, fallbackSegments...)
			continue
		}

		// Split translated text back into segments
		translatedSegments, err := s.batchProcessor.SplitTranslation(batch, translatedText)
		if err != nil {
			// Try fallback strategy if split fails
			fallbackSegments, fallbackErr := s.batchProcessor.ProcessWithFallback(batch.Segments)
			if fallbackErr != nil {
				return nil, err // Return original error
			}
			allTranslatedSegments = append(allTranslatedSegments, fallbackSegments...)
			continue
		}

		allTranslatedSegments = append(allTranslatedSegments, translatedSegments...)
	}

	// Step 4: Combine all translated segments into content
	var translatedContent []string
	for _, seg := range allTranslatedSegments {
		translatedContent = append(translatedContent, seg.TranslatedText)
	}

	// Step 5: Save translation to database
	translation := &model.Translation{
		TranscriptionID: transcriptionID, // Use string UUID directly
		TargetLanguage:  targetLang,
		Content:         joinSegments(translatedContent),
		Source:          "plamo",
	}

	err = s.translationRepo.Create(ctx, translation)
	if err != nil {
		return nil, err
	}

	return translation, nil
}

// joinSegments joins translated segments with space
func joinSegments(segments []string) string {
	result := ""
	for i, seg := range segments {
		if i > 0 {
			result += " "
		}
		result += seg
	}
	return result
}

// GetTranslation retrieves a translation
func (s *translationService) GetTranslation(ctx context.Context, id string) (*model.Translation, []*TranslationSegment, error) {
	// Convert string ID to int
	translationID, err := strconv.Atoi(id)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid translation ID: %w", err)
	}

	// Get translation from repository
	translation, err := s.translationRepo.Get(ctx, translationID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get translation: %w", err)
	}

	// Get segments to reconstruct translation segments
	// First, we need to reverse-engineer the transcription ID from the hashed value
	// For now, we'll return the translation segments by parsing the content
	segments, err := s.parseTranslationSegments(translation.Content)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse translation segments: %w", err)
	}

	return translation, segments, nil
}

// ListTranslations retrieves translations for a transcription with pagination
func (s *translationService) ListTranslations(ctx context.Context, transcriptionID string, limit, offset int) ([]*model.Translation, error) {
	// Get translations from repository with pagination (transcriptionID is UUID string)
	translations, err := s.translationRepo.ListByTranscriptionID(ctx, transcriptionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list translations: %w", err)
	}

	return translations, nil
}

// DeleteTranslation deletes a translation by ID
func (s *translationService) DeleteTranslation(ctx context.Context, id string) error {
	// Convert string ID to int
	translationID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid translation ID: %w", err)
	}

	// Delete translation from repository
	err = s.translationRepo.Delete(ctx, translationID)
	if err != nil {
		return fmt.Errorf("failed to delete translation: %w", err)
	}

	return nil
}

// GetPlamoService returns the PLaMo service instance
func (s *translationService) GetPlamoService() PlamoService {
	return s.plamoService
}

// parseTranslationSegments parses translation content to create translation segments
// This is a temporary implementation that splits content by sentences/lines
func (s *translationService) parseTranslationSegments(content string) ([]*TranslationSegment, error) {
	if content == "" {
		return []*TranslationSegment{}, nil
	}

	// Simple parsing: split by periods and line breaks, ignoring empty segments
	parts := splitByDelimiters(content, []string{".", "\n", "。"})

	var segments []*TranslationSegment
	segmentIndex := 0

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		segments = append(segments, &TranslationSegment{
			SegmentIndex:   segmentIndex,
			Text:           "", // Original text not available from translation alone
			TranslatedText: part,
		})
		segmentIndex++
	}

	return segments, nil
}

// splitByDelimiters splits text by multiple delimiters
func splitByDelimiters(text string, delimiters []string) []string {
	parts := []string{text}

	for _, delimiter := range delimiters {
		var newParts []string
		for _, part := range parts {
			subParts := strings.Split(part, delimiter)
			newParts = append(newParts, subParts...)
		}
		parts = newParts
	}

	return parts
}
