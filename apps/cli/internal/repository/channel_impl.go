package repository

import (
	"context"
	"errors"

	apperrors "github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Pool interface for abstracting pgx connection pool
type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

// channelRepository implements ChannelRepository using PostgreSQL
type channelRepository struct {
	pool Pool
}

// NewChannelRepository creates a new instance of ChannelRepository
func NewChannelRepository(pool Pool) ChannelRepository {
	return &channelRepository{
		pool: pool,
	}
}

// Create creates a new channel record
func (r *channelRepository) Create(ctx context.Context, channel *model.Channel) error {
	sql := "INSERT INTO channels (id, name, url) VALUES ($1, $2, $3)"
	_, err := r.pool.Exec(ctx, sql, channel.ID, channel.Name, channel.URL)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to create channel")
	}
	return nil
}

// GetByID retrieves a channel by its ID
func (r *channelRepository) GetByID(ctx context.Context, id string) (*model.Channel, error) {
	sql := "SELECT id, name, url FROM channels WHERE id = $1"
	row := r.pool.QueryRow(ctx, sql, id)

	var channel model.Channel
	err := row.Scan(&channel.ID, &channel.Name, &channel.URL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.Wrap(err, apperrors.CodeNotFound, "channel not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get channel")
	}

	return &channel, nil
}

// GetByURL retrieves a channel by its URL
func (r *channelRepository) GetByURL(ctx context.Context, url string) (*model.Channel, error) {
	sql := "SELECT id, name, url FROM channels WHERE url = $1"
	row := r.pool.QueryRow(ctx, sql, url)

	var channel model.Channel
	err := row.Scan(&channel.ID, &channel.Name, &channel.URL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.Wrap(err, apperrors.CodeNotFound, "channel not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get channel by URL")
	}

	return &channel, nil
}

// Update updates an existing channel record
func (r *channelRepository) Update(ctx context.Context, channel *model.Channel) error {
	sql := "UPDATE channels SET name = $2, url = $3 WHERE id = $1"
	_, err := r.pool.Exec(ctx, sql, channel.ID, channel.Name, channel.URL)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to update channel")
	}
	return nil
}

// Delete deletes a channel by its ID
func (r *channelRepository) Delete(ctx context.Context, id string) error {
	sql := "DELETE FROM channels WHERE id = $1"
	_, err := r.pool.Exec(ctx, sql, id)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete channel")
	}
	return nil
}

// List retrieves channels with pagination
func (r *channelRepository) List(ctx context.Context, limit, offset int) ([]*model.Channel, error) {
	sql := "SELECT id, name, url FROM channels ORDER BY id LIMIT $1 OFFSET $2"
	rows, err := r.pool.Query(ctx, sql, limit, offset)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to list channels")
	}
	defer rows.Close()

	var channels []*model.Channel
	for rows.Next() {
		var channel model.Channel
		err := rows.Scan(&channel.ID, &channel.Name, &channel.URL)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to scan channel row")
		}
		channels = append(channels, &channel)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to iterate channel rows")
	}

	return channels, nil
}
