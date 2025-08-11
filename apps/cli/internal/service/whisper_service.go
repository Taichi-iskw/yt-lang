package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// WhisperService defines operations for Whisper transcription
type WhisperService interface {
	// TranscribeAudio transcribes audio file using Whisper CLI
	TranscribeAudio(ctx context.Context, audioPath string, language string) (*model.WhisperResult, error)
}

// WhisperOptions represents configuration for Whisper transcription
type WhisperOptions struct {
	Model       string  // Model size: tiny, base, small, medium, large
	Language    string  // Language code: auto, en, ja, etc.
	OutputDir   string  // Directory for output files
	Temperature float64 // Temperature for sampling
}

// whisperService implements WhisperService using Whisper CLI
type whisperService struct {
	cmdRunner CmdRunner
	model     string // default model to use
}

// NewWhisperService creates a new WhisperService with default CmdRunner
func NewWhisperService() WhisperService {
	return &whisperService{
		cmdRunner: NewCmdRunner(),
		model:     "large",
	}
}

// NewWhisperServiceWithCmdRunner creates a new WhisperService with custom CmdRunner (for testing)
func NewWhisperServiceWithCmdRunner(cmdRunner CmdRunner, model string) WhisperService {
	return &whisperService{
		cmdRunner: cmdRunner,
		model:     model,
	}
}

// TranscribeAudio transcribes audio file using Whisper CLI
func (s *whisperService) TranscribeAudio(ctx context.Context, audioPath string, language string) (*model.WhisperResult, error) {
	// Validate input
	if audioPath == "" {
		return nil, errors.New(errors.CodeInvalidArg, "audio path is required")
	}

	// Create temp directory for output
	var tempDir string
	if ctxTempDir := ctx.Value("tempDir"); ctxTempDir != nil {
		tempDir = ctxTempDir.(string)
	}

	shouldCleanup := false
	if tempDir == "" {
		var err error
		tempDir, err = os.MkdirTemp("", "yt-lang-whisper-*")
		if err != nil {
			return nil, errors.Wrap(err, errors.CodeInternal, "failed to create temp directory")
		}
		shouldCleanup = true
		defer func() {
			if shouldCleanup {
				os.RemoveAll(tempDir)
			}
		}()
	}

	// Prepare whisper command arguments
	args := []string{
		audioPath,
		"--model", s.model,
		"--language", language,
		"--output_format", "json",
		"--output_dir", tempDir,
		"--temperature", "0",
	}

	// Execute whisper command
	_, err := s.cmdRunner.Run(ctx, "whisper", args...)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeExternal, "whisper execution failed")
	}

	// Read the output JSON file
	baseName := filepath.Base(audioPath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	jsonPath := filepath.Join(tempDir, baseName+".json")

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "failed to read whisper output")
	}

	// Parse JSON into WhisperResult
	var result model.WhisperResult
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "failed to parse whisper output")
	}

	return &result, nil
}
