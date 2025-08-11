package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Taichi-iskw/yt-lang/internal/config"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/transcription"
	"github.com/Taichi-iskw/yt-lang/internal/repository/video"
	"github.com/Taichi-iskw/yt-lang/internal/service/common"
	transcriptionSvc "github.com/Taichi-iskw/yt-lang/internal/service/transcription"
)

// transcriptionCmd represents the transcription command
var transcriptionCmd = &cobra.Command{
	Use:   "transcription",
	Short: "Transcription operations for videos",
	Long:  `Operations for creating and managing transcriptions of video audio.`,
}

// transcriptionCreateCmd creates a transcription for a video
var transcriptionCreateCmd = &cobra.Command{
	Use:   "create [VIDEO_ID]",
	Short: "Create transcription for a video",
	Long:  `Create a transcription for a video by downloading its audio using yt-dlp and processing with Whisper.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		videoID := args[0]

		// Get flags
		language, _ := cmd.Flags().GetString("language")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		format, _ := cmd.Flags().GetString("format")
		model, _ := cmd.Flags().GetString("model")

		// Create service with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if dryRun {
			// Dry-run mode: test transcription without saving to database
			return runDryRunMode(ctx, videoID, language, format, model)
		}

		// Normal mode: full transcription with database save
		// Load configuration
		cfg, err := config.NewConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create database connection
		dbPool, err := config.NewDatabasePool(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer dbPool.Close()

		// Create repositories
		transcriptionRepo := transcription.NewRepository(dbPool)
		segmentRepo := transcription.NewSegmentRepository(dbPool)
		videoRepo := video.NewRepository(dbPool)

		// Create services
		whisperService := transcriptionSvc.NewWhisperServiceWithCmdRunner(common.NewCmdRunner(), model)
		audioDownloadService := transcriptionSvc.NewAudioDownloadService()
		transcriptionService := transcriptionSvc.NewTranscriptionServiceWithAllDependencies(
			transcriptionRepo,
			segmentRepo,
			whisperService,
			audioDownloadService,
			videoRepo,
		)

		fmt.Printf("Creating transcription for video %s...\n", videoID)
		fmt.Printf("Language: %s\n", language)
		fmt.Printf("Downloading audio...\n")

		// Create transcription
		result, err := transcriptionService.CreateTranscription(ctx, videoID, language)
		if err != nil {
			return formatTranscriptionError(err, videoID)
		}

		fmt.Printf("✅ Transcription created successfully!\n")
		fmt.Printf("ID: %s\n", result.ID)
		fmt.Printf("Status: %s\n", result.Status)
		if result.DetectedLanguage != nil {
			fmt.Printf("Detected Language: %s\n", *result.DetectedLanguage)
		}

		return nil
	},
}

// transcriptionGetCmd retrieves a transcription by ID
var transcriptionGetCmd = &cobra.Command{
	Use:   "get [TRANSCRIPTION_ID]",
	Short: "Get transcription by ID",
	Long:  `Retrieve transcription and its segments by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		transcriptionID := args[0]

		// Create service with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Load configuration
		cfg, err := config.NewConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create database connection
		dbPool, err := config.NewDatabasePool(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer dbPool.Close()

		// Create repositories
		transcriptionRepo := transcription.NewRepository(dbPool)
		segmentRepo := transcription.NewSegmentRepository(dbPool)

		// Create transcription service
		transcriptionService := transcriptionSvc.NewTranscriptionServiceWithDependencies(
			transcriptionRepo,
			segmentRepo,
			nil, // whisperService not needed for get operation
		)

		// Get transcription
		transcriptionResult, segments, err := transcriptionService.GetTranscription(ctx, transcriptionID)
		if err != nil {
			return fmt.Errorf("failed to get transcription: %w", err)
		}

		// Check format flag
		format, _ := cmd.Flags().GetString("format")

		switch format {
		case "json":
			// Output as JSON
			result := map[string]interface{}{
				"transcription": transcriptionResult,
				"segments":      segments,
			}
			jsonData, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(jsonData))

		case "srt":
			// Output as SRT format
			fmt.Print(formatAsSRT(segments))

		default:
			// Default text format
			fmt.Printf("Transcription ID: %s\n", transcriptionResult.ID)
			fmt.Printf("Video ID: %s\n", transcriptionResult.VideoID)
			fmt.Printf("Language: %s\n", transcriptionResult.Language)
			fmt.Printf("Status: %s\n", transcriptionResult.Status)
			if transcriptionResult.DetectedLanguage != nil {
				fmt.Printf("Detected Language: %s\n", *transcriptionResult.DetectedLanguage)
			}
			fmt.Printf("Created: %s\n", transcriptionResult.CreatedAt.Format(time.RFC3339))
			if transcriptionResult.CompletedAt != nil {
				fmt.Printf("Completed: %s\n", transcriptionResult.CompletedAt.Format(time.RFC3339))
			}

			fmt.Printf("\n--- Segments (%d) ---\n", len(segments))
			for _, segment := range segments {
				fmt.Printf("[%s -> %s] %s\n", segment.StartTime, segment.EndTime, segment.Text)
			}
		}

		return nil
	},
}

