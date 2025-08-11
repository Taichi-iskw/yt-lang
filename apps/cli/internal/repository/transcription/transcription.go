package transcription

import (
	"context"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// TranscriptionRepository defines operations for Transcription persistence
type TranscriptionRepository interface {
	// Create creates a new transcription segment
	Create(ctx context.Context, transcription *model.Transcription) error

	// CreateBatch creates multiple transcription segments for a video
	CreateBatch(ctx context.Context, transcriptions []*model.Transcription) error

	// GetByVideoID retrieves all transcription segments for a video, ordered by start_time
	GetByVideoID(ctx context.Context, videoID string) ([]*model.Transcription, error)

	// GetByVideoIDAndLanguage retrieves transcription segments for a video in specific language
	GetByVideoIDAndLanguage(ctx context.Context, videoID, language string) ([]*model.Transcription, error)

	// GetByTimeRange retrieves transcription segments within a time range
	GetByTimeRange(ctx context.Context, videoID string, startTime, endTime float64) ([]*model.Transcription, error)

	// Delete deletes a transcription segment by ID
	Delete(ctx context.Context, id int) error

	// DeleteByVideoID deletes all transcription segments for a video
	DeleteByVideoID(ctx context.Context, videoID string) error
}
