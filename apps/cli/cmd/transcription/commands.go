package transcription

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Taichi-iskw/yt-lang/internal/config"
	"github.com/Taichi-iskw/yt-lang/internal/repository/transcription"
	transcriptionSvc "github.com/Taichi-iskw/yt-lang/internal/service/transcription"
)

// NewTranscriptionCmd creates and returns the transcription command
func NewTranscriptionCmd() *cobra.Command {
	// transcriptionCmd represents the transcription command
	transcriptionCmd := &cobra.Command{
		Use:   "transcription",
		Short: "Transcription operations for videos",
		Long:  `Operations for creating and managing transcriptions of video audio.`,
	}

	// Add subcommands
	transcriptionCmd.AddCommand(NewCreateCmd())
	transcriptionCmd.AddCommand(NewGetCmd())
	transcriptionCmd.AddCommand(NewListCmd())
	transcriptionCmd.AddCommand(NewDeleteCmd())

	return transcriptionCmd
}

func NewGetCmd() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get [TRANSCRIPTION_ID]",
		Short: "Get transcription by ID",
		Long:  `Retrieve and display a transcription with its segments by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			transcriptionID := args[0]

			// Get flags
			format, _ := cmd.Flags().GetString("format")

			// Create context
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Load database configuration
			cfg, err := config.NewConfig()
			if err != nil {
				return err
			}

			// Create database connection
			dbPool, err := config.NewDatabasePool(ctx, cfg)
			if err != nil {
				return err
			}
			defer dbPool.Close()

			// Create repositories and service
			transcriptionRepo := transcription.NewRepository(dbPool)
			segmentRepo := transcription.NewSegmentRepository(dbPool)

			transcriptionService := transcriptionSvc.NewTranscriptionServiceWithDependencies(
				transcriptionRepo,
				segmentRepo,
				nil, // WhisperService not needed for retrieval
			)

			// Retrieve transcription
			result, segments, err := transcriptionService.GetTranscription(ctx, transcriptionID)
			if err != nil {
				return err
			}

			// Display results based on format
			switch format {
			case "json":
				output := map[string]interface{}{
					"transcription": result,
					"segments":      segments,
				}
				jsonBytes, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(jsonBytes))

			case "srt":
				srtContent := formatAsSRT(segments)
				fmt.Print(srtContent)

			default: // text
				fmt.Printf("Transcription ID: %s\n", result.ID)
				fmt.Printf("Video ID: %s\n", result.VideoID)
				fmt.Printf("Language: %s\n", result.Language)
				fmt.Printf("Status: %s\n", result.Status)
				if result.DetectedLanguage != nil {
					fmt.Printf("Detected Language: %s\n", *result.DetectedLanguage)
				}
				fmt.Printf("Created: %s\n", result.CreatedAt.Format(time.RFC3339))
				if result.CompletedAt != nil {
					fmt.Printf("Completed: %s\n", result.CompletedAt.Format(time.RFC3339))
				}
				fmt.Printf("\nSegments (%d):\n", len(segments))

				for _, segment := range segments {
					fmt.Printf("[%s - %s] %s\n", segment.StartTime, segment.EndTime, segment.Text)
				}
			}

			return nil
		},
	}

	// Add flags
	getCmd.Flags().StringP("format", "f", "text", "Output format: text, json, srt")

	return getCmd
}

func NewListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list [VIDEO_ID]",
		Short: "List transcriptions for a video",
		Long:  `List all transcriptions for a specific video.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			videoID := args[0]

			// Create context
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Load database configuration
			cfg, err := config.NewConfig()
			if err != nil {
				return err
			}

			// Create database connection
			dbPool, err := config.NewDatabasePool(ctx, cfg)
			if err != nil {
				return err
			}
			defer dbPool.Close()

			// Create repositories and service
			transcriptionRepo := transcription.NewRepository(dbPool)
			segmentRepo := transcription.NewSegmentRepository(dbPool)

			transcriptionService := transcriptionSvc.NewTranscriptionServiceWithDependencies(
				transcriptionRepo,
				segmentRepo,
				nil, // WhisperService not needed for listing
			)

			// List transcriptions
			results, err := transcriptionService.ListTranscriptions(ctx, videoID)
			if err != nil {
				return err
			}

			// Display results
			if len(results) == 0 {
				fmt.Printf("No transcriptions found for video: %s\n", videoID)
				return nil
			}

			fmt.Printf("Transcriptions for video %s (%d found):\n\n", videoID, len(results))
			for _, t := range results {
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
				fmt.Println("---")
			}

			return nil
		},
	}

	return listCmd
}

func NewDeleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete [TRANSCRIPTION_ID]",
		Short: "Delete transcription by ID",
		Long:  `Delete a transcription and all its segments by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			transcriptionID := args[0]

			// Get confirmation flag
			force, _ := cmd.Flags().GetBool("force")

			if !force {
				fmt.Printf("Are you sure you want to delete transcription '%s'? [y/N]: ", transcriptionID)
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Deletion cancelled.")
					return nil
				}
			}

			// Create context
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Load database configuration
			cfg, err := config.NewConfig()
			if err != nil {
				return err
			}

			// Create database connection
			dbPool, err := config.NewDatabasePool(ctx, cfg)
			if err != nil {
				return err
			}
			defer dbPool.Close()

			// Create repositories and service
			transcriptionRepo := transcription.NewRepository(dbPool)
			segmentRepo := transcription.NewSegmentRepository(dbPool)

			transcriptionService := transcriptionSvc.NewTranscriptionServiceWithDependencies(
				transcriptionRepo,
				segmentRepo,
				nil, // WhisperService not needed for deletion
			)

			// Delete transcription
			err = transcriptionService.DeleteTranscription(ctx, transcriptionID)
			if err != nil {
				return err
			}

			fmt.Printf("✅ Transcription '%s' deleted successfully.\n", transcriptionID)
			return nil
		},
	}

	// Add flags
	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	return deleteCmd
}

