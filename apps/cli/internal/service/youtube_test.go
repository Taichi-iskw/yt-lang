package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Taichi-iskw/yt-lang/internal/model"
)

// mockCmdRunner is a mock implementation of CmdRunner for testing
type mockCmdRunner struct {
	mock.Mock
}

func (m *mockCmdRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	arguments := m.Called(ctx, name, args)
	return arguments.Get(0).([]byte), arguments.Error(1)
}

// mockChannelRepository is a mock implementation of ChannelRepository for testing
type mockChannelRepository struct {
	mock.Mock
}

func (m *mockChannelRepository) Create(ctx context.Context, channel *model.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *mockChannelRepository) GetByID(ctx context.Context, id string) (*model.Channel, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Channel), args.Error(1)
}

func (m *mockChannelRepository) GetByURL(ctx context.Context, url string) (*model.Channel, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*model.Channel), args.Error(1)
}

func (m *mockChannelRepository) Update(ctx context.Context, channel *model.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *mockChannelRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockChannelRepository) List(ctx context.Context, limit, offset int) ([]*model.Channel, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.Channel), args.Error(1)
}

// mockVideoRepository is a mock implementation of VideoRepository for testing
type mockVideoRepository struct {
	mock.Mock
}

func (m *mockVideoRepository) Create(ctx context.Context, video *model.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *mockVideoRepository) CreateBatch(ctx context.Context, videos []*model.Video) error {
	args := m.Called(ctx, videos)
	return args.Error(0)
}

func (m *mockVideoRepository) GetByID(ctx context.Context, id string) (*model.Video, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Video), args.Error(1)
}

func (m *mockVideoRepository) GetByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*model.Video, error) {
	args := m.Called(ctx, channelID, limit, offset)
	return args.Get(0).([]*model.Video), args.Error(1)
}

func (m *mockVideoRepository) Update(ctx context.Context, video *model.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *mockVideoRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockVideoRepository) List(ctx context.Context, limit, offset int) ([]*model.Video, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.Video), args.Error(1)
}

func TestYouTubeService_FetchChannelInfo(t *testing.T) {
	tests := []struct {
		name          string
		channelURL    string
		mockSetup     func(*mockCmdRunner)
		wantChannel   *model.Channel
		wantError     bool
		errorContains string
	}{
		{
			name:       "valid channel URL",
			channelURL: "https://www.youtube.com/@valid-channel",
			mockSetup: func(m *mockCmdRunner) {
				jsonResponse := `{
					"id": "123456789",
					"title": "Test Video",
					"channel": "Valid Channel",
					"channel_url": "https://www.youtube.com/@ValidChannel"
				}`
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(jsonResponse), nil)
			},
			wantChannel: &model.Channel{
				ID:   "UC123456789",
				Name: "Valid Channel",
				URL:  "https://www.youtube.com/@ValidChannel",
			},
			wantError: false,
		},
		{
			name:       "yt-dlp command fails",
			channelURL: "https://invalid-url",
			mockSetup: func(m *mockCmdRunner) {
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(""), assert.AnError)
			},
			wantChannel:   nil,
			wantError:     true,
			errorContains: "failed to fetch channel info with yt-dlp",
		},
		{
			name:          "empty channel URL",
			channelURL:    "",
			mockSetup:     func(m *mockCmdRunner) {},
			wantChannel:   nil,
			wantError:     true,
			errorContains: "channel URL is required",
		},
		{
			name:       "invalid JSON response",
			channelURL: "https://www.youtube.com/@test",
			mockSetup: func(m *mockCmdRunner) {
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte("invalid json"), nil)
			},
			wantChannel:   nil,
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
			channel, err := service.FetchChannelInfo(ctx, tt.channelURL)

			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, channel)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, channel)
				assert.Equal(t, tt.wantChannel.ID, channel.ID)
				assert.Equal(t, tt.wantChannel.Name, channel.Name)
				assert.Equal(t, tt.wantChannel.URL, channel.URL)
			}

			mockRunner.AssertExpectations(t)
		})
	}
}

