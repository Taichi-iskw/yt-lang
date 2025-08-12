package cmd

import (
	"github.com/Taichi-iskw/yt-lang/cmd/translation"
	"github.com/spf13/cobra"
)

// translationCmd represents the translation command
var translationCmd *cobra.Command

func init() {
	// Initialize translation command with real services
	translationCmd = createTranslationCommand()
	rootCmd.AddCommand(translationCmd)
}

// createTranslationCommand creates the translation command with real dependencies
func createTranslationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "translation",
		Short: "Manage translations (PLaMo-powered)",
		Long:  `Create, get, list, and delete translations for transcriptions using PLaMo CLI`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Translation commands:")
			cmd.Println("  create [TRANSCRIPTION_ID]  Create a new translation")
			cmd.Println("  get [TRANSLATION_ID]       Get a translation")
			cmd.Println("  list [TRANSCRIPTION_ID]    List translations for transcription")
			cmd.Println("  delete [TRANSLATION_ID]    Delete a translation")
			cmd.Println("")
			cmd.Println("Example:")
			cmd.Println("  ytlang translation create trans-123 --target-lang ja")
			cmd.Println("  ytlang translation create trans-123 --target-lang ja --dry-run")
			return nil
		},
	}

	// Create with real services using database connection
	return createTranslationCommandWithRealServices(cmd)
}

// createTranslationCommandWithRealServices creates the command with real database services
func createTranslationCommandWithRealServices(baseCmd *cobra.Command) *cobra.Command {
	// Add subcommands with dynamic service creation (each command creates its own DB connection)
	baseCmd.AddCommand(translation.NewCreateCommand(nil)) // Pass nil, commands will create their own services
	baseCmd.AddCommand(translation.NewGetCommand(nil))
	baseCmd.AddCommand(translation.NewListCommand(nil))
	baseCmd.AddCommand(translation.NewDeleteCommand(nil))

	return baseCmd
}
