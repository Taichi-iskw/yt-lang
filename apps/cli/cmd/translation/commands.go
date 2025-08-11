package translation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Taichi-iskw/yt-lang/internal/service/translation"
	"github.com/spf13/cobra"
)

// Command flags
var (
	targetLang string
	dryRun     bool
	format     string
)

// NewTranslationCommand creates the main translation command
func NewTranslationCommand(service translation.TranslationService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "translation",
		Short: "Manage translations",
		Long:  `Create, get, list, and delete translations for transcriptions`,
	}

	// Add subcommands
	cmd.AddCommand(NewCreateCommand(service))
	cmd.AddCommand(NewGetCommand(service))
	cmd.AddCommand(NewListCommand(service))
	cmd.AddCommand(NewDeleteCommand(service))

	return cmd
}

// NewCreateCommand creates the create translation command
func NewCreateCommand(service translation.TranslationService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [TRANSCRIPTION_ID]",
		Short: "Create a new translation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			transcriptionID := args[0]

			// Get flags
			targetLang, _ := cmd.Flags().GetString("target-lang")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			useServer, _ := cmd.Flags().GetBool("use-server")

			if dryRun {
				cmd.Println("DRY RUN: Would create translation for transcription", transcriptionID, "to", targetLang)
				return nil
			}

			ctx := context.Background()

			// Start PLaMo server if server mode is enabled
			if useServer {
				plamoService := service.GetPlamoService()
				cmd.Println("Starting PLaMo server...")
				if err := plamoService.StartServer(ctx); err != nil {
					return fmt.Errorf("failed to start PLaMo server: %w", err)
				}

				// Ensure server is stopped when command completes
				defer func() {
					cmd.Println("Stopping PLaMo server...")
					if err := plamoService.StopServer(); err != nil {
						cmd.Printf("Warning: failed to stop PLaMo server: %v\n", err)
					}
				}()
			}

			// Create translation
			translation, err := service.CreateTranslation(ctx, transcriptionID, targetLang)
			if err != nil {
				return fmt.Errorf("failed to create translation: %w", err)
			}

			cmd.Printf("Translation created successfully (ID: %d, Language: %s)\n",
				translation.ID, translation.TargetLanguage)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&targetLang, "target-lang", "ja", "Target language for translation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Perform a dry run without saving to database")
	cmd.Flags().Bool("use-server", true, "Use PLaMo server mode for better performance with multiple batches")

	return cmd
}

// NewGetCommand creates the get translation command
func NewGetCommand(service translation.TranslationService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [TRANSLATION_ID]",
		Short: "Get a translation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			translationID := args[0]

			// Get flags
			format, _ := cmd.Flags().GetString("format")

			// Get translation
			ctx := context.Background()
			translation, segments, err := service.GetTranslation(ctx, translationID)
			if err != nil {
				return fmt.Errorf("failed to get translation: %w", err)
			}

			// Format output
			switch format {
			case "json":
				output, err := json.MarshalIndent(translation, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format as JSON: %w", err)
				}
				cmd.Println(string(output))
			case "srt":
				// TODO: Implement SRT format
				cmd.Println("SRT format not yet implemented")
			default: // text
				cmd.Printf("Translation ID: %d\n", translation.ID)
				cmd.Printf("Target Language: %s\n", translation.TargetLanguage)
				cmd.Printf("Source: %s\n", translation.Source)
				cmd.Println("\nContent:")
				if segments != nil && len(segments) > 0 {
					for _, seg := range segments {
						cmd.Printf("%s -> %s\n", seg.Text, seg.TranslatedText)
					}
				} else {
					cmd.Println(translation.Content)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&format, "format", "text", "Output format (text, json, srt)")

	return cmd
}

// NewListCommand creates the list translations command
func NewListCommand(service translation.TranslationService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [TRANSCRIPTION_ID]",
		Short: "List all translations for a transcription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement list command
			cmd.Println("List command not yet implemented")
			return nil
		},
	}

	return cmd
}

// NewDeleteCommand creates the delete translation command
func NewDeleteCommand(service translation.TranslationService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [TRANSLATION_ID]",
		Short: "Delete a translation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement delete command
			cmd.Println("Delete command not yet implemented")
			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("force", false, "Force deletion without confirmation")

	return cmd
}
