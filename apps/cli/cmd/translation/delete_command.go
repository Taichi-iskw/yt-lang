package translation

import (
	"context"
	"fmt"

	"github.com/Taichi-iskw/yt-lang/internal/service/translation"
	"github.com/spf13/cobra"
)

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