package transcription

import (
	"context"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/repository/common"
	"github.com/jackc/pgx/v5"
)

// segmentRepository implements SegmentRepository using PostgreSQL
type segmentRepository struct {
	pool Pool
}

// NewSegmentRepository creates a new instance of SegmentRepository
func NewSegmentRepository(pool Pool) SegmentRepository {
	return &segmentRepository{
		pool: pool,
	}
}

// CreateBatch creates multiple transcription segments using COPY FROM for performance
func (r *segmentRepository) CreateBatch(ctx context.Context, segments []*model.TranscriptionSegment) error {
	if len(segments) == 0 {
		return nil // Nothing to insert
	}

	// Prepare data for COPY FROM
	rows := make([][]interface{}, len(segments))
	for i, segment := range segments {
		rows[i] = []interface{}{
			segment.TranscriptionID,
			segment.SegmentIndex,
			segment.StartTime,
			segment.EndTime,
			segment.Text,
			segment.Confidence,
		}
	}

	// Use COPY FROM for efficient bulk insert
	_, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"transcription_segments"},
		[]string{"transcription_id", "segment_index", "start_time", "end_time", "text", "confidence"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return common.HandlePostgreSQLError(err, "failed to create transcription segments")
	}

	return nil
}

// GetByTranscriptionID retrieves all segments for a transcription, ordered by segment_index
func (r *segmentRepository) GetByTranscriptionID(ctx context.Context, transcriptionID string) ([]*model.TranscriptionSegment, error) {
	sql := `SELECT id, transcription_id, segment_index, 
		start_time::text, end_time::text, text, confidence 
		FROM transcription_segments 
		WHERE transcription_id = $1 
		ORDER BY segment_index`

	rows, err := r.pool.Query(ctx, sql, transcriptionID)
	if err != nil {
		return nil, common.HandlePostgreSQLError(err, "failed to get transcription segments")
	}
	defer rows.Close()

	var segments []*model.TranscriptionSegment
	for rows.Next() {
		var segment model.TranscriptionSegment
		err := rows.Scan(
			&segment.ID,
			&segment.TranscriptionID,
			&segment.SegmentIndex,
			&segment.StartTime,
			&segment.EndTime,
			&segment.Text,
			&segment.Confidence,
		)
		if err != nil {
			return nil, common.HandlePostgreSQLError(err, "failed to scan transcription segment")
		}
		segments = append(segments, &segment)
	}

	return segments, nil
}

// GetByTimeRange retrieves segments within a time range
func (r *segmentRepository) GetByTimeRange(ctx context.Context, transcriptionID string, startTime, endTime string) ([]*model.TranscriptionSegment, error) {
	sql := `SELECT id, transcription_id, segment_index, 
		start_time::text, end_time::text, text, confidence 
		FROM transcription_segments 
		WHERE transcription_id = $1 
		AND start_time >= $2::interval 
		AND end_time <= $3::interval
		ORDER BY segment_index`

	rows, err := r.pool.Query(ctx, sql, transcriptionID, startTime, endTime)
	if err != nil {
		return nil, common.HandlePostgreSQLError(err, "failed to get transcription segments by time range")
	}
	defer rows.Close()

	var segments []*model.TranscriptionSegment
	for rows.Next() {
		var segment model.TranscriptionSegment
		err := rows.Scan(
			&segment.ID,
			&segment.TranscriptionID,
			&segment.SegmentIndex,
			&segment.StartTime,
			&segment.EndTime,
			&segment.Text,
			&segment.Confidence,
		)
		if err != nil {
			return nil, common.HandlePostgreSQLError(err, "failed to scan transcription segment")
		}
		segments = append(segments, &segment)
	}

	return segments, nil
}

// Delete deletes all segments for a transcription
func (r *segmentRepository) Delete(ctx context.Context, transcriptionID string) error {
	sql := "DELETE FROM transcription_segments WHERE transcription_id = $1"
	_, err := r.pool.Exec(ctx, sql, transcriptionID)
	if err != nil {
		return common.HandlePostgreSQLError(err, "failed to delete transcription segments")
	}
	return nil
}
