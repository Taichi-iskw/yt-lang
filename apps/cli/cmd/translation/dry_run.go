package translation

import (
	"context"
	"fmt"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/service/translation"
)

// DryRunService wraps the translation service for dry-run operations
type DryRunService struct {
	transcriptionRepo translation.TranscriptionRepository
	batchProcessor    translation.BatchProcessor
}

// NewDryRunService creates a new dry-run service
func NewDryRunService(
	transcriptionRepo translation.TranscriptionRepository,
	batchProcessor translation.BatchProcessor,
) *DryRunService {
	return &DryRunService{
		transcriptionRepo: transcriptionRepo,
		batchProcessor:    batchProcessor,
	}
}

// SimulateTranslation simulates a translation without actually executing it
func (s *DryRunService) SimulateTranslation(ctx context.Context, transcriptionID string, targetLang string) (*DryRunResult, error) {
	// Get transcription segments
	segments, err := s.transcriptionRepo.GetSegments(ctx, transcriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get segments: %w", err)
	}
	
	if len(segments) == 0 {
		return nil, fmt.Errorf("no segments found for transcription %s", transcriptionID)
	}
	
	// Create batches
	batches, err := s.batchProcessor.CreateBatches(segments, 7000)
	if err != nil {
		return nil, fmt.Errorf("failed to create batches: %w", err)
	}
	
	// Calculate statistics
	totalSegments := len(segments)
	totalBatches := len(batches)
	totalChars := 0
	for _, seg := range segments {
		totalChars += len(seg.Text)
	}
	
	// Estimate tokens
	estimatedTokens := totalChars / 4 // Rough estimate for English
	
	return &DryRunResult{
		TranscriptionID:  transcriptionID,
		TargetLanguage:   targetLang,
		TotalSegments:    totalSegments,
		TotalBatches:     totalBatches,
		TotalCharacters:  totalChars,
		EstimatedTokens:  estimatedTokens,
		EstimatedAPICall: totalBatches,
		Segments:         segments,
		Batches:          batches,
	}, nil
}

// DryRunResult contains the dry-run analysis results
type DryRunResult struct {
	TranscriptionID  string
	TargetLanguage   string
	TotalSegments    int
	TotalBatches     int
	TotalCharacters  int
	EstimatedTokens  int
	EstimatedAPICall int
	Segments         []*model.TranscriptionSegment
	Batches          []translation.SegmentBatch
}

// FormatDryRunResult formats the dry-run result for display
func FormatDryRunResult(result *DryRunResult) string {
	output := fmt.Sprintf(`DRY RUN ANALYSIS
================
Transcription ID: %s
Target Language: %s

Statistics:
- Total Segments: %d
- Total Characters: %d
- Estimated Tokens: %d
- Batches to Process: %d
- Estimated API Calls: %d

Batch Details:
`, result.TranscriptionID, result.TargetLanguage,
		result.TotalSegments, result.TotalCharacters,
		result.EstimatedTokens, result.TotalBatches,
		result.EstimatedAPICall)
	
	for i, batch := range result.Batches {
		output += fmt.Sprintf("  Batch %d: %d segments, separator: %s\n",
			i+1, len(batch.Segments), batch.Separator)
	}
	
	output += "\nThis is a dry run - no actual translation will be performed."
	
	return output
}