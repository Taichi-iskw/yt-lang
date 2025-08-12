-- Create translations table for storing translated subtitle segments
-- Segment-based design for proper YouTube subtitle generation with timing information
CREATE TABLE IF NOT EXISTS translations (
    id SERIAL PRIMARY KEY,
    transcription_segment_id UUID NOT NULL, -- Foreign key to transcription_segments.id
    target_language VARCHAR(10) NOT NULL,   -- Target language code (e.g., 'ja', 'en')
    translated_text TEXT NOT NULL,          -- Translated text content for the specific segment
    source VARCHAR(50) NOT NULL DEFAULT 'plamo', -- Translation source: plamo, google, etc
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Keep for audit/debugging
    
    -- Foreign key constraint to transcription_segments
    CONSTRAINT fk_translations_transcription_segment_id 
        FOREIGN KEY (transcription_segment_id) 
        REFERENCES transcription_segments(id) 
        ON DELETE CASCADE,
        
    -- Unique constraint: one translation per segment per target language per source
    CONSTRAINT unique_translation_per_segment_lang_source 
        UNIQUE(transcription_segment_id, target_language, source)
);

-- Essential indexes for performance
CREATE INDEX IF NOT EXISTS idx_translations_segment_id ON translations(transcription_segment_id);
CREATE INDEX IF NOT EXISTS idx_translations_segment_lang ON translations(transcription_segment_id, target_language);