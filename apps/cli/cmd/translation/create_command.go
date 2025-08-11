package translation

import (
	"context"
	"fmt"

	"github.com/Taichi-iskw/yt-lang/internal/service/translation"
	"github.com/spf13/cobra"
)

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
	cmd.Flags().String("target-lang", "ja", "Target language for translation")
	cmd.Flags().Bool("dry-run", false, "Perform a dry run without saving to database")

	return cmd
}
