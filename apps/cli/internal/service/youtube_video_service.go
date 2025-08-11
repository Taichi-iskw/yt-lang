package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
)

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
