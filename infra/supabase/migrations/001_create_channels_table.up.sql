-- Create channels table
CREATE TABLE IF NOT EXISTS channels (
    id VARCHAR(255) PRIMARY KEY,           -- YouTube channel ID (e.g., "UC123456789")
    name VARCHAR(500) NOT NULL,            -- Channel display name
    url VARCHAR(1000) NOT NULL UNIQUE      -- Channel URL from yt-dlp
);