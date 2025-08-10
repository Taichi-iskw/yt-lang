package repository

import (
	"context"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// ChannelRepository defines operations for Channel persistence
type ChannelRepository interface {
	// Create creates a new channel record
	Create(ctx context.Context, channel *model.Channel) error
	
	// GetByID retrieves a channel by its ID
	GetByID(ctx context.Context, id string) (*model.Channel, error)
	
	// GetByURL retrieves a channel by its URL
	GetByURL(ctx context.Context, url string) (*model.Channel, error)
	
	// Update updates an existing channel record
	Update(ctx context.Context, channel *model.Channel) error
	
	// Delete deletes a channel by its ID
	Delete(ctx context.Context, id string) error
	
	// List retrieves channels with pagination
	List(ctx context.Context, limit, offset int) ([]*model.Channel, error)
}