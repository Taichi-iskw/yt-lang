package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockCmdRunner for testing
type mockWhisperCmdRunner struct {
	mock.Mock
}

func (m *mockWhisperCmdRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	argsMock := m.Called(ctx, name, args)
	if argsMock.Get(0) == nil {
		return nil, argsMock.Error(1)
	}
	return argsMock.Get(0).([]byte), argsMock.Error(1)
}

func TestWhisperService_TranscribeAudio(t *testing.T) {
	tests := []struct {
		name        string
		audioPath   string
		language    string
		setup       func(*mockWhisperCmdRunner, string)
		wantErr     bool
		checkResult func(*testing.T, *model.WhisperResult)
	}{
		{
			name:      "successful transcription",
			audioPath: "/tmp/test-audio.wav",
			language:  "auto",
			setup: func(m *mockWhisperCmdRunner, tempDir string) {
				// Create mock Whisper JSON output
				whisperOutput := model.WhisperResult{
					Text:     "Hello, this is a test video. We're learning Go.",
					Language: "en",
					Segments: []model.WhisperSegment{
						{
							ID:         0,
							Start:      0.0,
							End:        2.5,
							Text:       "Hello, this is a test video.",
							Confidence: -0.5, // Whisper uses negative log probability
						},
						{
							ID:         1,
							Start:      2.5,
							End:        6.0,
							Text:       "We're learning Go.",
							Confidence: -0.8,
						},
					},
				}

				// Create JSON file in temp directory
				outputPath := filepath.Join(tempDir, "test-audio.json")
				jsonData, _ := json.Marshal(whisperOutput)
				os.WriteFile(outputPath, jsonData, 0644)

				// Mock whisper command execution
				m.On("Run", mock.Anything, "whisper", []string{
					"/tmp/test-audio.wav",
					"--model", "large",
					"--language", "auto",
					"--output_format", "json",
					"--output_dir", tempDir,
					"--temperature", "0",
				}).Return([]byte("Whisper execution successful"), nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *model.WhisperResult) {
				assert.NotNil(t, result)
				assert.Equal(t, "Hello, this is a test video. We're learning Go.", result.Text)
				assert.Equal(t, "en", result.Language)
				assert.Len(t, result.Segments, 2)
				assert.Equal(t, "Hello, this is a test video.", result.Segments[0].Text)
				assert.Equal(t, 0.0, result.Segments[0].Start)
				assert.Equal(t, 2.5, result.Segments[0].End)
			},
		},
		{
			name:      "whisper command fails",
			audioPath: "/tmp/test-audio.wav",
			language:  "ja",
			setup: func(m *mockWhisperCmdRunner, tempDir string) {
				m.On("Run", mock.Anything, "whisper", mock.Anything).
					Return(nil, assert.AnError)
			},
			wantErr: true,
			checkResult: func(t *testing.T, result *model.WhisperResult) {
				assert.Nil(t, result)
			},
		},
		{
			name:      "invalid audio file",
			audioPath: "",
			language:  "auto",
			setup: func(m *mockWhisperCmdRunner, tempDir string) {
				// No mock setup needed, should fail validation
			},
			wantErr: true,
			checkResult: func(t *testing.T, result *model.WhisperResult) {
				assert.Nil(t, result)
			},
		},
		{
			name:      "json parsing error",
			audioPath: "/tmp/test-audio.wav",
			language:  "en",
			setup: func(m *mockWhisperCmdRunner, tempDir string) {
				// Create invalid JSON file
				outputPath := filepath.Join(tempDir, "test-audio.json")
				os.WriteFile(outputPath, []byte("invalid json"), 0644)

				m.On("Run", mock.Anything, "whisper", mock.Anything).
					Return([]byte("Whisper execution successful"), nil)
			},
			wantErr: true,
			checkResult: func(t *testing.T, result *model.WhisperResult) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test
			tempDir, err := os.MkdirTemp("", "whisper-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			mockRunner := new(mockWhisperCmdRunner)
			tt.setup(mockRunner, tempDir)

			// Override tempDir creation in service
			service := NewWhisperServiceWithCmdRunner(mockRunner, "large")

			// Use a custom context with tempDir
			ctx := context.WithValue(context.Background(), "tempDir", tempDir)

			result, err := service.TranscribeAudio(ctx, tt.audioPath, tt.language)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.checkResult(t, result)
			mockRunner.AssertExpectations(t)
		})
	}
}
