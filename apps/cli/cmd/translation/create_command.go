package translation

import (
	"context"
	"fmt"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/config"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/transcription"
	"github.com/Taichi-iskw/yt-lang/internal/repository/translation"
	"github.com/Taichi-iskw/yt-lang/internal/service/common"
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
			if service != nil {
				translationService = service
			} else {
				// Create real service with database connections
				// Create context with timeout
				ctx, cancel := context.WithTimeout(context.Background(), 360*time.Minute)
				defer cancel()

				// Load database configuration
				cfg, err := config.NewConfig()
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

				// Create database connection
				dbPool, err := config.NewDatabasePool(ctx, cfg)
				if err != nil {
					return fmt.Errorf("failed to connect to database: %w", err)
				}
				defer dbPool.Close()

				// Create repositories
				transcriptionRepo := transcription.NewRepository(dbPool)
				segmentRepo := transcription.NewSegmentRepository(dbPool)
				translationRepo := translation.NewRepository(dbPool)

				// Create services
				cmdRunner := common.NewCmdRunner()
				plamoService := translationSvc.NewPlamoServerService(cmdRunner) // Use server mode for better performance
				batchProcessor := translationSvc.NewBatchProcessor()

				// Create translation service with real repositories
				translationService = translationSvc.NewTranslationService(
					&transcriptionRepoWrapper{
						transcriptionRepo: transcriptionRepo,
						segmentRepo:       segmentRepo,
					},
					translationRepo,
					plamoService,
					batchProcessor,
				)

				// Always start PLaMo server for better performance
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

// transcriptionRepoWrapper wraps transcription and segment repositories to implement TranscriptionRepository interface
type transcriptionRepoWrapper struct {
	transcriptionRepo transcription.Repository
	segmentRepo       transcription.SegmentRepository
}

// GetSegments implements TranscriptionRepository interface
func (w *transcriptionRepoWrapper) GetSegments(ctx context.Context, transcriptionID string) ([]*model.TranscriptionSegment, error) {
	return w.segmentRepo.GetByTranscriptionID(ctx, transcriptionID)
}

// Get implements TranscriptionRepository interface
func (w *transcriptionRepoWrapper) Get(ctx context.Context, id string) (*model.Transcription, error) {
	return w.transcriptionRepo.GetByID(ctx, id)
}
