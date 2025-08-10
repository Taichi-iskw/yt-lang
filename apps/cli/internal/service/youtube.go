package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// YouTubeService is interface for YouTube operations
type YouTubeService interface {
	FetchChannelInfo(ctx context.Context, channelURL string) (*model.Channel, error)
}

// youTubeService implements YouTubeService
type youTubeService struct {
	cmdRunner CmdRunner
}

// NewYouTubeService creates a new YouTubeService
func NewYouTubeService() YouTubeService {
	return NewYouTubeServiceWithCmdRunner(NewCmdRunner())
}

// NewYouTubeServiceWithCmdRunner creates a new YouTubeService with custom CmdRunner (for testing)
func NewYouTubeServiceWithCmdRunner(cmdRunner CmdRunner) YouTubeService {
	return &youTubeService{
		cmdRunner: cmdRunner,
	}
}

// ytDlpChannelInfo represents yt-dlp JSON output structure for channel info
type ytDlpChannelInfo struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Channel    string `json:"channel"`
	ChannelURL string `json:"channel_url"`
}

// FetchChannelInfo fetches channel information from YouTube URL using yt-dlp
func (s *youTubeService) FetchChannelInfo(ctx context.Context, channelURL string) (*model.Channel, error) {
	// Input validation
	if channelURL == "" {
		return nil, errors.New(errors.CodeInvalidArg, "channel URL is required")
	}

	// Execute yt-dlp command to get channel information
	args := []string{
		"--dump-json",
		"--playlist-items", "1", // Get only first video to extract channel info
		channelURL,
	}

	output, err := s.cmdRunner.Run(ctx, "yt-dlp", args...)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeExternal, "failed to fetch channel info with yt-dlp")
	}

	// Parse JSON response
	var ytInfo ytDlpChannelInfo
	if err := json.Unmarshal(output, &ytInfo); err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "failed to parse yt-dlp output")
	}

	// Convert to our model
	channel := &model.Channel{
		ID:   extractChannelID(ytInfo.ID),
		Name: ytInfo.Channel,
		URL:  ytInfo.ChannelURL,
	}

	return channel, nil
}

// extractChannelID extracts channel ID from various formats
func extractChannelID(id string) string {
	// yt-dlp may return video ID, need to handle channel ID extraction
	// For now, return as-is (will be improved later)
	if strings.HasPrefix(id, "UC") {
		return id
	}
	return "UC" + id
}
