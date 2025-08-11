-- Create translations table for storing translated subtitle segments
CREATE TABLE IF NOT EXISTS translations (
    id SERIAL PRIMARY KEY,
    transcription_id UUID NOT NULL,     -- Foreign key to transcriptions.id
    target_language VARCHAR(10) NOT NULL,  -- Target language code (e.g., 'ja', 'en')
    content TEXT NOT NULL,                 -- Translated text content for the same time segment
    source VARCHAR(50) NOT NULL DEFAULT 'plamo', -- Translation source: plamo, google, etc
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Keep for audit/debugging
    
    -- Foreign key constraint
    CONSTRAINT fk_translations_transcription_id 
        FOREIGN KEY (transcription_id) 
        REFERENCES transcriptions(id) 
        ON DELETE CASCADE,
        
    -- Unique constraint: one translation per transcription per target language per source
    CONSTRAINT unique_translation_per_transcription_lang_source 
        UNIQUE(transcription_id, target_language, source)
);

-- Essential indexes for performance
CREATE INDEX IF NOT EXISTS idx_translations_transcription_id ON translations(transcription_id);
CREATE INDEX IF NOT EXISTS idx_translations_transcription_lang ON translations(transcription_id, target_language);