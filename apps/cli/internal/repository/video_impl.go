package repository

import (
	"context"
	"errors"

	apperrors "github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/jackc/pgx/v5"
)

// videoRepository implements VideoRepository using PostgreSQL
type videoRepository struct {
	pool Pool
}

// NewVideoRepository creates a new instance of VideoRepository
func NewVideoRepository(pool Pool) VideoRepository {
	return &videoRepository{
		pool: pool,
	}
}

// Create creates a new video record
func (r *videoRepository) Create(ctx context.Context, video *model.Video) error {
	sql := "INSERT INTO videos (id, channel_id, title, url, duration) VALUES ($1, $2, $3, $4, $5)"
	_, err := r.pool.Exec(ctx, sql, video.ID, video.ChannelID, video.Title, video.URL, video.Duration)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to create video")
	}
	return nil
}

// CreateBatch creates multiple video records using bulk insert
func (r *videoRepository) CreateBatch(ctx context.Context, videos []*model.Video) error {
	if len(videos) == 0 {
		return nil
	}

	// Use multiple individual inserts for now
	// TODO: Implement with COPY FROM for better performance in future iteration
	sql := "INSERT INTO videos (id, channel_id, title, url, duration) VALUES ($1, $2, $3, $4, $5)"
	for _, video := range videos {
		_, err := r.pool.Exec(ctx, sql, video.ID, video.ChannelID, video.Title, video.URL, video.Duration)
		if err != nil {
			return apperrors.Wrap(err, apperrors.CodeInternal, "failed to create video in batch")
		}
	}

	return nil
}

// GetByID retrieves a video by its ID
func (r *videoRepository) GetByID(ctx context.Context, id string) (*model.Video, error) {
	sql := "SELECT id, channel_id, title, url, duration FROM videos WHERE id = $1"
	row := r.pool.QueryRow(ctx, sql, id)

	var video model.Video
	err := row.Scan(&video.ID, &video.ChannelID, &video.Title, &video.URL, &video.Duration)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.Wrap(err, apperrors.CodeNotFound, "video not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get video")
	}

	return &video, nil
}

// GetByChannelID retrieves videos by channel ID with pagination
func (r *videoRepository) GetByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*model.Video, error) {
	sql := "SELECT id, channel_id, title, url, duration FROM videos WHERE channel_id = $1 ORDER BY id LIMIT $2 OFFSET $3"
	rows, err := r.pool.Query(ctx, sql, channelID, limit, offset)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get videos by channel ID")
	}
	defer rows.Close()

	videos := []*model.Video{}
	for rows.Next() {
		var video model.Video
		err := rows.Scan(&video.ID, &video.ChannelID, &video.Title, &video.URL, &video.Duration)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to scan video row")
		}
		videos = append(videos, &video)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to iterate video rows")
	}

	return videos, nil
}

// Update updates an existing video record
func (r *videoRepository) Update(ctx context.Context, video *model.Video) error {
	sql := "UPDATE videos SET channel_id = $2, title = $3, url = $4, duration = $5 WHERE id = $1"
	_, err := r.pool.Exec(ctx, sql, video.ID, video.ChannelID, video.Title, video.URL, video.Duration)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to update video")
	}
	return nil
}

// Delete deletes a video by its ID
func (r *videoRepository) Delete(ctx context.Context, id string) error {
	sql := "DELETE FROM videos WHERE id = $1"
	_, err := r.pool.Exec(ctx, sql, id)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete video")
	}
	return nil
}

// List retrieves videos with pagination
func (r *videoRepository) List(ctx context.Context, limit, offset int) ([]*model.Video, error) {
	sql := "SELECT id, channel_id, title, url, duration FROM videos ORDER BY id LIMIT $1 OFFSET $2"
	rows, err := r.pool.Query(ctx, sql, limit, offset)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to list videos")
	}
	defer rows.Close()

	videos := []*model.Video{}
	for rows.Next() {
		var video model.Video
		err := rows.Scan(&video.ID, &video.ChannelID, &video.Title, &video.URL, &video.Duration)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to scan video row")
		}
		videos = append(videos, &video)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to iterate video rows")
	}

	return videos, nil
}
