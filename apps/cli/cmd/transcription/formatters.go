package transcription

import (
	"fmt"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// formatAsSRT formats transcription segments as SRT subtitle format
func formatAsSRT(segments []*model.TranscriptionSegment) string {
	var result strings.Builder

	for i, segment := range segments {
		// SRT format: sequence number, timestamp, text, blank line
		result.WriteString(fmt.Sprintf("%d\n", i+1))
		result.WriteString(fmt.Sprintf("%s --> %s\n",
			formatTimeForSRT(segment.StartTime),
			formatTimeForSRT(segment.EndTime)))
		result.WriteString(fmt.Sprintf("%s\n\n", segment.Text))
	}

	return result.String()
}

// formatTimeForSRT converts PostgreSQL interval format to SRT timestamp format
func formatTimeForSRT(intervalTime string) string {
	// Convert "HH:MM:SS.sss" to "HH:MM:SS,sss" (SRT uses comma for milliseconds)
	return strings.Replace(intervalTime, ".", ",", 1)
}

// formatWhisperResultAsSRT formats WhisperResult as SRT subtitle format
func formatWhisperResultAsSRT(result *model.WhisperResult) string {
	var output strings.Builder

	for i, segment := range result.Segments {
		// SRT format: sequence number, timestamp, text, blank line
		output.WriteString(fmt.Sprintf("%d\n", i+1))
		output.WriteString(fmt.Sprintf("%s --> %s\n",
			formatSecondsToSRTTime(segment.Start),
			formatSecondsToSRTTime(segment.End)))
		output.WriteString(fmt.Sprintf("%s\n\n", segment.Text))
	}

	return output.String()
}

// formatSecondsToSRTTime converts seconds (float64) to SRT timestamp format
func formatSecondsToSRTTime(seconds float64) string {
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	secs := totalSeconds % 60
	milliseconds := int((seconds - float64(totalSeconds)) * 1000)

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, milliseconds)
}

// formatTranscriptionError provides user-friendly error messages for transcription failures
func formatTranscriptionError(err error, videoID string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check for specific error patterns and provide helpful messages
	switch {
	case strings.Contains(errMsg, "video is not available"):
		return fmt.Errorf("❌ Video '%s' is not available. Please check:\n   • Video ID is correct\n   • Video is not private or deleted\n   • You have internet connection", videoID)
	case strings.Contains(errMsg, "yt-dlp is not installed"):
		return fmt.Errorf("❌ yt-dlp is required but not installed.\n   • Install: pip install yt-dlp\n   • Or visit: https://github.com/yt-dlp/yt-dlp")
	case strings.Contains(errMsg, "Whisper is not installed"):
		return fmt.Errorf("❌ Whisper is required but not installed.\n   • Install: pip install openai-whisper\n   • Or visit: https://github.com/openai/whisper")
	case strings.Contains(errMsg, "insufficient memory"):
		return fmt.Errorf("❌ Not enough memory for transcription.\n   • Try using a smaller model: --model tiny or --model base\n   • Close other applications to free memory")
	case strings.Contains(errMsg, "network connection error"):
		return fmt.Errorf("❌ Network connection failed.\n   • Check your internet connection\n   • Verify firewall/proxy settings")
	case strings.Contains(errMsg, "rate limited"):
		return fmt.Errorf("❌ YouTube rate limit reached.\n   • Wait a few minutes and try again\n   • Consider using a VPN if the issue persists")
	case strings.Contains(errMsg, "unsupported model"):
		return fmt.Errorf("❌ Invalid Whisper model specified.\n   • Available models: tiny, base, small, medium, large\n   • Recommended: base (balanced) or large (high quality)")
	case strings.Contains(errMsg, "unsupported language"):
		return fmt.Errorf("❌ Invalid language code specified.\n   • Use language codes like: en, ja, es, fr, de\n   • Or use 'auto' for automatic detection")
	default:
		return fmt.Errorf("❌ Transcription failed for video '%s':\n   %s", videoID, errMsg)
	}
}