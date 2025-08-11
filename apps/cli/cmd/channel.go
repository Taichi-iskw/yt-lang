package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Taichi-iskw/yt-lang/internal/config"
	"github.com/Taichi-iskw/yt-lang/internal/repository"
	"github.com/Taichi-iskw/yt-lang/internal/service"
)

// channelCmd represents the channel command
var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "YouTube channel operations",
	Long:  `Operations for managing YouTube channels.`,
}

// channelInfoCmd fetches channel information
var channelInfoCmd = &cobra.Command{
	Use:   "info [URL]",
	Short: "Fetch YouTube channel information",
	Long:  `Fetch and display YouTube channel information using yt-dlp.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		channelURL := args[0]

		// Create service with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Create YouTube service
		youtubeService := service.NewYouTubeService()

		// Fetch channel info
		channel, err := youtubeService.FetchChannelInfo(ctx, channelURL)
		if err != nil {
			return fmt.Errorf("failed to fetch channel info: %w", err)
		}

		// Display result as JSON
		result, err := json.MarshalIndent(channel, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format result: %w", err)
		}

		fmt.Println(string(result))
		return nil
	},
}

// channelSaveCmd saves channel information to database
var channelSaveCmd = &cobra.Command{
	Use:   "save [URL]",
	Short: "Save YouTube channel information to database",
	Long:  `Fetch YouTube channel information and save it to the database.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		channelURL := args[0]

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
		channelRepo := repository.NewChannelRepository(dbPool)
		videoRepo := repository.NewVideoRepository(dbPool)

		// Create YouTube service with repositories
		youtubeService := service.NewYouTubeServiceWithRepositories(
			service.NewCmdRunner(),
			channelRepo,
			videoRepo,
		)

		// Save channel info
		channel, err := youtubeService.SaveChannelInfo(ctx, channelURL)
		if err != nil {
			return fmt.Errorf("failed to save channel info: %w", err)
		}

		// Display result as JSON
		result, err := json.MarshalIndent(channel, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format result: %w", err)
		}

		fmt.Printf("Channel saved successfully:\n%s\n", string(result))
		return nil
	},
}

func init() {
	channelCmd.AddCommand(channelInfoCmd)
	channelCmd.AddCommand(channelSaveCmd)
	rootCmd.AddCommand(channelCmd)
}
