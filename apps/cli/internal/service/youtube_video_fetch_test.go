package service

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
		channelURL    string
		limit         int
		mockSetup     func(*mockCmdRunner)
		wantVideos    []*model.Video
		wantError     bool
		errorContains string
	}{
		{
			name:       "valid channel URL with videos",
			channelURL: "https://www.youtube.com/@valid-channel",
			limit:      3,
			mockSetup: func(m *mockCmdRunner) {
				jsonResponse := `{"id": "video1", "title": "Test Video 1", "channel_id": "UC123456789", "webpage_url": "https://www.youtube.com/watch?v=video1", "duration": 300}
{"id": "video2", "title": "Test Video 2", "channel_id": "UC123456789", "webpage_url": "https://www.youtube.com/watch?v=video2", "duration": 150}`
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(jsonResponse), nil)
			},
			wantVideos: []*model.Video{
				{
					ID:        "video1",
					ChannelID: "UC123456789",
					Title:     "Test Video 1",
					URL:       "https://www.youtube.com/watch?v=video1",
					Duration:  300,
				},
				{
					ID:        "video2",
					ChannelID: "UC123456789",
					Title:     "Test Video 2",
					URL:       "https://www.youtube.com/watch?v=video2",
					Duration:  150,
				},
			},
			wantError: false,
		},
		{
			name:          "empty channel URL",
			channelURL:    "",
			limit:         10,
			mockSetup:     func(m *mockCmdRunner) {},
			wantVideos:    nil,
			wantError:     true,
			errorContains: "channel URL is required",
		},
		{
			name:          "invalid limit",
			channelURL:    "https://www.youtube.com/@test",
			limit:         0,
			mockSetup:     func(m *mockCmdRunner) {},
			wantVideos:    nil,
			wantError:     true,
			errorContains: "limit must be greater than 0",
		},
		{
			name:       "yt-dlp command fails",
			channelURL: "https://invalid-url",
			limit:      5,
			mockSetup: func(m *mockCmdRunner) {
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(""), assert.AnError)
			},
			wantVideos:    nil,
			wantError:     true,
			errorContains: "failed to fetch channel videos with yt-dlp",
		},
		{
			name:       "invalid JSON response",
			channelURL: "https://www.youtube.com/@test",
			limit:      5,
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
			videos, err := service.FetchChannelVideos(ctx, tt.channelURL, tt.limit)

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
