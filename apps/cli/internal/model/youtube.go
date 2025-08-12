package model

import "time"

// WhisperResult represents the JSON output from Whisper CLI
type WhisperResult struct {
	Text     string           `json:"text"`
	Segments []WhisperSegment `json:"segments"`
	Language string           `json:"language"`
}

// WhisperSegment represents individual segment from Whisper output
type WhisperSegment struct {
	ID         int     `json:"id"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	Confidence float64 `json:"avg_logprob"` // Whisper uses avg_logprob for confidence
}

// Channel represents YouTube channel information
type Channel struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	URL  string `json:"url" db:"url"`
}

// Video represents YouTube video information
type Video struct {
	ID        string  `json:"id" db:"id"`
	ChannelID string  `json:"channel_id" db:"channel_id"`
	Title     string  `json:"title" db:"title"`
	URL       string  `json:"url" db:"url"`
	Duration  float64 `json:"duration" db:"duration"`
}

// Transcription represents video transcription metadata (Option B: Normalized)
type Transcription struct {
	ID               string     `json:"id" db:"id"`
	VideoID          string     `json:"video_id" db:"video_id"`
	Language         string     `json:"language" db:"language"`
	Status           string     `json:"status" db:"status"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	CompletedAt      *time.Time `json:"completed_at" db:"completed_at"`
	ErrorMessage     *string    `json:"error_message" db:"error_message"`
	DetectedLanguage *string    `json:"detected_language" db:"detected_language"`
	TotalDuration    *string    `json:"total_duration" db:"total_duration"` // INTERVAL as string
}

// TranscriptionSegment represents individual whisper segment
type TranscriptionSegment struct {
	ID              string   `json:"id" db:"id"`
	TranscriptionID string   `json:"transcription_id" db:"transcription_id"`
	SegmentIndex    int      `json:"segment_index" db:"segment_index"`
	StartTime       string   `json:"start_time" db:"start_time"` // INTERVAL as string
	EndTime         string   `json:"end_time" db:"end_time"`     // INTERVAL as string
	Text            string   `json:"text" db:"text"`
	Confidence      *float64 `json:"confidence" db:"confidence"`
}

// Translation represents translated transcription
type Translation struct {
	ID              int       `json:"id" db:"id"`                             // SERIAL PRIMARY KEY (PostgreSQL generates)
	TranscriptionID string    `json:"transcription_id" db:"transcription_id"` // UUID referencing transcriptions.id
	TargetLanguage  string    `json:"target_language" db:"target_language"`
	Content         string    `json:"content" db:"content"`
	Source          string    `json:"source" db:"source"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}
