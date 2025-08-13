package translation

import (
	"context"
	"fmt"
	"time"

	translationSvc "github.com/Taichi-iskw/yt-lang/internal/service/translation"
	"github.com/spf13/cobra"
)

// NewCreateCommand creates the create translation command
func NewCreateCommand(service translationSvc.TranslationService) *cobra.Command {
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

			// Use provided service if available (for testing), otherwise create real service
			var translationService translationSvc.TranslationService
			var cleanup func()
			
			if service != nil {
				translationService = service
			} else {
				// Create service using factory with PLaMo server support
				ctx, cancel := context.WithTimeout(context.Background(), 360*time.Minute)
				defer cancel()
				
				factory := NewServiceFactory()
				var err error
				
				// Use the version that starts PLaMo server for better performance
				cmd.Println("Starting PLaMo server...")
				translationService, cleanup, err = factory.CreateServiceWithPlamoServer(ctx)
				if err != nil {
					return fmt.Errorf("failed to create translation service: %w", err)
				}
				
				// Ensure cleanup is called when command completes
				defer func() {
					cmd.Println("Stopping PLaMo server...")
					cleanup()
				}()
			}

			// Create context with timeout for translation
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			// Create translation
			translationResult, err := translationService.CreateTranslation(ctx, transcriptionID, targetLang)
			if err != nil {
				return fmt.Errorf("failed to create translation: %w", err)
			}

			cmd.Printf("Translation created successfully (ID: %d, Language: %s)\n",
				translationResult.ID, translationResult.TargetLanguage)
			return nil
		},
	}

	// Add flags
	cmd.Flags().String("target-lang", "ja", "Target language for translation")
	cmd.Flags().Bool("dry-run", false, "Perform a dry run without saving to database")

	return cmd
}