// transcriptionListCmd lists transcriptions for a video
var transcriptionListCmd = &cobra.Command{
	Use:   "list [VIDEO_ID]",
	Short: "List transcriptions for a video",
	Long:  `List all transcriptions for a specific video.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		videoID := args[0]

		// Create service with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Load configuration
		cfg, err := config.NewConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create database connection
		dbPool, err := config.NewDatabasePool(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer dbPool.Close()

		// Create repositories
		transcriptionRepo := transcription.NewRepository(dbPool)
		segmentRepo := transcription.NewSegmentRepository(dbPool)

		// Create transcription service
		transcriptionService := transcriptionSvc.NewTranscriptionServiceWithDependencies(
			transcriptionRepo,
			segmentRepo,
			nil, // whisperService not needed for list operation
		)

		// List transcriptions
		transcriptions, err := transcriptionService.ListTranscriptions(ctx, videoID)
		if err != nil {
			return fmt.Errorf("failed to list transcriptions: %w", err)
		}

		// Check if no transcriptions found
		if len(transcriptions) == 0 {
			fmt.Printf("No transcriptions found for video ID: %s\n", videoID)
			return nil
		}

		// Display results
		fmt.Printf("Found %d transcription(s) for video %s:\n\n", len(transcriptions), videoID)
		for _, t := range transcriptions {
			fmt.Printf("ID: %s\n", t.ID)
			fmt.Printf("Language: %s\n", t.Language)
			fmt.Printf("Status: %s\n", t.Status)
			if t.DetectedLanguage != nil {
				fmt.Printf("Detected Language: %s\n", *t.DetectedLanguage)
			}
			fmt.Printf("Created: %s\n", t.CreatedAt.Format(time.RFC3339))
			if t.CompletedAt != nil {
				fmt.Printf("Completed: %s\n", t.CompletedAt.Format(time.RFC3339))
			}
			if t.ErrorMessage != nil {
				fmt.Printf("Error: %s\n", *t.ErrorMessage)
			}
			fmt.Println("---")
		}

		return nil
	},
}

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

// transcriptionDeleteCmd deletes a transcription
var transcriptionDeleteCmd = &cobra.Command{
	Use:   "delete [TRANSCRIPTION_ID]",
	Short: "Delete transcription by ID",
	Long:  `Delete a transcription and all its segments.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		transcriptionID := args[0]

		// Create service with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Load configuration
		cfg, err := config.NewConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create database connection
		dbPool, err := config.NewDatabasePool(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer dbPool.Close()

		// Create repositories
		transcriptionRepo := transcription.NewRepository(dbPool)
		segmentRepo := transcription.NewSegmentRepository(dbPool)

		// Create transcription service
		transcriptionService := transcriptionSvc.NewTranscriptionServiceWithDependencies(
			transcriptionRepo,
			segmentRepo,
			nil, // whisperService not needed for delete operation
		)

		// Confirm deletion
		confirm, _ := cmd.Flags().GetBool("confirm")
		if !confirm {
			fmt.Printf("Are you sure you want to delete transcription %s? Use --confirm flag to proceed.\n", transcriptionID)
			return nil
		}

		// Delete transcription
		err = transcriptionService.DeleteTranscription(ctx, transcriptionID)
		if err != nil {
			return fmt.Errorf("failed to delete transcription: %w", err)
		}

		fmt.Printf("✅ Transcription %s deleted successfully!\n", transcriptionID)
		return nil
	},
}

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

// formatSecondsToTime converts seconds (float64) to HH:MM:SS format
func formatSecondsToTime(seconds float64) string {
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	secs := totalSeconds % 60
	milliseconds := int((seconds - float64(totalSeconds)) * 1000)

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, milliseconds)
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

func init() {
	// Add flags to create command
	transcriptionCreateCmd.Flags().String("language", "auto", "Language code for transcription (auto, en, ja, etc.)")
	transcriptionCreateCmd.Flags().Bool("dry-run", false, "Test transcription without saving to database")
	transcriptionCreateCmd.Flags().String("format", "text", "Output format for dry-run: text, json, srt")
	transcriptionCreateCmd.Flags().String("model", "base", "Whisper model to use (tiny, base, small, medium, large)")

	// Add format flag to get command
	transcriptionGetCmd.Flags().String("format", "text", "Output format: text, json, srt")

	// Add confirm flag to delete command
	transcriptionDeleteCmd.Flags().Bool("confirm", false, "Confirm deletion without prompt")

	transcriptionCmd.AddCommand(transcriptionCreateCmd)
	transcriptionCmd.AddCommand(transcriptionGetCmd)
	transcriptionCmd.AddCommand(transcriptionListCmd)
	transcriptionCmd.AddCommand(transcriptionDeleteCmd)
	rootCmd.AddCommand(transcriptionCmd)
}
