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

func TestYouTubeService_SaveChannelVideos(t *testing.T) {
	tests := []struct {
		name             string
		channelURL       string
		limit            int
		cmdRunnerSetup   func(*mockCmdRunner)
		channelRepoSetup func(*mockChannelRepository)
		videoRepoSetup   func(*mockVideoRepository)
		wantVideos       []*model.Video
		wantError        bool
		errorContains    string
	}{
		{
			name:       "successful save new videos with new channel",
			channelURL: "https://www.youtube.com/@test-channel",
			limit:      2,
			cmdRunnerSetup: func(m *mockCmdRunner) {
				// First call for FetchChannelInfo
				channelInfoResponse := `{
					"id": "video1",
					"title": "Test Video 1",
					"channel": "Test Channel",
					"channel_url": "https://www.youtube.com/@TestChannel"
				}`
				m.On("Run", mock.Anything, "yt-dlp", []string{"--dump-json", "--playlist-items", "1", "https://www.youtube.com/@test-channel"}).
					Return([]byte(channelInfoResponse), nil)

				// Second call for FetchChannelVideos
				videosResponse := `{"id": "video1", "title": "Test Video 1", "channel_id": "UC123456789", "webpage_url": "https://www.youtube.com/watch?v=video1", "duration": 300}
{"id": "video2", "title": "Test Video 2", "channel_id": "UC123456789", "webpage_url": "https://www.youtube.com/watch?v=video2", "duration": 150}`
				m.On("Run", mock.Anything, "yt-dlp", []string{"--dump-json", "--flat-playlist", "--playlist-end", "2", "https://www.youtube.com/@test-channel"}).
					Return([]byte(videosResponse), nil)
			},
			channelRepoSetup: func(m *mockChannelRepository) {
				// Channel doesn't exist, so we need to save it first
				m.On("GetByURL", mock.Anything, "https://www.youtube.com/@TestChannel").
					Return((*model.Channel)(nil), assert.AnError)
				m.On("Create", mock.Anything, mock.AnythingOfType("*model.Channel")).
					Return(nil)
			},
			videoRepoSetup: func(m *mockVideoRepository) {
				// Batch upsert videos
				m.On("UpsertBatch", mock.Anything, mock.AnythingOfType("[]*model.Video")).
					Return(nil)
			},
			wantVideos: []*model.Video{{
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
			name:       "fetch videos fails",
			channelURL: "https://invalid-url",
			limit:      5,
			cmdRunnerSetup: func(m *mockCmdRunner) {
				// First call for FetchChannelInfo fails
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(""), assert.AnError)
			},
			channelRepoSetup: func(m *mockChannelRepository) {},
			videoRepoSetup:   func(m *mockVideoRepository) {},
			wantVideos:       nil,
			wantError:        true,
			errorContains:    "failed to fetch channel info with yt-dlp",
		},
		{
			name:       "video batch creation fails",
			channelURL: "https://www.youtube.com/@test-channel",
			limit:      1,
			cmdRunnerSetup: func(m *mockCmdRunner) {
				// First call for FetchChannelInfo
				channelInfoResponse := `{
					"id": "video1",
					"title": "Test Video 1",
					"channel": "Test Channel",
					"channel_url": "https://www.youtube.com/@TestChannel"
				}`
				m.On("Run", mock.Anything, "yt-dlp", []string{"--dump-json", "--playlist-items", "1", "https://www.youtube.com/@test-channel"}).
					Return([]byte(channelInfoResponse), nil)

				// Second call for FetchChannelVideos
				videosResponse := `{"id": "video1", "title": "Test Video 1", "channel_id": "UC123456789", "webpage_url": "https://www.youtube.com/watch?v=video1", "duration": 300}`
				m.On("Run", mock.Anything, "yt-dlp", []string{"--dump-json", "--flat-playlist", "--playlist-end", "1", "https://www.youtube.com/@test-channel"}).
					Return([]byte(videosResponse), nil)
			},
			channelRepoSetup: func(m *mockChannelRepository) {
				// Channel already exists
				existingChannel := &model.Channel{
					ID:   "UC123456789",
					Name: "Test Channel",
					URL:  "https://www.youtube.com/@TestChannel",
				}
				m.On("GetByURL", mock.Anything, "https://www.youtube.com/@TestChannel").
					Return(existingChannel, nil)
			},
			videoRepoSetup: func(m *mockVideoRepository) {
				// Batch upsert videos fails
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
			mockChannelRepo := new(mockChannelRepository)
			mockVideoRepo := new(mockVideoRepository)

			tt.cmdRunnerSetup(mockRunner)
			tt.channelRepoSetup(mockChannelRepo)
			tt.videoRepoSetup(mockVideoRepo)

			service := NewYouTubeServiceWithRepositories(mockRunner, mockChannelRepo, mockVideoRepo)
			videos, err := service.SaveChannelVideos(ctx, tt.channelURL, tt.limit)

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
			mockChannelRepo.AssertExpectations(t)
			mockVideoRepo.AssertExpectations(t)
		})
	}
}

func TestYouTubeService_ListVideos(t *testing.T) {
	tests := []struct {
		name           string
		channelID      string
		limit          int
		offset         int
		videoRepoSetup func(*mockVideoRepository)
		wantVideos     []*model.Video
		wantError      bool
		errorContains  string
	}{
		{
			name:      "successful list videos",
			channelID: "UC123456789",
			limit:     10,
			offset:    0,
			videoRepoSetup: func(m *mockVideoRepository) {
				videos := []*model.Video{
					{
						ID:        "video1",
						ChannelID: "UC123456789",
						Title:     "Video 1",
						URL:       "https://www.youtube.com/watch?v=video1",
						Duration:  300,
					},
					{
						ID:        "video2",
						ChannelID: "UC123456789",
						Title:     "Video 2",
						URL:       "https://www.youtube.com/watch?v=video2",
						Duration:  180,
					},
				}
				m.On("GetByChannelID", mock.Anything, "UC123456789", 10, 0).Return(videos, nil)
			},
			wantVideos: []*model.Video{
				{
					ID:        "video1",
					ChannelID: "UC123456789",
					Title:     "Video 1",
					URL:       "https://www.youtube.com/watch?v=video1",
					Duration:  300,
				},
				{
					ID:        "video2",
					ChannelID: "UC123456789",
					Title:     "Video 2",
					URL:       "https://www.youtube.com/watch?v=video2",
					Duration:  180,
				},
			},
			wantError: false,
		},
		{
			name:           "empty channel ID",
			channelID:      "",
			limit:          10,
			offset:         0,
			videoRepoSetup: func(m *mockVideoRepository) {},
			wantVideos:     nil,
			wantError:      true,
			errorContains:  "channel ID is required",
		},
		{
			name:      "empty list",
			channelID: "UC123456789",
			limit:     10,
			offset:    0,
			videoRepoSetup: func(m *mockVideoRepository) {
				m.On("GetByChannelID", mock.Anything, "UC123456789", 10, 0).Return([]*model.Video{}, nil)
			},
			wantVideos: []*model.Video{},
			wantError:  false,
		},
		{
			name:      "repository error",
			channelID: "UC123456789",
			limit:     10,
			offset:    0,
			videoRepoSetup: func(m *mockVideoRepository) {
				m.On("GetByChannelID", mock.Anything, "UC123456789", 10, 0).Return([]*model.Video(nil), assert.AnError)
			},
			wantVideos:    nil,
			wantError:     true,
			errorContains: "failed to list videos",
		},
		{
			name:      "with pagination offset",
			channelID: "UC123456789",
			limit:     5,
			offset:    10,
			videoRepoSetup: func(m *mockVideoRepository) {
				videos := []*model.Video{
					{
						ID:        "video11",
						ChannelID: "UC123456789",
						Title:     "Video 11",
						URL:       "https://www.youtube.com/watch?v=video11",
						Duration:  240,
					},
				}
				m.On("GetByChannelID", mock.Anything, "UC123456789", 5, 10).Return(videos, nil)
			},
			wantVideos: []*model.Video{
				{
					ID:        "video11",
					ChannelID: "UC123456789",
					Title:     "Video 11",
					URL:       "https://www.youtube.com/watch?v=video11",
					Duration:  240,
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockRunner := new(mockCmdRunner)
			mockChannelRepo := new(mockChannelRepository)
			mockVideoRepo := new(mockVideoRepository)

			tt.videoRepoSetup(mockVideoRepo)

			service := NewYouTubeServiceWithRepositories(mockRunner, mockChannelRepo, mockVideoRepo)
			videos, err := service.ListVideos(ctx, tt.channelID, tt.limit, tt.offset)

			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, videos)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.wantVideos), len(videos))
				for i, want := range tt.wantVideos {
					assert.Equal(t, want.ID, videos[i].ID)
					assert.Equal(t, want.ChannelID, videos[i].ChannelID)
					assert.Equal(t, want.Title, videos[i].Title)
					assert.Equal(t, want.URL, videos[i].URL)
					assert.Equal(t, want.Duration, videos[i].Duration)
				}
			}

			mockVideoRepo.AssertExpectations(t)
		})
	}
}
