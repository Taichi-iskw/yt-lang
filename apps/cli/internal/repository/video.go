package repository

import (
	"context"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// VideoRepository defines operations for Video persistence
type VideoRepository interface {
	// Create creates a new video record
	Create(ctx context.Context, video *model.Video) error
	
	// GetByID retrieves a video by its ID
	GetByID(ctx context.Context, id string) (*model.Video, error)
	
	// GetByChannelID retrieves videos by channel ID with pagination
	GetByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*model.Video, error)
	
	// Update updates an existing video record
	Update(ctx context.Context, video *model.Video) error
	
	// Delete deletes a video by its ID
	Delete(ctx context.Context, id string) error
	
	// List retrieves videos with pagination
	List(ctx context.Context, limit, offset int) ([]*model.Video, error)
}