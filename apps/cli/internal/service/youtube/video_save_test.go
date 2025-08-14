package youtube

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

func TestYouTubeService_SaveChannelVideosWithChannelID(t *testing.T) {
	tests := []struct {
		name           string
		channelID      string
		limit          int
		cmdRunnerSetup func(*mockCmdRunner)
		videoRepoSetup func(*mockVideoRepository)
		wantVideos     []*model.Video
		wantError      bool
		errorContains  string
	}{
		{
			name:      "successful save videos with channel ID",
			channelID: "UC123456789abcdef",
			limit:     2,
			cmdRunnerSetup: func(m *mockCmdRunner) {
				videosResponse := `{"id": "video1", "title": "Test Video 1", "channel_id": "UC123456789abcdef", "webpage_url": "https://www.youtube.com/watch?v=video1", "duration": 300.0}
{"id": "video2", "title": "Test Video 2", "channel_id": "UC123456789abcdef", "webpage_url": "https://www.youtube.com/watch?v=video2", "duration": 150.0}`
				m.On("Run", mock.Anything, "yt-dlp", []string{"--dump-json", "--flat-playlist", "--playlist-end", "2", "https://www.youtube.com/channel/UC123456789abcdef"}).
					Return([]byte(videosResponse), nil)
			},
			videoRepoSetup: func(m *mockVideoRepository) {
				m.On("UpsertBatch", mock.Anything, mock.AnythingOfType("[]*model.Video")).
					Return(nil)
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
			name:           "invalid channel ID format",
			channelID:      "invalid-format",
			limit:          2,
			cmdRunnerSetup: func(m *mockCmdRunner) {}, // No need to set up mock as validation fails first
			videoRepoSetup: func(m *mockVideoRepository) {},
			wantVideos:     nil,
			wantError:      true,
			errorContains:  "invalid channel ID format",
		},
		{
			name:      "fetch videos fails",
			channelID: "UC987654321fedcba",
			limit:     2,
			cmdRunnerSetup: func(m *mockCmdRunner) {
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(""), assert.AnError)
			},
			videoRepoSetup: func(m *mockVideoRepository) {},
			wantVideos:     nil,
			wantError:      true,
			errorContains:  "failed to fetch channel videos",
		},
		{
			name:      "video batch save fails",
			channelID: "UC555666777888999",
			limit:     1,
			cmdRunnerSetup: func(m *mockCmdRunner) {
				videosResponse := `{"id": "video1", "title": "Test Video", "channel_id": "UC555666777888999", "webpage_url": "https://www.youtube.com/watch?v=video1", "duration": 300.0}`
				m.On("Run", mock.Anything, "yt-dlp", []string{"--dump-json", "--flat-playlist", "--playlist-end", "1", "https://www.youtube.com/channel/UC555666777888999"}).
					Return([]byte(videosResponse), nil)
			},
			videoRepoSetup: func(m *mockVideoRepository) {
				m.On("UpsertBatch", mock.Anything, mock.AnythingOfType("[]*model.Video")).
					Return(assert.AnError)
			},
			wantVideos:    nil,
			wantError:     true,
			errorContains: "failed to save videos to database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockRunner := new(mockCmdRunner)
			mockVideoRepo := new(mockVideoRepository)

			tt.cmdRunnerSetup(mockRunner)
			tt.videoRepoSetup(mockVideoRepo)

			service := NewYouTubeServiceWithRepositories(mockRunner, nil, mockVideoRepo)
			videos, err := service.SaveChannelVideos(ctx, tt.channelID, tt.limit)

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
			mockVideoRepo.AssertExpectations(t)
		})
	}
}
