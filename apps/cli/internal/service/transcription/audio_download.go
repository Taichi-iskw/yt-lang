package transcription

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/service/common"
)

// AudioDownloadService defines operations for downloading audio from videos
type AudioDownloadService interface {
	// DownloadAudio downloads audio from a video URL using yt-dlp
	DownloadAudio(ctx context.Context, videoURL string, outputDir string) (string, error)
}

// audioDownloadService implements AudioDownloadService using yt-dlp
type audioDownloadService struct {
	cmdRunner common.CmdRunner
}

// NewAudioDownloadService creates a new AudioDownloadService with default CmdRunner
func NewAudioDownloadService() AudioDownloadService {
	return &audioDownloadService{
		cmdRunner: common.NewCmdRunner(),
	}
}

// NewAudioDownloadServiceWithCmdRunner creates a new AudioDownloadService with custom CmdRunner (for testing)
func NewAudioDownloadServiceWithCmdRunner(cmdRunner common.CmdRunner) AudioDownloadService {
	return &audioDownloadService{
		cmdRunner: cmdRunner,
	}
}

// DownloadAudio downloads audio from a video URL using yt-dlp
func (s *audioDownloadService) DownloadAudio(ctx context.Context, videoURL string, outputDir string) (string, error) {
	// Validate input
	if videoURL == "" {
		return "", errors.New(errors.CodeInvalidArg, "video URL is required")
	}
	if outputDir == "" {
		return "", errors.New(errors.CodeInvalidArg, "output directory is required")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", errors.Wrap(err, errors.CodeInternal, "failed to create output directory")
	}

	// Prepare yt-dlp command arguments for audio-only download
	args := []string{
		"-x",                     // Extract audio only
		"--audio-format", "best", // Use best available audio format
		"--audio-quality", "0", // Best quality
		"--output", filepath.Join(outputDir, "%(title)s.%(ext)s"), // Output template
		videoURL,
	}

	// Execute yt-dlp command
	_, err := s.cmdRunner.Run(ctx, "yt-dlp", args...)
	if err != nil {
		return "", errors.Wrap(err, errors.CodeExternal, s.formatYtDlpError(err, videoURL))
	}

	// Find the downloaded audio file
	// Since we can't predict the exact filename, we need to scan the output directory
	audioPath, err := s.findDownloadedAudio(outputDir)
	if err != nil {
		return "", errors.Wrap(err, errors.CodeInternal, "failed to find downloaded audio file")
	}

	return audioPath, nil
}

// findDownloadedAudio finds the most recently downloaded audio file in the output directory
func (s *audioDownloadService) findDownloadedAudio(outputDir string) (string, error) {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to read output directory: %w", err)
	}

	// Look for audio files (common extensions from yt-dlp)
	audioExtensions := []string{".m4a", ".mp3", ".webm", ".ogg", ".wav", ".opus"}
	var audioFiles []string
	var allFiles []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		allFiles = append(allFiles, name)

		for _, ext := range audioExtensions {
			if filepath.Ext(name) == ext {
				audioFiles = append(audioFiles, filepath.Join(outputDir, name))
				break
			}
		}
	}

	if len(audioFiles) == 0 {
		if len(allFiles) == 0 {
			return "", fmt.Errorf("no files found in output directory - audio download may have failed")
		}
		return "", fmt.Errorf("no audio files found. Downloaded files: %v. Supported formats: %v",
			allFiles, audioExtensions)
	}

	// Return the first audio file found
	// In a more sophisticated implementation, we could return the most recent one
	return audioFiles[0], nil
}

// formatYtDlpError provides user-friendly error messages for yt-dlp failures
func (s *audioDownloadService) formatYtDlpError(err error, videoURL string) string {
	errMsg := err.Error()

	// Check for common yt-dlp error patterns
	switch {
	case strings.Contains(errMsg, "Video unavailable"):
		return "video is not available (may be private, deleted, or region-blocked)"
	case strings.Contains(errMsg, "Private video"):
		return "video is private and cannot be downloaded"
	case strings.Contains(errMsg, "Video removed"):
		return "video has been removed by the uploader"
	case strings.Contains(errMsg, "This video is not available"):
		return "video is not available (check the video URL)"
	case strings.Contains(errMsg, "No such file or directory") && strings.Contains(errMsg, "yt-dlp"):
		return "yt-dlp is not installed or not found in PATH. Please install yt-dlp"
	case strings.Contains(errMsg, "network"):
		return "network connection error - please check your internet connection"
	case strings.Contains(errMsg, "HTTP Error 404"):
		return "video not found - please check the video ID"
	case strings.Contains(errMsg, "403"):
		return "access denied - video may be region-blocked or require login"
	case strings.Contains(errMsg, "429"):
		return "rate limited by YouTube - please try again later"
	default:
		// Extract video ID from URL for better error context
		videoID := extractVideoIDFromURL(videoURL)
		if videoID != "" {
			return fmt.Sprintf("failed to download audio from video '%s' - %s", videoID, errMsg)
		}
		return fmt.Sprintf("audio download failed - %s", errMsg)
	}
}

// extractVideoIDFromURL extracts video ID from YouTube URL
func extractVideoIDFromURL(url string) string {
	if strings.Contains(url, "watch?v=") {
		parts := strings.Split(url, "watch?v=")
		if len(parts) > 1 {
			videoID := strings.Split(parts[1], "&")[0]
			return videoID
		}
	}
	return ""
}
