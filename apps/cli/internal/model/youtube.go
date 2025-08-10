package model

// Channel represents YouTube channel information
type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Video represents YouTube video information
type Video struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Duration  int    `json:"duration"`
}
