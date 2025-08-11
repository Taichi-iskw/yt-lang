package translation

import (
	"github.com/Taichi-iskw/yt-lang/internal/service/translation"
	"github.com/spf13/cobra"
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
