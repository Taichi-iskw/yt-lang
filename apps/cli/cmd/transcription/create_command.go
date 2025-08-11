package transcription

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Taichi-iskw/yt-lang/internal/config"
	"github.com/Taichi-iskw/yt-lang/internal/repository/transcription"
	"github.com/Taichi-iskw/yt-lang/internal/repository/video"
	"github.com/Taichi-iskw/yt-lang/internal/service/common"
	transcriptionSvc "github.com/Taichi-iskw/yt-lang/internal/service/transcription"
)

func NewCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create [VIDEO_ID]",
		Short: "Create transcription for a video",
		Long:  `Create a transcription for a video by downloading its audio using yt-dlp and processing with Whisper.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			videoID := args[0]

			// Get flags
			language, _ := cmd.Flags().GetString("language")
			model, _ := cmd.Flags().GetString("model")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			format, _ := cmd.Flags().GetString("format")

			// Create service with timeout context
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()

			if dryRun {
				// Dry-run mode: test transcription without saving to database
				return runDryRunMode(ctx, videoID, language, format, model)
			}

			// Load database configuration
			cfg, err := config.NewConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create database connection
			dbPool, err := config.NewDatabasePool(ctx, cfg)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbPool.Close()

			// Create repositories and services
			transcriptionRepo := transcription.NewRepository(dbPool)
			segmentRepo := transcription.NewSegmentRepository(dbPool)
			videoRepo := video.NewRepository(dbPool)
			whisperService := transcriptionSvc.NewWhisperServiceWithCmdRunner(common.NewCmdRunner(), model)
			audioDownloadService := transcriptionSvc.NewAudioDownloadService()

			transcriptionService := transcriptionSvc.NewTranscriptionServiceWithAllDependencies(
				transcriptionRepo,
				segmentRepo,
				whisperService,
				audioDownloadService,
				videoRepo,
			)

			// Execute transcription
			result, err := transcriptionService.CreateTranscription(ctx, videoID, language)
			if err != nil {
				return fmt.Errorf("failed to create transcription: %w", err)
			}

			fmt.Printf("✅ Transcription created successfully!\n")
			fmt.Printf("ID: %s\n", result.ID)
			fmt.Printf("Video ID: %s\n", result.VideoID)
			fmt.Printf("Language: %s\n", result.Language)
			fmt.Printf("Status: %s\n", result.Status)
			if result.DetectedLanguage != nil {
				fmt.Printf("Detected Language: %s\n", *result.DetectedLanguage)
			}
			fmt.Printf("Created: %s\n", result.CreatedAt.Format(time.RFC3339))

			return nil
		},
	}

	// Add flags
	createCmd.Flags().StringP("language", "l", "auto", "Language for transcription (e.g., 'en', 'ja', 'auto')")
	createCmd.Flags().StringP("model", "m", "base", "Whisper model to use (tiny, base, small, medium, large)")
	createCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode - test transcription without saving to database")
	createCmd.Flags().StringP("format", "f", "text", "Output format (text, json, srt)")

	return createCmd
}
