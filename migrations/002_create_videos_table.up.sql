-- Create videos table
CREATE TABLE IF NOT EXISTS videos (
    id VARCHAR(255) PRIMARY KEY,           -- YouTube video ID (e.g., "dQw4w9WgXcQ")
    channel_id VARCHAR(255) NOT NULL,      -- Foreign key to channels.id
    title VARCHAR(1000) NOT NULL,          -- Video title
    url VARCHAR(1000) NOT NULL UNIQUE,     -- Video URL
    duration REAL DEFAULT 0,               -- Duration in seconds (float64 for precision)
    
    -- Foreign key constraint
    CONSTRAINT fk_videos_channel_id 
        FOREIGN KEY (channel_id) 
        REFERENCES channels(id) 
        ON DELETE CASCADE
);

-- Create index for foreign key lookups (recommended for performance)
CREATE INDEX IF NOT EXISTS idx_videos_channel_id ON videos(channel_id);