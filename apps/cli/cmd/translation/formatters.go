package translation

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// Formatter defines interface for output formatting
type Formatter interface {
	Format(translation *model.Translation, segments []*TranslationSegment) (string, error)
}

// TranslationSegment represents a translated segment (temporary definition)
type TranslationSegment struct {
	Text           string
	TranslatedText string
}

// TextFormatter formats output as plain text
type TextFormatter struct{}

// Format formats translation as plain text
func (f *TextFormatter) Format(translation *model.Translation, segments []*TranslationSegment) (string, error) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Translation ID: %d\n", translation.ID))
	output.WriteString(fmt.Sprintf("Target Language: %s\n", translation.TargetLanguage))
	output.WriteString(fmt.Sprintf("Source: %s\n", translation.Source))
	output.WriteString(fmt.Sprintf("Created At: %s\n", translation.CreatedAt.Format(time.RFC3339)))
	output.WriteString("\n")

	if segments != nil && len(segments) > 0 {
		output.WriteString("Segments:\n")
		output.WriteString("=========\n")
		for i, seg := range segments {
			output.WriteString(fmt.Sprintf("[%d] %s\n    → %s\n", i+1, seg.Text, seg.TranslatedText))
		}
	} else {
		output.WriteString("TranslatedText:\n")
		output.WriteString("========\n")
		output.WriteString(translation.TranslatedText)
		output.WriteString("\n")
	}

	return output.String(), nil
}

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

// Format formats translation as JSON
func (f *JSONFormatter) Format(translation *model.Translation, segments []*TranslationSegment) (string, error) {
	type Output struct {
		Translation *model.Translation    `json:"translation"`
		Segments    []*TranslationSegment `json:"segments,omitempty"`
	}

	output := Output{
		Translation: translation,
		Segments:    segments,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// SRTFormatter formats output as SRT subtitle format
type SRTFormatter struct{}

// Format formats translation as SRT
func (f *SRTFormatter) Format(translation *model.Translation, segments []*TranslationSegment) (string, error) {
	if segments == nil || len(segments) == 0 {
		return "", fmt.Errorf("SRT format requires segments with timing information")
	}

	var output strings.Builder

	for i, seg := range segments {
		// Subtitle number
		output.WriteString(fmt.Sprintf("%d\n", i+1))

		// Timing (placeholder - would need actual timing from TranscriptionSegment)
		// Format: 00:00:00,000 --> 00:00:05,000
		startTime := formatSRTTime(i * 5) // Placeholder: 5 seconds per segment
		endTime := formatSRTTime((i + 1) * 5)
		output.WriteString(fmt.Sprintf("%s --> %s\n", startTime, endTime))

		// Translated text
		output.WriteString(seg.TranslatedText)
		output.WriteString("\n\n")
	}

	return output.String(), nil
}

// formatSRTTime formats seconds into SRT time format (00:00:00,000)
func formatSRTTime(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d,000", hours, minutes, secs)
}

// GetFormatter returns the appropriate formatter based on format string
func GetFormatter(format string) (Formatter, error) {
	switch strings.ToLower(format) {
	case "text", "txt":
		return &TextFormatter{}, nil
	case "json":
		return &JSONFormatter{}, nil
	case "srt":
		return &SRTFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
