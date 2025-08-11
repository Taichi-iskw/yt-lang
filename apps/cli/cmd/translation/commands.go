package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

			if dryRun {
				cmd.Println("DRY RUN: Would create translation for transcription", transcriptionID, "to", targetLang)
				return nil
			}

			ctx := context.Background()

			// Always start PLaMo server for better performance
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
				// Format as SRT subtitle file
				if segments != nil && len(segments) > 0 {
					for i, seg := range segments {
						// SRT format: sequence number, timing, content, blank line
						cmd.Printf("%d\n", i+1)
						// Since we don't have timing info from parsed segments,
						// use estimated timing based on segment index
						startTime := formatSRTTime(i * 3) // 3 seconds per segment estimate
						endTime := formatSRTTime((i + 1) * 3)
						cmd.Printf("%s --> %s\n", startTime, endTime)
						cmd.Printf("%s\n\n", seg.TranslatedText)
					}
				} else {
					// Fallback: split content into lines for SRT format
					lines := strings.Split(translation.Content, "\n")
					for i, line := range lines {
						line = strings.TrimSpace(line)
						if line == "" {
							continue
						}
						cmd.Printf("%d\n", i+1)
						startTime := formatSRTTime(i * 3)
						endTime := formatSRTTime((i + 1) * 3)
						cmd.Printf("%s --> %s\n", startTime, endTime)
						cmd.Printf("%s\n\n", line)
					}
				}
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
			transcriptionID := args[0]

			// Get flags
			limit, _ := cmd.Flags().GetInt("limit")
			offset, _ := cmd.Flags().GetInt("offset")

			ctx := context.Background()
			translations, err := service.ListTranslations(ctx, transcriptionID, limit, offset)
			if err != nil {
				return fmt.Errorf("failed to list translations: %w", err)
			}

			if len(translations) == 0 {
				cmd.Println("No translations found for transcription", transcriptionID)
				return nil
			}

			// Display translations
			cmd.Printf("Translations for transcription %s:\n\n", transcriptionID)
			for _, translation := range translations {
				cmd.Printf("ID: %d\n", translation.ID)
				cmd.Printf("Target Language: %s\n", translation.TargetLanguage)
				cmd.Printf("Source: %s\n", translation.Source)
				cmd.Printf("Created: %s\n", translation.CreatedAt.Format("2006-01-02 15:04:05"))
				cmd.Printf("Content Preview: %s\n", truncateString(translation.Content, 100))
				cmd.Println("---")
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().Int("limit", 10, "Maximum number of translations to list")
	cmd.Flags().Int("offset", 0, "Number of translations to skip")

	return cmd
}

// NewDeleteCommand creates the delete translation command
func NewDeleteCommand(service translation.TranslationService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [TRANSLATION_ID]",
		Short: "Delete a translation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			translationID := args[0]

			// Get flags
			force, _ := cmd.Flags().GetBool("force")

			// Confirmation prompt if not forced
			if !force {
				cmd.Printf("Are you sure you want to delete translation %s? (y/N): ", translationID)
				var response string
				fmt.Scanln(&response)

				if response != "y" && response != "Y" && response != "yes" {
					cmd.Println("Deletion cancelled")
					return nil
				}
			}

			ctx := context.Background()
			err := service.DeleteTranslation(ctx, translationID)
			if err != nil {
				return fmt.Errorf("failed to delete translation: %w", err)
			}

			cmd.Printf("Translation %s deleted successfully\n", translationID)
			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("force", false, "Force deletion without confirmation")

	return cmd
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
