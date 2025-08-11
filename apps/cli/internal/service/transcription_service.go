package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/transcription"
	"github.com/Taichi-iskw/yt-lang/internal/repository/video"
)

// TranscriptionService defines operations for transcription management
type TranscriptionService interface {
	// CreateTranscription creates a new transcription for a video by downloading its audio
	CreateTranscription(ctx context.Context, videoID string, language string) (*model.Transcription, error)

	// GetTranscription retrieves transcription and its segments by ID
	GetTranscription(ctx context.Context, id string) (*model.Transcription, []*model.TranscriptionSegment, error)

	// ListTranscriptions lists transcriptions for a video
	ListTranscriptions(ctx context.Context, videoID string) ([]*model.Transcription, error)

	// DeleteTranscription deletes transcription and its segments
	DeleteTranscription(ctx context.Context, id string) error
}

// transcriptionService implements TranscriptionService
type transcriptionService struct {
	transcriptionRepo transcription.Repository
	segmentRepo       transcription.SegmentRepository
	whisperService    WhisperService
	audioDownloadSvc  AudioDownloadService
	videoRepo         video.Repository
}

// NewTranscriptionService creates a new TranscriptionService with default dependencies
func NewTranscriptionService() TranscriptionService {
	return &transcriptionService{
		whisperService:   NewWhisperService(),
		audioDownloadSvc: NewAudioDownloadService(),
	}
}

// NewTranscriptionServiceWithDependencies creates a new TranscriptionService with custom dependencies (for testing)
func NewTranscriptionServiceWithDependencies(transcriptionRepo transcription.Repository, segmentRepo transcription.SegmentRepository, whisperService WhisperService) TranscriptionService {
	return &transcriptionService{
		transcriptionRepo: transcriptionRepo,
		segmentRepo:       segmentRepo,
		whisperService:    whisperService,
	}
}

// NewTranscriptionServiceWithAllDependencies creates a new TranscriptionService with all dependencies (for CLI)
func NewTranscriptionServiceWithAllDependencies(transcriptionRepo transcription.Repository, segmentRepo transcription.SegmentRepository, whisperService WhisperService, audioDownloadSvc AudioDownloadService, videoRepo video.Repository) TranscriptionService {
	return &transcriptionService{
		transcriptionRepo: transcriptionRepo,
		segmentRepo:       segmentRepo,
		whisperService:    whisperService,
		audioDownloadSvc:  audioDownloadSvc,
		videoRepo:         videoRepo,
	}
}

// CreateTranscription creates a new transcription for a video by downloading its audio
func (s *transcriptionService) CreateTranscription(ctx context.Context, videoID string, language string) (*model.Transcription, error) {
	// Get video information from database
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeNotFound, "video not found")
	}

	// Create temporary directory for audio download
	tempDir, err := os.MkdirTemp("", "yt-lang-audio-*")
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "failed to create temp directory")
	}
	defer os.RemoveAll(tempDir)

	// Download audio from video URL
	audioPath, err := s.audioDownloadSvc.DownloadAudio(ctx, video.URL, tempDir)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeExternal, "failed to download audio")
	}

	// Check if transcription already exists
	existing, err := s.transcriptionRepo.GetByVideoIDAndLanguage(ctx, videoID, language)
	if err == nil {
		return existing, nil
	}

	// Create new transcription record
	transcription := &model.Transcription{
		VideoID:   videoID,
		Language:  language,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := s.transcriptionRepo.Create(ctx, transcription); err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "failed to create transcription record")
	}

	// Perform transcription in background (for now, synchronously)
	err = s.processTranscription(ctx, transcription, audioPath)
	if err != nil {
		// Update status to failed
		errorMsg := "whisper transcription failed"
		s.transcriptionRepo.UpdateStatus(ctx, transcription.ID, "failed", &errorMsg)
		return nil, err
	}

	return transcription, nil
}

// processTranscription handles the actual transcription process
func (s *transcriptionService) processTranscription(ctx context.Context, transcription *model.Transcription, audioPath string) error {
	// Execute Whisper transcription
	result, err := s.whisperService.TranscribeAudio(ctx, audioPath, transcription.Language)
	if err != nil {
		return errors.Wrap(err, errors.CodeExternal, "whisper transcription failed")
	}

	// Convert Whisper segments to TranscriptionSegments
	segments := make([]*model.TranscriptionSegment, len(result.Segments))
	for i, seg := range result.Segments {
		segments[i] = &model.TranscriptionSegment{
			TranscriptionID: transcription.ID,
			SegmentIndex:    i,
			StartTime:       formatDuration(seg.Start),
			EndTime:         formatDuration(seg.End),
			Text:            seg.Text,
			Confidence:      &seg.Confidence,
		}
	}

	// Save segments to database
	if err := s.segmentRepo.CreateBatch(ctx, segments); err != nil {
		return errors.Wrap(err, errors.CodeInternal, "failed to save transcription segments")
	}

	// Update transcription status and metadata
	transcription.Status = "completed"
	transcription.DetectedLanguage = &result.Language
	now := time.Now()
	transcription.CompletedAt = &now

	if err := s.transcriptionRepo.UpdateStatus(ctx, transcription.ID, "completed", nil); err != nil {
		return errors.Wrap(err, errors.CodeInternal, "failed to update transcription status")
	}

	return nil
}

// GetTranscription retrieves transcription and its segments by ID
func (s *transcriptionService) GetTranscription(ctx context.Context, id string) (*model.Transcription, []*model.TranscriptionSegment, error) {
	// Get transcription
	transcription, err := s.transcriptionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.CodeNotFound, "transcription not found")
	}

	// Get segments
	segments, err := s.segmentRepo.GetByTranscriptionID(ctx, id)
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.CodeInternal, "failed to get transcription segments")
	}

	return transcription, segments, nil
}

// ListTranscriptions lists transcriptions for a video
func (s *transcriptionService) ListTranscriptions(ctx context.Context, videoID string) ([]*model.Transcription, error) {
	transcriptions, err := s.transcriptionRepo.GetByVideoID(ctx, videoID)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "failed to list transcriptions")
	}

	return transcriptions, nil
}

// DeleteTranscription deletes transcription and its segments
func (s *transcriptionService) DeleteTranscription(ctx context.Context, id string) error {
	// Delete segments first (foreign key constraint)
	if err := s.segmentRepo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, errors.CodeInternal, "failed to delete transcription segments")
	}

	// Delete transcription
	if err := s.transcriptionRepo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, errors.CodeInternal, "failed to delete transcription")
	}

	return nil
}

// formatDuration converts seconds to PostgreSQL INTERVAL format
func formatDuration(seconds float64) string {
	duration := time.Duration(seconds * float64(time.Second))
	return fmt.Sprintf("%02d:%02d:%06.3f",
		int(duration.Hours()),
		int(duration.Minutes())%60,
		float64(duration.Nanoseconds()%int64(time.Minute))/float64(time.Second))
}
