package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Taichi-iskw/yt-lang/internal/service"
)

// channelCmd represents the channel command
var channelCmd = &cobra.Command{
	Use:   "channel [URL]",
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

func init() {
	rootCmd.AddCommand(channelCmd)
}