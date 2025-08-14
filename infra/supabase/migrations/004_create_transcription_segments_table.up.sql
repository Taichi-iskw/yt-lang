-- Create transcription_segments table for storing individual whisper segments (Option B: Normalized)
CREATE TABLE IF NOT EXISTS transcription_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transcription_id UUID NOT NULL REFERENCES transcriptions(id) ON DELETE CASCADE,
    segment_index INTEGER NOT NULL, -- Sequence order (starting from 0)
    start_time INTERVAL NOT NULL,
    end_time INTERVAL NOT NULL,
    text TEXT NOT NULL,
    confidence FLOAT, -- Whisper confidence score (0.0-1.0)
    
    UNIQUE(transcription_id, segment_index),
    
    -- Time consistency checks
    CONSTRAINT check_time_order 
        CHECK (start_time < end_time),
    CONSTRAINT check_positive_times
        CHECK (start_time >= '00:00:00' AND end_time >= '00:00:00')
);

-- Essential indexes for performance
CREATE INDEX IF NOT EXISTS idx_transcription_segments_transcription_id ON transcription_segments(transcription_id);
CREATE INDEX IF NOT EXISTS idx_transcription_segments_transcription_index ON transcription_segments(transcription_id, segment_index);
CREATE INDEX IF NOT EXISTS idx_transcription_segments_time_range ON transcription_segments(transcription_id, start_time, end_time);