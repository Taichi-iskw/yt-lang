-- Create transcriptions table for storing video subtitle segments
CREATE TABLE IF NOT EXISTS transcriptions (
    id SERIAL PRIMARY KEY,
    video_id VARCHAR(255) NOT NULL,        -- Foreign key to videos.id
    start_time DECIMAL(10,3) NOT NULL,     -- Start time in seconds (e.g., 12.345)
    end_time DECIMAL(10,3) NOT NULL,       -- End time in seconds (e.g., 15.678)
    content TEXT NOT NULL,                 -- Transcription text content for this segment
    language VARCHAR(10) NOT NULL DEFAULT 'en', -- Language code (e.g., 'en', 'ja')
    source VARCHAR(50) NOT NULL DEFAULT 'whisper', -- Source: whisper, youtube, etc
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Keep for audit/debugging
    
    -- Foreign key constraint
    CONSTRAINT fk_transcriptions_video_id 
        FOREIGN KEY (video_id) 
        REFERENCES videos(id) 
        ON DELETE CASCADE,
        
    -- Ensure start_time < end_time
    CONSTRAINT check_time_order 
        CHECK (start_time < end_time),
        
    -- Prevent negative times
    CONSTRAINT check_positive_times
        CHECK (start_time >= 0 AND end_time >= 0)
);

-- Essential indexes for performance
CREATE INDEX IF NOT EXISTS idx_transcriptions_video_id ON transcriptions(video_id);
CREATE INDEX IF NOT EXISTS idx_transcriptions_video_time ON transcriptions(video_id, start_time);
CREATE INDEX IF NOT EXISTS idx_transcriptions_video_lang ON transcriptions(video_id, language);