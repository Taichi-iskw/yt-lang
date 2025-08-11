//go:build integration

package service

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/channel"
	"github.com/Taichi-iskw/yt-lang/internal/repository/transcription"
	"github.com/Taichi-iskw/yt-lang/internal/repository/video"
)

// mockAudioDownloadServiceIntegration for integration testing
type mockAudioDownloadServiceIntegration struct {
	audioFilePath string
}

func (m *mockAudioDownloadServiceIntegration) DownloadAudio(ctx context.Context, videoURL string, outputDir string) (string, error) {
	// Return pre-created audio file path for testing
	return m.audioFilePath, nil
}

// mockWhisperServiceIntegration for integration testing
type mockWhisperServiceIntegration struct {
	whisperResult *model.WhisperResult
}

func (m *mockWhisperServiceIntegration) TranscribeAudio(ctx context.Context, audioPath string, language string) (*model.WhisperResult, error) {
	// Return mock whisper result for testing
	return m.whisperResult, nil
}

func TestTranscriptionService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start PostgreSQL container
	ctx := context.Background()
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("ytlang_test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate container: %v", err)
		}
	}()

	// Get connection details
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Create database connection
	dbPool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	defer dbPool.Close()

	// Run migrations
	err = runMigrations(ctx, dbPool)
	require.NoError(t, err)

	// Create test audio file
	tempDir, err := os.MkdirTemp("", "transcription-integration-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	audioFilePath := filepath.Join(tempDir, "test-audio.wav")
	err = os.WriteFile(audioFilePath, []byte("fake audio content"), 0644)
	require.NoError(t, err)

	// Create repositories
	channelRepo := channel.NewRepository(dbPool)
	videoRepo := video.NewRepository(dbPool)
	transcriptionRepo := transcription.NewRepository(dbPool)
	segmentRepo := transcription.NewSegmentRepository(dbPool)

	// Create test channel
	testChannel := &model.Channel{
		ID:   "test-channel-123",
		Name: "Test Channel",
		URL:  "https://youtube.com/@testchannel",
	}
	err = channelRepo.Create(ctx, testChannel)
	require.NoError(t, err)

	// Create test video
	testVideo := &model.Video{
		ID:        "test-video-123",
		Title:     "Test Video",
		URL:       "https://youtube.com/watch?v=test123",
		ChannelID: "test-channel-123",
		Duration:  120,
	}
	err = videoRepo.Create(ctx, testVideo)
	require.NoError(t, err)

	// Create mock services
	mockAudioSvc := &mockAudioDownloadServiceIntegration{
		audioFilePath: audioFilePath,
	}

	mockWhisperSvc := &mockWhisperServiceIntegration{
		whisperResult: &model.WhisperResult{
			Text:     "This is a test transcription.",
			Language: "en",
			Segments: []model.WhisperSegment{
				{
					ID:         0,
					Start:      0.0,
					End:        3.5,
					Text:       "This is a test transcription.",
					Confidence: -0.3,
				},
			},
		},
	}

	// Create transcription service
	transcriptionService := NewTranscriptionServiceWithAllDependencies(
		transcriptionRepo,
		segmentRepo,
		mockWhisperSvc,
		mockAudioSvc,
		videoRepo,
	)

	t.Run("CreateTranscription_Success", func(t *testing.T) {
		// Test transcription creation
		result, err := transcriptionService.CreateTranscription(ctx, "test-video-123", "auto")
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify transcription was saved
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, "test-video-123", result.VideoID)
		assert.Equal(t, "auto", result.Language)
		assert.Equal(t, "completed", result.Status)
		assert.Equal(t, "en", *result.DetectedLanguage)

		// Verify segments were saved
		segments, err := segmentRepo.GetByTranscriptionID(ctx, result.ID)
		require.NoError(t, err)
		assert.Len(t, segments, 1)
		assert.Equal(t, "This is a test transcription.", segments[0].Text)
		assert.Equal(t, "00:00:00", segments[0].StartTime)
		assert.Equal(t, "00:00:03.5", segments[0].EndTime)
	})

	t.Run("CreateTranscription_AlreadyExists", func(t *testing.T) {
		// Create another video for this test
		testVideo2 := &model.Video{
			ID:        "test-video-456",
			Title:     "Test Video 2",
			URL:       "https://youtube.com/watch?v=test456",
			ChannelID: "test-channel-123",
			Duration:  90,
		}
		err = videoRepo.Create(ctx, testVideo2)
		require.NoError(t, err)

		// Create first transcription
		result1, err := transcriptionService.CreateTranscription(ctx, "test-video-456", "ja")
		require.NoError(t, err)

		// Try to create another transcription for same video/language
		result2, err := transcriptionService.CreateTranscription(ctx, "test-video-456", "ja")
		require.NoError(t, err)

		// Should return the existing transcription
		assert.Equal(t, result1.ID, result2.ID)
	})

	t.Run("GetTranscription_Success", func(t *testing.T) {
		// Create transcription first
		created, err := transcriptionService.CreateTranscription(ctx, "test-video-123", "en")
		require.NoError(t, err)

		// Get transcription
		transcriptionResult, segments, err := transcriptionService.GetTranscription(ctx, created.ID)
		require.NoError(t, err)

		assert.NotNil(t, transcriptionResult)
		assert.Equal(t, created.ID, transcriptionResult.ID)
		assert.Len(t, segments, 1)
		assert.Equal(t, "This is a test transcription.", segments[0].Text)
	})

	t.Run("ListTranscriptions_Success", func(t *testing.T) {
		// List transcriptions for test video
		transcriptions, err := transcriptionService.ListTranscriptions(ctx, "test-video-123")
		require.NoError(t, err)

		// Should have at least 2 transcriptions (auto and en from previous tests)
		assert.GreaterOrEqual(t, len(transcriptions), 2)

		// Verify transcriptions belong to correct video
		for _, transcription := range transcriptions {
			assert.Equal(t, "test-video-123", transcription.VideoID)
		}
	})

	t.Run("DeleteTranscription_Success", func(t *testing.T) {
		// Create test video for deletion
		testVideo3 := &model.Video{
			ID:        "test-video-789",
			Title:     "Test Video for Deletion",
			URL:       "https://youtube.com/watch?v=test789",
			ChannelID: "test-channel-123",
			Duration:  60,
		}
		err = videoRepo.Create(ctx, testVideo3)
		require.NoError(t, err)

		// Create transcription
		created, err := transcriptionService.CreateTranscription(ctx, "test-video-789", "fr")
		require.NoError(t, err)

		// Delete transcription
		err = transcriptionService.DeleteTranscription(ctx, created.ID)
		require.NoError(t, err)

		// Verify transcription is deleted
		_, _, err = transcriptionService.GetTranscription(ctx, created.ID)
		assert.Error(t, err) // Should return not found error
	})
}

// runMigrations runs database migrations for testing
func runMigrations(ctx context.Context, dbPool *pgxpool.Pool) error {
	// Get the migrations directory path (relative to this test file)
	migrationsDir := filepath.Join("..", "..", "..", "..", "migrations")

	// Read migration files
	migrationFiles, err := readMigrationFiles(migrationsDir)
	if err != nil {
		return err
	}

	// Execute each migration
	for _, migrationSQL := range migrationFiles {
		if _, err := dbPool.Exec(ctx, migrationSQL); err != nil {
			return err
		}
	}

	return nil
}

// readMigrationFiles reads all .up.sql files from migrations directory
func readMigrationFiles(migrationsDir string) ([]string, error) {
	var migrations []string
	var migrationFiles []string

	// Walk through the migrations directory
	err := filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only include .up.sql files
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort migration files to ensure proper order
	sort.Strings(migrationFiles)

	// Read each migration file
	for _, file := range migrationFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, string(content))
	}

	return migrations, nil
}
