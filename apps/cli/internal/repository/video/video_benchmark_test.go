//go:build integration

package video

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/channel"
	"github.com/Taichi-iskw/yt-lang/internal/repository/common"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// BenchmarkVideoRepository_InsertMethods compares COPY FROM vs individual INSERT performance
func BenchmarkVideoRepository_InsertMethods(b *testing.B) {
	// Setup test database for benchmarking
	pool := setupBenchmarkDB(b)
	defer func() {
		if pool != nil {
			pool.Close()
		}
	}()

	// Create repository
	repo := NewRepository(pool)

	// Create test channel first
	channelRepo := channel.NewRepository(pool)
	ctx := context.Background()

	channel := &model.Channel{
		ID:   "UC_benchmark",
		Name: "Benchmark Channel",
		URL:  "https://www.youtube.com/channel/UC_benchmark",
	}
	err := channelRepo.Create(ctx, channel)
	require.NoError(b, err)

	// Test different batch sizes
	testSizes := []int{10, 100, 1000, 5000}

	for _, size := range testSizes {
		// Generate test data
		videos := generateTestVideos(channel.ID, size)

		b.Run(fmt.Sprintf("CopyFrom_BatchSize_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				// Clean up before each iteration
				cleanupVideos(b, pool, channel.ID)
				b.StartTimer()

				// Benchmark COPY FROM
				err := repo.CreateBatch(ctx, videos)
				require.NoError(b, err)
			}
		})

		b.Run(fmt.Sprintf("IndividualInsert_BatchSize_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				// Clean up before each iteration
				cleanupVideos(b, pool, channel.ID)
				b.StartTimer()

				// Benchmark individual INSERT
				for _, video := range videos {
					err := repo.Create(ctx, video)
					require.NoError(b, err)
				}
			}
		})

		b.Run(fmt.Sprintf("TransactionBatch_BatchSize_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				// Clean up before each iteration
				cleanupVideos(b, pool, channel.ID)
				b.StartTimer()

				// Benchmark transaction-based batch INSERT
				err := createBatchWithTransaction(ctx, pool, videos)
				require.NoError(b, err)
			}
		})
	}
}

// BenchmarkVideoRepository_MemoryUsage measures memory allocation patterns
func BenchmarkVideoRepository_MemoryUsage(b *testing.B) {
	// Setup test database
	pool := setupBenchmarkDB(b)
	defer func() {
		if pool != nil {
			pool.Close()
		}
	}()

	repo := NewRepository(pool)
	channelRepo := channel.NewRepository(pool)
	ctx := context.Background()

	// Create test channel
	channel := &model.Channel{
		ID:   "UC_memory_test",
		Name: "Memory Test Channel",
		URL:  "https://www.youtube.com/channel/UC_memory_test",
	}
	err := channelRepo.Create(ctx, channel)
	require.NoError(b, err)

	// Large batch for memory testing
	videos := generateTestVideos(channel.ID, 10000)

	b.Run("CopyFrom_MemoryAllocation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs() // Report memory allocations

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			cleanupVideos(b, pool, channel.ID)
			b.StartTimer()

			err := repo.CreateBatch(ctx, videos)
			require.NoError(b, err)
		}
	})

	b.Run("IndividualInsert_MemoryAllocation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs() // Report memory allocations

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			cleanupVideos(b, pool, channel.ID)
			b.StartTimer()

			for _, video := range videos {
				err := repo.Create(ctx, video)
				require.NoError(b, err)
			}
		}
	})
}

// generateTestVideos creates test video data for benchmarking
func generateTestVideos(channelID string, count int) []*model.Video {
	videos := make([]*model.Video, count)

	for i := 0; i < count; i++ {
		videos[i] = &model.Video{
			ID:        fmt.Sprintf("benchmark_video_%d", i),
			ChannelID: channelID,
			Title:     fmt.Sprintf("Benchmark Video %d - Performance Test", i),
			URL:       fmt.Sprintf("https://www.youtube.com/watch?v=benchmark_%d", i),
			Duration:  180 + (i % 300), // Vary duration between 180-480 seconds
		}
	}

	return videos
}

// setupBenchmarkDB creates a dedicated PostgreSQL database for benchmarking
func setupBenchmarkDB(b *testing.B) Pool {
	ctx := context.Background()

	// Define PostgreSQL container request
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "benchmarkdb",
			"POSTGRES_USER":     "benchuser",
			"POSTGRES_PASSWORD": "benchpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	// Start PostgreSQL container
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(b, err)

	// Get host and port
	host, err := postgresContainer.Host(ctx)
	require.NoError(b, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(b, err)

	// Build connection string
	connStr := fmt.Sprintf("postgres://benchuser:benchpass@%s:%s/benchmarkdb?sslmode=disable", host, port.Port())

	// Run migrations
	err = common.RunMigrations(connStr)
	require.NoError(b, err)

	// Create connection pool with optimized settings for benchmarking
	config, err := pgxpool.ParseConfig(connStr)
	require.NoError(b, err)

	// Optimize pool for benchmarking
	config.MaxConns = 10
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(b, err)

	// Store container for cleanup
	b.Cleanup(func() {
		if pool != nil {
			pool.Close()
		}
		if postgresContainer != nil {
			postgresContainer.Terminate(ctx)
		}
	})

	return pool
}

// cleanupVideos removes all videos for a channel (used between benchmark iterations)
func cleanupVideos(b *testing.B, pool Pool, channelID string) {
	ctx := context.Background()
	_, err := pool.Exec(ctx, "DELETE FROM videos WHERE channel_id = $1", channelID)
	require.NoError(b, err)
}

// createBatchWithTransaction inserts videos using a single transaction with prepared statements
func createBatchWithTransaction(ctx context.Context, pool Pool, videos []*model.Video) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepare statement for reuse
	_, err = tx.Prepare(ctx, "insert_video",
		"INSERT INTO videos (id, channel_id, title, url, duration) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Insert all videos using prepared statement
	for _, video := range videos {
		_, err := tx.Exec(ctx, "insert_video", video.ID, video.ChannelID, video.Title, video.URL, video.Duration)
		if err != nil {
			return fmt.Errorf("failed to execute prepared statement: %w", err)
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
