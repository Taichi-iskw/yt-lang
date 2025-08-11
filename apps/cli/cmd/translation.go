package cmd

import (
	"context"

	"github.com/Taichi-iskw/yt-lang/cmd/translation"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	translationSvc "github.com/Taichi-iskw/yt-lang/internal/service/translation"
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
	// For now, return a simple command that shows it's not fully integrated
	// In a real scenario, you would need database connection and configuration
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

	// Try to create with real services if possible
	realCmd := tryCreateWithRealServices()
	if realCmd != nil {
		return realCmd
	}

	// Fallback: create with mock services for demonstration
	return createTranslationCommandWithMocks(cmd)
}

// tryCreateWithRealServices attempts to create the command with real database connection
func tryCreateWithRealServices() *cobra.Command {
	// For now, always return nil to use mock services
	// Real database integration would be implemented here when config is available
	return nil
}

// createTranslationCommandWithMocks creates the command with mock services
func createTranslationCommandWithMocks(baseCmd *cobra.Command) *cobra.Command {
	// Create mock services for basic functionality testing
	mockTranscriptionRepo := &mockTranscriptionRepo{}
	mockTranslationRepo := &mockTranslationRepo{}
	mockCmdRunner := &translationSvc.MockCmdRunner{
		RunFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			// Simple mock that returns a basic translation
			if name == "plamo-translate" && len(args) >= 3 {
				return []byte("Mock translation result"), nil
			}
			return []byte("Mock output"), nil
		},
	}

	plamoService := translationSvc.NewPlamoService(mockCmdRunner)
	batchProcessor := translationSvc.NewBatchProcessor()

	translationService := translationSvc.NewTranslationService(
		mockTranscriptionRepo,
		mockTranslationRepo,
		plamoService,
		batchProcessor,
	)

	// Add subcommands to base command
	baseCmd.AddCommand(translation.NewCreateCommand(translationService))
	baseCmd.AddCommand(translation.NewGetCommand(translationService))
	baseCmd.AddCommand(translation.NewListCommand(translationService))
	baseCmd.AddCommand(translation.NewDeleteCommand(translationService))

	return baseCmd
}

// Mock implementations for fallback
type mockTranscriptionRepo struct{}

func (m *mockTranscriptionRepo) GetSegments(ctx context.Context, transcriptionID string) ([]*model.TranscriptionSegment, error) {
	// Return mock segments for testing
	return []*model.TranscriptionSegment{
		{ID: "seg-1", TranscriptionID: transcriptionID, Text: "Hello world"},
		{ID: "seg-2", TranscriptionID: transcriptionID, Text: "This is a test"},
	}, nil
}

func (m *mockTranscriptionRepo) Get(ctx context.Context, id string) (*model.Transcription, error) {
	return &model.Transcription{
		ID:       id,
		VideoID:  "video-1",
		Language: "en",
		Status:   "completed",
	}, nil
}

type mockTranslationRepo struct{}

func (m *mockTranslationRepo) Get(ctx context.Context, id int) (*model.Translation, error) {
	// Mock get translation
	return &model.Translation{
		ID:              id,
		TranscriptionID: 1,
		TargetLanguage:  "ja",
		Content:         "Mock translation",
		Source:          "plamo",
	}, nil
}

func (m *mockTranslationRepo) ListByTranscriptionID(ctx context.Context, transcriptionID int, limit, offset int) ([]*model.Translation, error) {
	// Mock list translations
	return []*model.Translation{
		{ID: 1, TranscriptionID: transcriptionID, TargetLanguage: "ja", Content: "Mock translation 1", Source: "plamo"},
		{ID: 2, TranscriptionID: transcriptionID, TargetLanguage: "en", Content: "Mock translation 2", Source: "plamo"},
	}, nil
}

func (m *mockTranslationRepo) Delete(ctx context.Context, id int) error {
	// Mock successful deletion
	return nil
}

func (m *mockTranslationRepo) Create(ctx context.Context, translation *model.Translation) error {
	// Mock successful creation
	translation.ID = 1
	return nil
}
