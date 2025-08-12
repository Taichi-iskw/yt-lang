package translation

import (
	"context"

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

// translationRepository implements TranslationRepository
type translationRepository struct {
	pool Pool
}

// NewTranslationRepository creates a new translation repository
func NewTranslationRepository(pool Pool) TranslationRepository {
	return &translationRepository{
		pool: pool,
	}
}

// NewRepository creates a new translation repository (alias for consistency with other repos)
func NewRepository(pool Pool) TranslationRepository {
	return NewTranslationRepository(pool)
}

// Create creates a new translation record
func (r *translationRepository) Create(ctx context.Context, translation *model.Translation) error {
	query := `
		INSERT INTO translations (transcription_id, target_language, content, source)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		translation.TranscriptionID,
		translation.TargetLanguage,
		translation.Content,
		translation.Source).Scan(&translation.ID, &translation.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

// Get retrieves a translation by ID
func (r *translationRepository) Get(ctx context.Context, id int) (*model.Translation, error) {
	query := `
		SELECT id, transcription_id, target_language, content, source, created_at
		FROM translations
		WHERE id = $1`

	var translation model.Translation
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&translation.ID, &translation.TranscriptionID, &translation.TargetLanguage,
			&translation.Content, &translation.Source, &translation.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &translation, nil
}

// GetByTranscriptionIDAndLanguage retrieves translation by transcription ID and target language
func (r *translationRepository) GetByTranscriptionIDAndLanguage(ctx context.Context, transcriptionID string, targetLanguage string) (*model.Translation, error) {
	query := `
		SELECT id, transcription_id, target_language, content, source, created_at
		FROM translations
		WHERE transcription_id = $1 AND target_language = $2`

	var translation model.Translation
	err := r.pool.QueryRow(ctx, query, transcriptionID, targetLanguage).
		Scan(&translation.ID, &translation.TranscriptionID, &translation.TargetLanguage,
			&translation.Content, &translation.Source, &translation.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &translation, nil
}

// Delete removes a translation record
func (r *translationRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM translations WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// CreateBatch creates multiple translations (placeholder implementation)
func (r *translationRepository) CreateBatch(ctx context.Context, translations []*model.Translation) error {
	if len(translations) == 0 {
		return nil
	}

	// Prepare data for COPY FROM
	rows := make([][]interface{}, len(translations))
	for i, t := range translations {
		rows[i] = []interface{}{
			t.TranscriptionID,
			t.TargetLanguage,
			t.Content,
			t.Source,
		}
	}

	// Use CopyFrom for efficient bulk insert
	columns := []string{"transcription_id", "target_language", "content", "source"}
	count, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"translations"},
		columns,
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return err
	}

	// Optionally, you can log the number of rows inserted
	_ = count

	return nil
}

// GetByTranscriptionID retrieves all translations for a transcription (placeholder implementation)
func (r *translationRepository) GetByTranscriptionID(ctx context.Context, transcriptionID string) ([]*model.Translation, error) {
	// TODO: implement
	return []*model.Translation{}, nil
}

// ListByTranscriptionID retrieves translations for a transcription segment with pagination
func (r *translationRepository) ListByTranscriptionID(ctx context.Context, transcriptionID string, limit, offset int) ([]*model.Translation, error) {
	query := `
		SELECT id, transcription_id, target_language, content, source, created_at
		FROM translations
		WHERE transcription_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, transcriptionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translations []*model.Translation
	for rows.Next() {
		var translation model.Translation
		err := rows.Scan(&translation.ID, &translation.TranscriptionID, &translation.TargetLanguage,
			&translation.Content, &translation.Source, &translation.CreatedAt)
		if err != nil {
			return nil, err
		}
		translations = append(translations, &translation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return translations, nil
}

// GetByVideoIDAndLanguage retrieves translations by video ID and language (placeholder implementation)
func (r *translationRepository) GetByVideoIDAndLanguage(ctx context.Context, videoID, targetLanguage string) ([]*model.Translation, error) {
	// TODO: implement
	return []*model.Translation{}, nil
}

// Update updates a translation (placeholder implementation)
func (r *translationRepository) Update(ctx context.Context, translation *model.Translation) error {
	// TODO: implement
	return nil
}

// DeleteByTranscriptionID deletes translations by transcription ID (placeholder implementation)
func (r *translationRepository) DeleteByTranscriptionID(ctx context.Context, transcriptionID string) error {
	// TODO: implement
	return nil
}

// DeleteByVideoID deletes translations by video ID (placeholder implementation)
func (r *translationRepository) DeleteByVideoID(ctx context.Context, videoID string) error {
	// TODO: implement
	return nil
}
