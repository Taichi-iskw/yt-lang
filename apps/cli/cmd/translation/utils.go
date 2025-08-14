package translation

import (
	"context"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/transcription"
)

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// transcriptionRepoWrapper wraps transcription and segment repositories to implement TranscriptionRepository interface
type transcriptionRepoWrapper struct {
	transcriptionRepo transcription.Repository
	segmentRepo       transcription.SegmentRepository
}

// GetSegments implements TranscriptionRepository interface
func (w *transcriptionRepoWrapper) GetSegments(ctx context.Context, transcriptionID string) ([]*model.TranscriptionSegment, error) {
	return w.segmentRepo.GetByTranscriptionID(ctx, transcriptionID)
}

// Get implements TranscriptionRepository interface
func (w *transcriptionRepoWrapper) Get(ctx context.Context, id string) (*model.Transcription, error) {
	return w.transcriptionRepo.GetByID(ctx, id)
}
