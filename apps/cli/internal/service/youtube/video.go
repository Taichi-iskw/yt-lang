package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// FetchChannelVideos fetches video list from YouTube channel ID using yt-dlp
func (s *youTubeService) FetchChannelVideos(ctx context.Context, channelID string, limit int) ([]*model.Video, error) {
	// Input validation
	if channelID == "" {
		return nil, errors.New(errors.CodeInvalidArg, "channel ID is required")
	}

	// Validate channel ID format (must start with UC)
	if !strings.HasPrefix(channelID, "UC") {
		return nil, errors.New(errors.CodeInvalidArg, "invalid channel ID format (must start with UC)")
	}

	// Build yt-dlp command arguments with channel ID
	channelURL := "https://www.youtube.com/channel/" + channelID
	args := []string{
		"--dump-json",
		"--flat-playlist",
		channelURL,
	}

	// Add limit if specified (0 means no limit - fetch all videos)
	if limit > 0 {
		// Insert limit arguments after --dump-json and --flat-playlist
		limitArgs := []string{"--playlist-end", fmt.Sprintf("%d", limit)}
		args = append(args[:2], append(limitArgs, args[2:]...)...)
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

		// Use the input channel ID (we know it's correct)
		// If yt-dlp returns a different channel_id, we trust our input more
		videoChannelID := channelID
		if ytInfo.ChannelID != "" && ytInfo.ChannelID != channelID {
			// Log the discrepancy but use our input channel ID
			// In a real implementation, you might want to log this
		}

		// Convert to our model
		video := &model.Video{
			ID:        ytInfo.ID,
			ChannelID: videoChannelID,
			Title:     ytInfo.Title,
			URL:       ytInfo.URL,
			Duration:  ytInfo.Duration,
		}
		videos = append(videos, video)
	}

	return videos, nil
}

// SaveChannelVideos fetches channel videos from YouTube channel ID and saves them to database
func (s *youTubeService) SaveChannelVideos(ctx context.Context, channelID string, limit int) ([]*model.Video, error) {
	// Note: We assume the channel already exists in database with this channel ID
	// In a complete implementation, you might want to verify this first

	// Fetch videos from the channel
	videos, err := s.FetchChannelVideos(ctx, channelID, limit)
	if err != nil {
		return nil, err
	}

	// Save videos to database using upsert batch (handles duplicates)
	err = s.videoRepo.UpsertBatch(ctx, videos)
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
