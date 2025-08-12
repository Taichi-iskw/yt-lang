package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Taichi-iskw/yt-lang/internal/config"
	"github.com/Taichi-iskw/yt-lang/internal/repository/channel"
	"github.com/Taichi-iskw/yt-lang/internal/repository/video"
	"github.com/Taichi-iskw/yt-lang/internal/service/common"
	youtubeSvc "github.com/Taichi-iskw/yt-lang/internal/service/youtube"
)

// videoCmd represents the video command
var videoCmd = &cobra.Command{
	Use:   "video",
	Short: "YouTube video operations",
	Long:  `Operations for managing YouTube videos from channels.`,
}


// videoSaveCmd saves videos from a channel to database
var videoSaveCmd = &cobra.Command{
	Use:   "save [CHANNEL_URL]",
	Short: "Save videos from a YouTube channel to database",
	Long:  `Fetch videos from a YouTube channel and save them to the database.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		channelURL := args[0]

		// Create service with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Load configuration
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
		channelRepo := channel.NewRepository(dbPool)
		videoRepo := video.NewRepository(dbPool)

		// Create YouTube service with repositories
		youtubeService := youtubeSvc.NewYouTubeServiceWithRepositories(
			common.NewCmdRunner(),
			channelRepo,
			videoRepo,
		)

		// Get dry-run flag
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// If dry-run, fetch videos without saving (limit = 0 means all videos)
		if dryRun {
			videos, err := youtubeService.FetchChannelVideos(ctx, channelURL, 0)
			if err != nil {
				return fmt.Errorf("failed to fetch videos (dry-run): %w", err)
			}

			fmt.Printf("[DRY RUN] Would save %d video(s):\n", len(videos))
			result, err := json.MarshalIndent(videos, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format result: %w", err)
			}
			fmt.Println(string(result))
			return nil
		}

		// Save videos (limit = 0 means all videos)
		videos, err := youtubeService.SaveChannelVideos(ctx, channelURL, 0)
		if err != nil {
			return fmt.Errorf("failed to save videos: %w", err)
		}

		// Display result as JSON
		result, err := json.MarshalIndent(videos, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format result: %w", err)
		}

		fmt.Printf("%d video(s) saved successfully:\n%s\n", len(videos), string(result))
		return nil
	},
}

// videoListCmd lists videos for a specific channel
var videoListCmd = &cobra.Command{
	Use:   "list [CHANNEL_ID]",
	Short: "List videos for a specific channel",
	Long:  `List videos for a specific channel saved in the database.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		channelID := args[0]

		// Create service with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Load configuration
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
		channelRepo := channel.NewRepository(dbPool)
		videoRepo := video.NewRepository(dbPool)

		// Create YouTube service with repositories
		youtubeService := youtubeSvc.NewYouTubeServiceWithRepositories(
			common.NewCmdRunner(),
			channelRepo,
			videoRepo,
		)

		// Get pagination flags
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		// List videos
		videos, err := youtubeService.ListVideos(ctx, channelID, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to list videos: %w", err)
		}

		// Check if no videos found
		if len(videos) == 0 {
			fmt.Printf("No videos found for channel ID: %s\n", channelID)
			return nil
		}

		// Display result as JSON
		result, err := json.MarshalIndent(videos, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format result: %w", err)
		}

		fmt.Printf("Found %d video(s) for channel %s:\n%s\n", len(videos), channelID, string(result))
		return nil
	},
}

func init() {
	// Add flags to save command
	videoSaveCmd.Flags().Bool("dry-run", false, "Preview videos without saving to database")

	// Add pagination flags to list command
	videoListCmd.Flags().Int("limit", 10, "Maximum number of videos to retrieve")
	videoListCmd.Flags().Int("offset", 0, "Number of videos to skip")

	videoCmd.AddCommand(videoSaveCmd)
	videoCmd.AddCommand(videoListCmd)
	rootCmd.AddCommand(videoCmd)
}
