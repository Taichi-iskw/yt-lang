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
