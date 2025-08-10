package model

import "time"

// Channel represents YouTube channel information
type Channel struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	URL  string `json:"url" db:"url"`
}

// Video represents YouTube video information
type Video struct {
	ID        string `json:"id" db:"id"`
	ChannelID string `json:"channel_id" db:"channel_id"`
	Title     string `json:"title" db:"title"`
	URL       string `json:"url" db:"url"`
	Duration  int    `json:"duration" db:"duration"` // duration in seconds
}

// Transcription represents a subtitle segment for a video
type Transcription struct {
	ID        int       `json:"id" db:"id"`
	VideoID   string    `json:"video_id" db:"video_id"`
	StartTime float64   `json:"start_time" db:"start_time"` // Start time in seconds
	EndTime   float64   `json:"end_time" db:"end_time"`     // End time in seconds  
	Content   string    `json:"content" db:"content"`       // Text content for this time segment
	Language  string    `json:"language" db:"language"`
	Source    string    `json:"source" db:"source"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Translation represents translated transcription
type Translation struct {
	ID              int       `json:"id" db:"id"`
	TranscriptionID int       `json:"transcription_id" db:"transcription_id"`
	TargetLanguage  string    `json:"target_language" db:"target_language"`
	Content         string    `json:"content" db:"content"`
	Source          string    `json:"source" db:"source"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}
