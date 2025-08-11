package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Taichi-iskw/yt-lang/internal/errors"
)

// AudioDownloadService defines operations for downloading audio from videos
type AudioDownloadService interface {
	// DownloadAudio downloads audio from a video URL using yt-dlp
	DownloadAudio(ctx context.Context, videoURL string, outputDir string) (string, error)
}

// audioDownloadService implements AudioDownloadService using yt-dlp
type audioDownloadService struct {
	cmdRunner CmdRunner
}

// NewAudioDownloadService creates a new AudioDownloadService with default CmdRunner
func NewAudioDownloadService() AudioDownloadService {
	return &audioDownloadService{
		cmdRunner: NewCmdRunner(),
	}
}

// NewAudioDownloadServiceWithCmdRunner creates a new AudioDownloadService with custom CmdRunner (for testing)
func NewAudioDownloadServiceWithCmdRunner(cmdRunner CmdRunner) AudioDownloadService {
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
		return "", errors.Wrap(err, errors.CodeExternal, "yt-dlp audio download failed")
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
	audioExtensions := []string{".m4a", ".mp3", ".webm", ".ogg", ".wav"}
	var audioFiles []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, ext := range audioExtensions {
			if filepath.Ext(name) == ext {
				audioFiles = append(audioFiles, filepath.Join(outputDir, name))
				break
			}
		}
	}

	if len(audioFiles) == 0 {
		return "", fmt.Errorf("no audio files found in output directory")
	}

	// Return the first audio file found
	// In a more sophisticated implementation, we could return the most recent one
	return audioFiles[0], nil
}
