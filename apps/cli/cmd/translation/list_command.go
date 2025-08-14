package translation

import (
	"context"
	"fmt"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/service/translation"
	"github.com/spf13/cobra"
)

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

			// Use provided service if available (for testing), otherwise create real service
			var translationService translation.TranslationService
			var cleanup func()

			if service != nil {
				translationService = service
			} else {
				// Create service using factory
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				factory := NewServiceFactory()
				var err error
				translationService, cleanup, err = factory.CreateService(ctx)
				if err != nil {
					return fmt.Errorf("failed to create translation service: %w", err)
				}
				defer cleanup()
			}

			ctx := context.Background()
			translations, err := translationService.ListTranslations(ctx, transcriptionID, limit, offset)
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
				cmd.Printf("Content Preview: %s\n", truncateString(translation.TranslatedText, 100))
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
