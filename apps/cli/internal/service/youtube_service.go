package service

import (
	"context"

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
