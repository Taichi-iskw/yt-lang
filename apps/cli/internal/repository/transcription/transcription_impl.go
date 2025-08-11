package transcription

import (
	"context"
	"errors"

	apperrors "github.com/Taichi-iskw/yt-lang/internal/errors"
	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/common"
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

// transcriptionRepository implements Repository using PostgreSQL
type transcriptionRepository struct {
	pool Pool
}

// NewRepository creates a new instance of Repository
func NewRepository(pool Pool) Repository {
	return &transcriptionRepository{
		pool: pool,
	}
}

// Create creates a new transcription record
func (r *transcriptionRepository) Create(ctx context.Context, transcription *model.Transcription) error {
	sql := `INSERT INTO transcriptions 
		(video_id, language, status, created_at, completed_at, error_message, detected_language, total_duration) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err := r.pool.QueryRow(ctx, sql,
		transcription.VideoID,
		transcription.Language,
		transcription.Status,
		transcription.CreatedAt,
		transcription.CompletedAt,
		transcription.ErrorMessage,
		transcription.DetectedLanguage,
		transcription.TotalDuration,
	).Scan(&transcription.ID)
	if err != nil {
		return common.HandlePostgreSQLError(err, "failed to create transcription")
	}
	return nil
}

// GetByID retrieves a transcription by its ID
func (r *transcriptionRepository) GetByID(ctx context.Context, id string) (*model.Transcription, error) {
	sql := `SELECT id, video_id, language, status, created_at, completed_at, error_message, detected_language, total_duration 
		FROM transcriptions WHERE id = $1`
	row := r.pool.QueryRow(ctx, sql, id)

	var transcription model.Transcription
	err := row.Scan(
		&transcription.ID,
		&transcription.VideoID,
		&transcription.Language,
		&transcription.Status,
		&transcription.CreatedAt,
		&transcription.CompletedAt,
		&transcription.ErrorMessage,
		&transcription.DetectedLanguage,
		&transcription.TotalDuration,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.Wrap(err, apperrors.CodeNotFound, "transcription not found")
		}
		return nil, common.HandlePostgreSQLError(err, "failed to get transcription")
	}
	return &transcription, nil
}

// GetByVideoID retrieves all transcriptions for a video
func (r *transcriptionRepository) GetByVideoID(ctx context.Context, videoID string) ([]*model.Transcription, error) {
	sql := `SELECT id, video_id, language, status, created_at, completed_at, error_message, detected_language, total_duration 
		FROM transcriptions WHERE video_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, sql, videoID)
	if err != nil {
		return nil, common.HandlePostgreSQLError(err, "failed to get transcriptions by video ID")
	}
	defer rows.Close()

	var transcriptions []*model.Transcription
	for rows.Next() {
		var transcription model.Transcription
		err := rows.Scan(
			&transcription.ID,
			&transcription.VideoID,
			&transcription.Language,
			&transcription.Status,
			&transcription.CreatedAt,
			&transcription.CompletedAt,
			&transcription.ErrorMessage,
			&transcription.DetectedLanguage,
			&transcription.TotalDuration,
		)
		if err != nil {
			return nil, common.HandlePostgreSQLError(err, "failed to scan transcription")
		}
		transcriptions = append(transcriptions, &transcription)
	}

	return transcriptions, nil
}

// GetByVideoIDAndLanguage retrieves a transcription for a video in specific language
func (r *transcriptionRepository) GetByVideoIDAndLanguage(ctx context.Context, videoID, language string) (*model.Transcription, error) {
	sql := `SELECT id, video_id, language, status, created_at, completed_at, error_message, detected_language, total_duration 
		FROM transcriptions WHERE video_id = $1 AND language = $2`
	row := r.pool.QueryRow(ctx, sql, videoID, language)

	var transcription model.Transcription
	err := row.Scan(
		&transcription.ID,
		&transcription.VideoID,
		&transcription.Language,
		&transcription.Status,
		&transcription.CreatedAt,
		&transcription.CompletedAt,
		&transcription.ErrorMessage,
		&transcription.DetectedLanguage,
		&transcription.TotalDuration,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.Wrap(err, apperrors.CodeNotFound, "transcription not found")
		}
		return nil, common.HandlePostgreSQLError(err, "failed to get transcription")
	}
	return &transcription, nil
}

// UpdateStatus updates the status of a transcription
func (r *transcriptionRepository) UpdateStatus(ctx context.Context, id string, status string, errorMessage *string) error {
	sql := `UPDATE transcriptions SET status = $2, error_message = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, sql, id, status, errorMessage)
	if err != nil {
		return common.HandlePostgreSQLError(err, "failed to update transcription status")
	}
	return nil
}

// Delete deletes a transcription by ID
func (r *transcriptionRepository) Delete(ctx context.Context, id string) error {
	sql := "DELETE FROM transcriptions WHERE id = $1"
	_, err := r.pool.Exec(ctx, sql, id)
	if err != nil {
		return common.HandlePostgreSQLError(err, "failed to delete transcription")
	}
	return nil
}
