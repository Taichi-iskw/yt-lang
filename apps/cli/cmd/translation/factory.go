package translation

import (
	"context"
	"fmt"

	"github.com/Taichi-iskw/yt-lang/internal/config"
	"github.com/Taichi-iskw/yt-lang/internal/repository/transcription"
	translationRepo "github.com/Taichi-iskw/yt-lang/internal/repository/translation"
	"github.com/Taichi-iskw/yt-lang/internal/service/common"
	"github.com/Taichi-iskw/yt-lang/internal/service/translation"
)

// ServiceFactory creates translation service instances
type ServiceFactory struct{}

// NewServiceFactory creates a new service factory
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

// CreateService creates a new translation service with all dependencies
func (f *ServiceFactory) CreateService(ctx context.Context) (translation.TranslationService, func(), error) {
	// Load database configuration
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create database connection
	dbPool, err := config.NewDatabasePool(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create repositories
	transcriptionRepository := transcription.NewRepository(dbPool)
	segmentRepo := transcription.NewSegmentRepository(dbPool)
	translationRepository := translationRepo.NewRepository(dbPool)

	// Create services
	cmdRunner := common.NewCmdRunner()
	plamoService := translation.NewPlamoServerService(cmdRunner)
	batchProcessor := translation.NewBatchProcessor()

	// Create translation service with real repositories
	translationService := translation.NewTranslationService(
		&transcriptionRepoWrapper{
			transcriptionRepo: transcriptionRepository,
			segmentRepo:       segmentRepo,
		},
		translationRepository,
		plamoService,
		batchProcessor,
	)

	// Cleanup function
	cleanup := func() {
		dbPool.Close()
	}

	return translationService, cleanup, nil
}

// CreateServiceWithPlamoServer creates a translation service and starts the PLaMo server
func (f *ServiceFactory) CreateServiceWithPlamoServer(ctx context.Context) (translation.TranslationService, func(), error) {
	service, dbCleanup, err := f.CreateService(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the PLaMo service through the interface
	plamoService := service.GetPlamoService()

	// Type assert to PlamoServerService if it supports server mode
	if serverService, ok := plamoService.(*translation.PlamoServerService); ok {
		// Start PLaMo server
		if err := serverService.StartServer(ctx); err != nil {
			dbCleanup()
			return nil, nil, fmt.Errorf("failed to start PLaMo server: %w", err)
		}

		// Combined cleanup function
		cleanup := func() {
			// Stop PLaMo server
			if err := serverService.StopServer(); err != nil {
				// Log error but don't fail cleanup
				fmt.Printf("Warning: failed to stop PLaMo server: %v\n", err)
			}
			dbCleanup()
		}
		return service, cleanup, nil
	}

	// If not a server service, just return with db cleanup
	return service, dbCleanup, nil
}
