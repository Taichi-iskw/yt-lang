package translation

import (
	"context"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// TranslationRepository defines operations for Translation persistence
type TranslationRepository interface {
	// Get retrieves a translation by ID
	Get(ctx context.Context, id int) (*model.Translation, error)

	// Create creates a new translation for a transcription segment
	Create(ctx context.Context, translation *model.Translation) error

	// CreateBatch creates multiple translations for transcription segments
	CreateBatch(ctx context.Context, translations []*model.Translation) error

	// GetByTranscriptionID retrieves all translations for a transcription segment
	GetByTranscriptionID(ctx context.Context, transcriptionID string) ([]*model.Translation, error)

	// ListByTranscriptionID retrieves translations for a transcription segment with pagination
	ListByTranscriptionID(ctx context.Context, transcriptionID string, limit, offset int) ([]*model.Translation, error)

	// GetByTranscriptionIDAndLanguage retrieves translation for specific target language
	GetByTranscriptionIDAndLanguage(ctx context.Context, transcriptionID string, targetLanguage string) (*model.Translation, error)

	// GetByVideoIDAndLanguage retrieves all translations for a video in specific target language
	// This method joins with transcriptions table to get all translations for a video
	GetByVideoIDAndLanguage(ctx context.Context, videoID, targetLanguage string) ([]*model.Translation, error)

	// Update updates an existing translation
	Update(ctx context.Context, translation *model.Translation) error

	// Delete deletes a translation by ID
	Delete(ctx context.Context, id int) error

	// DeleteByTranscriptionID deletes all translations for a transcription segment
	DeleteByTranscriptionID(ctx context.Context, transcriptionID string) error

	// DeleteByVideoID deletes all translations for a video (via transcription segments)
	DeleteByVideoID(ctx context.Context, videoID string) error
}
