package transcription

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Taichi-iskw/yt-lang/internal/service/common"
	transcriptionSvc "github.com/Taichi-iskw/yt-lang/internal/service/transcription"
)

// runDryRunMode runs transcription in dry-run mode (no database save)
// This function directly uses services without repository layer
func runDryRunMode(ctx context.Context, videoID, language, format, model string) error {
	// Create services (no database needed)
	whisperService := transcriptionSvc.NewWhisperServiceWithCmdRunner(common.NewCmdRunner(), model)
	audioDownloadService := transcriptionSvc.NewAudioDownloadService()

	fmt.Printf("🎵 Testing transcription for video %s (dry-run mode)...\n", videoID)
	fmt.Printf("Language: %s\n", language)
	fmt.Printf("Format: %s\n", format)
	fmt.Printf("\n📥 Downloading audio...\n")

	// Download audio to temporary directory
	tmpDir, err := os.MkdirTemp("", "transcription-test-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	audioPath, err := audioDownloadService.DownloadAudio(ctx, videoURL, tmpDir)
	if err != nil {
		return formatTranscriptionError(err, videoID)
	}

	fmt.Printf("✅ Audio downloaded: %s\n", audioPath)
	fmt.Printf("\n🎙️ Running transcription...\n")

	// Run transcription
	whisperResult, err := whisperService.TranscribeAudio(ctx, audioPath, language)
	if err != nil {
		return formatTranscriptionError(err, videoID)
	}

	fmt.Printf("✅ Transcription completed!\n")
	fmt.Printf("Detected Language: %s\n", whisperResult.Language)
	fmt.Printf("Total segments: %d\n", len(whisperResult.Segments))
	fmt.Printf("ℹ️  Results not saved to database (dry-run mode)\n")
	fmt.Println()

	// Format and output results
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(whisperResult, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))

	case "srt":
		fmt.Print(formatWhisperResultAsSRT(whisperResult))

	case "text":
		fmt.Printf("Full Text:\n%s\n\n", whisperResult.Text)
		fmt.Printf("--- Segments (%d) ---\n", len(whisperResult.Segments))
		for _, segment := range whisperResult.Segments {
			startTime := formatSecondsToTime(segment.Start)
			endTime := formatSecondsToTime(segment.End)
			fmt.Printf("[%s -> %s] %s\n", startTime, endTime, segment.Text)
		}

	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json, srt)", format)
	}

	return nil
}

// formatSecondsToTime converts float64 seconds to HH:MM:SS.mmm format
func formatSecondsToTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	milliseconds := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, milliseconds)
}