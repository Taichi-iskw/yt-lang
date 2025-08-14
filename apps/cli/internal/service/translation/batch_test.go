package translation

import (
	"testing"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchProcessor_CreateBatches(t *testing.T) {
	tests := []struct {
		name            string
		segments        []*model.TranscriptionSegment
		maxTokens       int
		expectedBatches int
		wantErr         bool
	}{
		{
			name: "single small batch",
			segments: []*model.TranscriptionSegment{
				{Text: "Hello"},
				{Text: "World"},
			},
			maxTokens:       1000,
			expectedBatches: 1,
			wantErr:         false,
		},
		{
			name: "multiple batches needed",
			segments: []*model.TranscriptionSegment{
				{Text: "This is a very long text that should exceed token limits when combined"},
				{Text: "Another long text segment that also takes many tokens"},
				{Text: "Short"},
			},
			maxTokens:       10, // Very small limit to force multiple batches
			expectedBatches: 3,
			wantErr:         false,
		},
		{
			name:            "empty segments",
			segments:        []*model.TranscriptionSegment{},
			maxTokens:       1000,
			expectedBatches: 0,
			wantErr:         false,
		},
		{
			name: "invalid max tokens",
			segments: []*model.TranscriptionSegment{
				{Text: "Hello"},
			},
			maxTokens: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewBatchProcessor()

			batches, err := processor.CreateBatches(tt.segments, tt.maxTokens)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, batches, tt.expectedBatches)

			// Verify all segments are included
			totalSegments := 0
			for _, batch := range batches {
				totalSegments += len(batch.Segments)
			}
			assert.Equal(t, len(tt.segments), totalSegments)
		})
	}
}

func TestBatchProcessor_SplitTranslation(t *testing.T) {
	tests := []struct {
		name        string
		batch       SegmentBatch
		translation string
		wantErr     bool
	}{
		{
			name: "successful split with underscore separator",
			batch: SegmentBatch{
				Segments: []*model.TranscriptionSegment{
					{ID: "1", Text: "Hello"},
					{ID: "2", Text: "World"},
				},
				Separator:    "__",
				CombinedText: "Hello__World",
			},
			translation: "こんにちは__世界",
			wantErr:     false,
		},
		{
			name: "separator count mismatch",
			batch: SegmentBatch{
				Segments: []*model.TranscriptionSegment{
					{ID: "1", Text: "Hello"},
					{ID: "2", Text: "World"},
				},
				Separator:    "__",
				CombinedText: "Hello__World",
			},
			translation: "こんにちは世界", // Missing separator
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewBatchProcessor()

			// Use the internal method for testing
			bp := processor.(*batchProcessor)
			segments, err := bp.splitAndValidateTranslation(tt.batch, tt.translation)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, segments, len(tt.batch.Segments))
		})
	}
}
