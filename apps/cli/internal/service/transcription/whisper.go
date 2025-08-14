package transcription

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/service/common"
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
	cmdRunner common.CmdRunner
	model     string // default model to use
}

// NewWhisperService creates a new WhisperService with default CmdRunner
func NewWhisperService() WhisperService {
	return &whisperService{
		cmdRunner: common.NewCmdRunner(),
		model:     "large",
	}
}

// NewWhisperServiceWithCmdRunner creates a new WhisperService with custom CmdRunner (for testing)
func NewWhisperServiceWithCmdRunner(cmdRunner common.CmdRunner, model string) WhisperService {
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
		"--output_format", "json",
		"--output_dir", tempDir,
		"--temperature", "0",
	}

	// Add language parameter only if not auto-detection
	if language != "" && language != "auto" {
		args = append(args, "--language", language)
	}

	// Execute whisper command
	_, err := s.cmdRunner.Run(ctx, "whisper", args...)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeExternal, s.formatWhisperError(err, audioPath, language))
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

// formatWhisperError provides user-friendly error messages for Whisper failures
func (s *whisperService) formatWhisperError(err error, audioPath, language string) string {
	errMsg := err.Error()

	// Check for common Whisper error patterns
	switch {
	case strings.Contains(errMsg, "No such file or directory") && strings.Contains(errMsg, "whisper"):
		return "Whisper is not installed. Please install OpenAI Whisper: pip install openai-whisper"
	case strings.Contains(errMsg, "No module named"):
		return "Whisper dependencies missing. Please reinstall: pip install --upgrade openai-whisper"
	case strings.Contains(errMsg, "CUDA"):
		return "GPU/CUDA error detected. Whisper will fallback to CPU processing (this may be slower)"
	case strings.Contains(errMsg, "not enough memory") || strings.Contains(errMsg, "OutOfMemoryError"):
		return fmt.Sprintf("insufficient memory for model '%s'. Try using a smaller model (tiny, base, small)", s.model)
	case strings.Contains(errMsg, "Invalid language"):
		return fmt.Sprintf("unsupported language '%s'. Use language codes like 'en', 'ja', 'es' or 'auto'", language)
	case strings.Contains(errMsg, "Invalid model"):
		return fmt.Sprintf("unsupported model '%s'. Available models: tiny, base, small, medium, large", s.model)
	case strings.Contains(errMsg, "Could not load model"):
		return fmt.Sprintf("failed to load Whisper model '%s'. The model may need to be downloaded on first use", s.model)
	case strings.Contains(errMsg, "File not found") || strings.Contains(errMsg, "No such file"):
		return fmt.Sprintf("audio file not found: %s", filepath.Base(audioPath))
	case strings.Contains(errMsg, "Unsupported format") || strings.Contains(errMsg, "format not supported"):
		return fmt.Sprintf("unsupported audio format: %s", filepath.Ext(audioPath))
	case strings.Contains(errMsg, "exit status 2"):
		return fmt.Sprintf("Whisper processing failed. This may be due to corrupted audio or unsupported format (%s)", filepath.Ext(audioPath))
	default:
		return fmt.Sprintf("transcription failed with model '%s' - %s", s.model, errMsg)
	}
}
