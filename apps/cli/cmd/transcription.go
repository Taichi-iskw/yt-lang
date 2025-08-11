package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Taichi-iskw/yt-lang/internal/config"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/transcription"
	"github.com/Taichi-iskw/yt-lang/internal/repository/video"
	"github.com/Taichi-iskw/yt-lang/internal/service"
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

		// Get language flag
		language, _ := cmd.Flags().GetString("language")

		// Create service with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
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
		videoRepo := video.NewRepository(dbPool)

		// Create services
		whisperService := service.NewWhisperService()
		audioDownloadService := service.NewAudioDownloadService()
		transcriptionService := service.NewTranscriptionServiceWithAllDependencies(
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
			return fmt.Errorf("failed to create transcription: %w", err)
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
		transcriptionService := service.NewTranscriptionServiceWithDependencies(
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
		transcriptionService := service.NewTranscriptionServiceWithDependencies(
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
		transcriptionService := service.NewTranscriptionServiceWithDependencies(
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

func init() {
	// Add language flag to create command
	transcriptionCreateCmd.Flags().String("language", "auto", "Language code for transcription (auto, en, ja, etc.)")

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
