package youtube

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

func TestYouTubeService_FetchChannelVideos(t *testing.T) {
	tests := []struct {
		name          string
		channelID     string
		limit         int
		mockSetup     func(*mockCmdRunner)
		wantVideos    []*model.Video
		wantError     bool
		errorContains string
	}{
		{
			name:      "valid channel URL with videos",
			channelID: "UC123456789abcdef",
			limit:     3,
			mockSetup: func(m *mockCmdRunner) {
				jsonResponse := `{"id": "video1", "title": "Test Video 1", "channel_id": "UC123456789", "webpage_url": "https://www.youtube.com/watch?v=video1", "duration": 300.0}
{"id": "video2", "title": "Test Video 2", "channel_id": "UC123456789", "webpage_url": "https://www.youtube.com/watch?v=video2", "duration": 150.0}`
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(jsonResponse), nil)
			},
			wantVideos: []*model.Video{
				{
					ID:        "video1",
					ChannelID: "UC123456789abcdef",
					Title:     "Test Video 1",
					URL:       "https://www.youtube.com/watch?v=video1",
					Duration:  300.0,
				},
				{
					ID:        "video2",
					ChannelID: "UC123456789abcdef",
					Title:     "Test Video 2",
					URL:       "https://www.youtube.com/watch?v=video2",
					Duration:  150.0,
				},
			},
			wantError: false,
		},
		{
			name:      "duration as float64 conversion",
			channelID: "UC987654321fedcba",
			limit:     1,
			mockSetup: func(m *mockCmdRunner) {
				// yt-dlp returns duration as float64 (e.g., 214.0)
				jsonResponse := `{"id": "video1", "title": "Test Video", "channel_id": "UC123456789", "webpage_url": "https://www.youtube.com/watch?v=video1", "duration": 214.0}`
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(jsonResponse), nil)
			},
			wantVideos: []*model.Video{
				{
					ID:        "video1",
					ChannelID: "UC987654321fedcba",
					Title:     "Test Video",
					URL:       "https://www.youtube.com/watch?v=video1",
					Duration:  214.0, // Keep as float64
				},
			},
			wantError: false,
		},
		{
			name:      "channel ID with no limit",
			channelID: "UCuAXFkgsw1L7xaCfnd5JJOw",
			limit:     0, // No limit - fetch all videos
			mockSetup: func(m *mockCmdRunner) {
				// Expect channel ID to be converted to URL with no --playlist-end
				expectedArgs := []string{
					"--dump-json",
					"--flat-playlist",
					"https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
				}
				jsonResponse := `{"id": "video1", "title": "Test Video", "channel_id": "UCuAXFkgsw1L7xaCfnd5JJOw", "webpage_url": "https://www.youtube.com/watch?v=video1", "duration": 214.0}`
				m.On("Run", mock.Anything, "yt-dlp", expectedArgs).
					Return([]byte(jsonResponse), nil)
			},
			wantVideos: []*model.Video{
				{
					ID:        "video1",
					ChannelID: "UCuAXFkgsw1L7xaCfnd5JJOw",
					Title:     "Test Video",
					URL:       "https://www.youtube.com/watch?v=video1",
					Duration:  214.0,
				},
			},
			wantError: false,
		},
		{
			name:          "empty channel ID",
			channelID:     "",
			limit:         10,
			mockSetup:     func(m *mockCmdRunner) {},
			wantVideos:    nil,
			wantError:     true,
			errorContains: "channel ID is required",
		},
		{
			name:      "zero limit means fetch all videos",
			channelID: "UC111222333444555",
			limit:     0, // Should fetch all videos without limit
			mockSetup: func(m *mockCmdRunner) {
				// Expect no --playlist-end argument when limit is 0
				expectedArgs := []string{
					"--dump-json",
					"--flat-playlist",
					"https://www.youtube.com/channel/UC111222333444555",
				}
				jsonResponse := `{"id": "video1", "title": "Test Video", "channel_id": "UC123456789", "webpage_url": "https://www.youtube.com/watch?v=video1", "duration": 214.0}`
				m.On("Run", mock.Anything, "yt-dlp", expectedArgs).
					Return([]byte(jsonResponse), nil)
			},
			wantVideos: []*model.Video{
				{
					ID:        "video1",
					ChannelID: "UC111222333444555",
					Title:     "Test Video",
					URL:       "https://www.youtube.com/watch?v=video1",
					Duration:  214.0,
				},
			},
			wantError: false,
		},
		{
			name:          "invalid channel ID format",
			channelID:     "invalid-id",
			limit:         5,
			mockSetup:     func(m *mockCmdRunner) {}, // No need to set up mock as validation fails first
			wantVideos:    nil,
			wantError:     true,
			errorContains: "invalid channel ID format",
		},
		{
			name:      "yt-dlp command fails",
			channelID: "UC999888777666555",
			limit:     5,
			mockSetup: func(m *mockCmdRunner) {
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(""), assert.AnError)
			},
			wantVideos:    nil,
			wantError:     true,
			errorContains: "failed to fetch channel videos with yt-dlp",
		},
		{
			name:      "invalid JSON response",
			channelID: "UC111222333444555",
			limit:     5,
			mockSetup: func(m *mockCmdRunner) {
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte("invalid json"), nil)
			},
			wantVideos:    nil,
			wantError:     true,
			errorContains: "failed to parse yt-dlp output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockRunner := new(mockCmdRunner)
			tt.mockSetup(mockRunner)

			service := NewYouTubeServiceWithCmdRunner(mockRunner)
			videos, err := service.FetchChannelVideos(ctx, tt.channelID, tt.limit)

			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, videos)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, videos)
				assert.Equal(t, len(tt.wantVideos), len(videos))
				for i, want := range tt.wantVideos {
					assert.Equal(t, want.ID, videos[i].ID)
					assert.Equal(t, want.ChannelID, videos[i].ChannelID)
					assert.Equal(t, want.Title, videos[i].Title)
					assert.Equal(t, want.URL, videos[i].URL)
					assert.Equal(t, want.Duration, videos[i].Duration)
				}
			}

			mockRunner.AssertExpectations(t)
		})
	}
}
