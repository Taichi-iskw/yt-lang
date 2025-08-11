package transcription

import (
	"context"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// Repository defines operations for Transcription persistence (Option B: Normalized)
type Repository interface {
	// Transcription metadata operations
	Create(ctx context.Context, transcription *model.Transcription) error
	GetByID(ctx context.Context, id string) (*model.Transcription, error)
	GetByVideoID(ctx context.Context, videoID string) ([]*model.Transcription, error)
	GetByVideoIDAndLanguage(ctx context.Context, videoID, language string) (*model.Transcription, error)
	UpdateStatus(ctx context.Context, id string, status string, errorMessage *string) error
	Delete(ctx context.Context, id string) error
}

// SegmentRepository defines operations for TranscriptionSegment persistence
type SegmentRepository interface {
	// Segment operations
	CreateBatch(ctx context.Context, segments []*model.TranscriptionSegment) error
	GetByTranscriptionID(ctx context.Context, transcriptionID string) ([]*model.TranscriptionSegment, error)
	GetByTimeRange(ctx context.Context, transcriptionID string, startTime, endTime string) ([]*model.TranscriptionSegment, error)
	Delete(ctx context.Context, transcriptionID string) error
}
