-- Create transcriptions table for storing video transcription metadata (Option B: Normalized)
CREATE TABLE IF NOT EXISTS transcriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id VARCHAR(255) NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- Language code: 'ja', 'en', 'auto', etc.
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- Status: 'pending', 'processing', 'completed', 'failed', 'cancelled'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    
    -- Whisper result metadata (minimal)
    detected_language VARCHAR(10), -- Actual language detected by Whisper
    total_duration INTERVAL,
    
    UNIQUE(video_id, language) -- Prevent duplicates
);

-- Essential indexes for performance
CREATE INDEX IF NOT EXISTS idx_transcriptions_video_id ON transcriptions(video_id);
CREATE INDEX IF NOT EXISTS idx_transcriptions_status ON transcriptions(status);
CREATE INDEX IF NOT EXISTS idx_transcriptions_video_lang ON transcriptions(video_id, language);