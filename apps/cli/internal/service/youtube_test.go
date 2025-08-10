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