func TestYouTubeService_SaveChannelInfo(t *testing.T) {
	tests := []struct {
		name             string
		channelURL       string
		cmdRunnerSetup   func(*mockCmdRunner)
		channelRepoSetup func(*mockChannelRepository)
		wantChannel      *model.Channel
		wantError        bool
		errorContains    string
	}{
		{
			name:       "successful save new channel",
			channelURL: "https://www.youtube.com/@valid-channel",
			cmdRunnerSetup: func(m *mockCmdRunner) {
				jsonResponse := `{
					"id": "123456789",
					"title": "Test Video",
					"channel": "Valid Channel",
					"channel_url": "https://www.youtube.com/@ValidChannel"
				}`
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(jsonResponse), nil)
			},
			channelRepoSetup: func(m *mockChannelRepository) {
				// First check if channel exists by URL - returns NOT_FOUND
				m.On("GetByURL", mock.Anything, "https://www.youtube.com/@ValidChannel").
					Return((*model.Channel)(nil), assert.AnError)
				// Then create the channel
				m.On("Create", mock.Anything, mock.AnythingOfType("*model.Channel")).
					Return(nil)
			},
			wantChannel: &model.Channel{
				ID:   "UC123456789",
				Name: "Valid Channel",
				URL:  "https://www.youtube.com/@ValidChannel",
			},
			wantError: false,
		},
		{
			name:       "channel already exists - return existing",
			channelURL: "https://www.youtube.com/@existing-channel",
			cmdRunnerSetup: func(m *mockCmdRunner) {
				jsonResponse := `{
					"id": "987654321",
					"title": "Test Video",
					"channel": "Existing Channel",
					"channel_url": "https://www.youtube.com/@ExistingChannel"
				}`
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(jsonResponse), nil)
			},
			channelRepoSetup: func(m *mockChannelRepository) {
				existingChannel := &model.Channel{
					ID:   "UC987654321",
					Name: "Existing Channel",
					URL:  "https://www.youtube.com/@ExistingChannel",
				}
				// Channel exists in DB
				m.On("GetByURL", mock.Anything, "https://www.youtube.com/@ExistingChannel").
					Return(existingChannel, nil)
			},
			wantChannel: &model.Channel{
				ID:   "UC987654321",
				Name: "Existing Channel",
				URL:  "https://www.youtube.com/@ExistingChannel",
			},
			wantError: false,
		},
		{
			name:       "fetch channel info fails",
			channelURL: "https://invalid-url",
			cmdRunnerSetup: func(m *mockCmdRunner) {
				m.On("Run", mock.Anything, "yt-dlp", mock.AnythingOfType("[]string")).
					Return([]byte(""), assert.AnError)
			},
			channelRepoSetup: func(m *mockChannelRepository) {},
			wantChannel:      nil,
			wantError:        true,
			errorContains:    "failed to fetch channel info with yt-dlp",
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

			service := NewYouTubeServiceWithRepositories(mockRunner, mockChannelRepo, mockVideoRepo)
			channel, err := service.SaveChannelInfo(ctx, tt.channelURL)

			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, channel)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, channel)
				assert.Equal(t, tt.wantChannel.ID, channel.ID)
				assert.Equal(t, tt.wantChannel.Name, channel.Name)
				assert.Equal(t, tt.wantChannel.URL, channel.URL)
			}

			mockRunner.AssertExpectations(t)
			mockChannelRepo.AssertExpectations(t)
			mockVideoRepo.AssertExpectations(t)
		})
	}
}

func TestYouTubeService_ListChannels(t *testing.T) {
	tests := []struct {
		name             string
		limit            int
		offset           int
		channelRepoSetup func(*mockChannelRepository)
		wantChannels     []*model.Channel
		wantError        bool
		errorContains    string
	}{
		{
			name:   "successful list channels",
			limit:  10,
			offset: 0,
			channelRepoSetup: func(m *mockChannelRepository) {
				channels := []*model.Channel{
					{
						ID:   "UC123456789",
						Name: "Channel 1",
						URL:  "https://www.youtube.com/@Channel1",
					},
					{
						ID:   "UC987654321",
						Name: "Channel 2",
						URL:  "https://www.youtube.com/@Channel2",
					},
				}
				m.On("List", mock.Anything, 10, 0).Return(channels, nil)
			},
			wantChannels: []*model.Channel{
				{
					ID:   "UC123456789",
					Name: "Channel 1",
					URL:  "https://www.youtube.com/@Channel1",
				},
				{
					ID:   "UC987654321",
					Name: "Channel 2",
					URL:  "https://www.youtube.com/@Channel2",
				},
			},
			wantError: false,
		},
		{
			name:   "empty list",
			limit:  10,
			offset: 0,
			channelRepoSetup: func(m *mockChannelRepository) {
				m.On("List", mock.Anything, 10, 0).Return([]*model.Channel{}, nil)
			},
			wantChannels: []*model.Channel{},
			wantError:    false,
		},
		{
			name:   "repository error",
			limit:  10,
			offset: 0,
			channelRepoSetup: func(m *mockChannelRepository) {
				m.On("List", mock.Anything, 10, 0).Return([]*model.Channel(nil), assert.AnError)
			},
			wantChannels:  nil,
			wantError:     true,
			errorContains: "failed to list channels",
		},
		{
			name:   "with pagination offset",
			limit:  5,
			offset: 10,
			channelRepoSetup: func(m *mockChannelRepository) {
				channels := []*model.Channel{
					{
						ID:   "UC111111111",
						Name: "Channel 11",
						URL:  "https://www.youtube.com/@Channel11",
					},
				}
				m.On("List", mock.Anything, 5, 10).Return(channels, nil)
			},
			wantChannels: []*model.Channel{
				{
					ID:   "UC111111111",
					Name: "Channel 11",
					URL:  "https://www.youtube.com/@Channel11",
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

			tt.channelRepoSetup(mockChannelRepo)

			service := NewYouTubeServiceWithRepositories(mockRunner, mockChannelRepo, mockVideoRepo)
			channels, err := service.ListChannels(ctx, tt.limit, tt.offset)

			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, channels)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.wantChannels), len(channels))
				for i, want := range tt.wantChannels {
					assert.Equal(t, want.ID, channels[i].ID)
					assert.Equal(t, want.Name, channels[i].Name)
					assert.Equal(t, want.URL, channels[i].URL)
				}
			}

			mockChannelRepo.AssertExpectations(t)
		})
	}
}

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
				// Batch create videos
				m.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*model.Video")).
					Return(nil)
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
				// Batch create videos fails
				m.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*model.Video")).
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
