package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository"
)

// YouTubeService is interface for YouTube operations
type YouTubeService interface {
	FetchChannelInfo(ctx context.Context, channelURL string) (*model.Channel, error)
	SaveChannelInfo(ctx context.Context, channelURL string) (*model.Channel, error)
	ListChannels(ctx context.Context, limit, offset int) ([]*model.Channel, error)
	FetchChannelVideos(ctx context.Context, channelURL string, limit int) ([]*model.Video, error)
	SaveChannelVideos(ctx context.Context, channelURL string, limit int) ([]*model.Video, error)
	ListVideos(ctx context.Context, channelID string, limit, offset int) ([]*model.Video, error)
}

// youTubeService implements YouTubeService
type youTubeService struct {
	cmdRunner   CmdRunner
	channelRepo repository.ChannelRepository
	videoRepo   repository.VideoRepository
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

// NewYouTubeServiceWithRepositories creates a new YouTubeService with custom repositories (for testing)
func NewYouTubeServiceWithRepositories(cmdRunner CmdRunner, channelRepo repository.ChannelRepository, videoRepo repository.VideoRepository) YouTubeService {
	return &youTubeService{
		cmdRunner:   cmdRunner,
		channelRepo: channelRepo,
		videoRepo:   videoRepo,
	}
}

// ytDlpChannelInfo represents yt-dlp JSON output structure for channel info
type ytDlpChannelInfo struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Channel    string `json:"channel"`
	ChannelURL string `json:"channel_url"`
}

// ytDlpVideoInfo represents yt-dlp JSON output structure for video info
type ytDlpVideoInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ChannelID string `json:"channel_id"`
	URL       string `json:"webpage_url"`
	Duration  int    `json:"duration"`
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

// SaveChannelInfo fetches channel information from YouTube URL and saves it to database
func (s *youTubeService) SaveChannelInfo(ctx context.Context, channelURL string) (*model.Channel, error) {
	// First, fetch channel info using existing method
	channel, err := s.FetchChannelInfo(ctx, channelURL)
	if err != nil {
		return nil, err
	}

	// Check if channel already exists in database
	existingChannel, err := s.channelRepo.GetByURL(ctx, channel.URL)
	if err == nil {
		// Channel already exists, return existing one
		return existingChannel, nil
	}

	// Channel doesn't exist, create new one
	err = s.channelRepo.Create(ctx, channel)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "failed to save channel to database")
	}

	return channel, nil
}

// ListChannels retrieves all saved channels with pagination
func (s *youTubeService) ListChannels(ctx context.Context, limit, offset int) ([]*model.Channel, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	// Fetch channels from repository
	channels, err := s.channelRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "failed to list channels")
	}

	return channels, nil
}

// FetchChannelVideos fetches video list from YouTube channel URL using yt-dlp
func (s *youTubeService) FetchChannelVideos(ctx context.Context, channelURL string, limit int) ([]*model.Video, error) {
	// Input validation
	if channelURL == "" {
		return nil, errors.New(errors.CodeInvalidArg, "channel URL is required")
	}
	if limit <= 0 {
		return nil, errors.New(errors.CodeInvalidArg, "limit must be greater than 0")
	}

	// Execute yt-dlp command to get video list from channel
	args := []string{
		"--dump-json",
		"--flat-playlist",
		"--playlist-end", fmt.Sprintf("%d", limit),
		channelURL,
	}

	output, err := s.cmdRunner.Run(ctx, "yt-dlp", args...)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeExternal, "failed to fetch channel videos with yt-dlp")
	}

	// Parse JSON response (yt-dlp outputs one JSON object per line)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	videos := make([]*model.Video, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		var ytInfo ytDlpVideoInfo
		if err := json.Unmarshal([]byte(line), &ytInfo); err != nil {
			return nil, errors.Wrap(err, errors.CodeInternal, "failed to parse yt-dlp output")
		}

		// Convert to our model
		video := &model.Video{
			ID:        ytInfo.ID,
			ChannelID: ytInfo.ChannelID,
			Title:     ytInfo.Title,
			URL:       ytInfo.URL,
			Duration:  ytInfo.Duration,
		}
		videos = append(videos, video)
	}

	return videos, nil
}

// SaveChannelVideos fetches channel videos from YouTube URL and saves them to database
func (s *youTubeService) SaveChannelVideos(ctx context.Context, channelURL string, limit int) ([]*model.Video, error) {
	// First, ensure the channel exists in database
	_, err := s.SaveChannelInfo(ctx, channelURL)
	if err != nil {
		return nil, err
	}

	// Fetch videos from the channel
	videos, err := s.FetchChannelVideos(ctx, channelURL, limit)
	if err != nil {
		return nil, err
	}

	// Save videos to database using batch create
	err = s.videoRepo.CreateBatch(ctx, videos)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "failed to save videos to database")
	}

	return videos, nil
}

// ListVideos retrieves videos for a specific channel with pagination
func (s *youTubeService) ListVideos(ctx context.Context, channelID string, limit, offset int) ([]*model.Video, error) {
	// Input validation
	if channelID == "" {
		return nil, errors.New(errors.CodeInvalidArg, "channel ID is required")
	}

	// Validate pagination parameters
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	// Fetch videos from repository
	videos, err := s.videoRepo.GetByChannelID(ctx, channelID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "failed to list videos")
	}

	return videos, nil
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
