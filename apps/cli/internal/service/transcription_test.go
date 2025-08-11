package service

import (
	"context"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockTranscriptionRepository for testing
type mockTranscriptionRepository struct {
	mock.Mock
}

func (m *mockTranscriptionRepository) Create(ctx context.Context, transcription *model.Transcription) error {
	args := m.Called(ctx, transcription)
	return args.Error(0)
}

func (m *mockTranscriptionRepository) GetByID(ctx context.Context, id string) (*model.Transcription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Transcription), args.Error(1)
}

func (m *mockTranscriptionRepository) GetByVideoID(ctx context.Context, videoID string) ([]*model.Transcription, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Transcription), args.Error(1)
}

func (m *mockTranscriptionRepository) GetByVideoIDAndLanguage(ctx context.Context, videoID, language string) (*model.Transcription, error) {
	args := m.Called(ctx, videoID, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Transcription), args.Error(1)
}

func (m *mockTranscriptionRepository) UpdateStatus(ctx context.Context, id string, status string, errorMessage *string) error {
	args := m.Called(ctx, id, status, errorMessage)
	return args.Error(0)
}

func (m *mockTranscriptionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// mockSegmentRepository for testing
type mockSegmentRepository struct {
	mock.Mock
}

func (m *mockSegmentRepository) CreateBatch(ctx context.Context, segments []*model.TranscriptionSegment) error {
	args := m.Called(ctx, segments)
	return args.Error(0)
}

func (m *mockSegmentRepository) GetByTranscriptionID(ctx context.Context, transcriptionID string) ([]*model.TranscriptionSegment, error) {
	args := m.Called(ctx, transcriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.TranscriptionSegment), args.Error(1)
}

func (m *mockSegmentRepository) GetByTimeRange(ctx context.Context, transcriptionID string, startTime, endTime string) ([]*model.TranscriptionSegment, error) {
	args := m.Called(ctx, transcriptionID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.TranscriptionSegment), args.Error(1)
}

func (m *mockSegmentRepository) Delete(ctx context.Context, transcriptionID string) error {
	args := m.Called(ctx, transcriptionID)
	return args.Error(0)
}

// mockWhisperService for testing
type mockWhisperService struct {
	mock.Mock
}

func (m *mockWhisperService) TranscribeAudio(ctx context.Context, audioPath string, language string) (*model.WhisperResult, error) {
	args := m.Called(ctx, audioPath, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.WhisperResult), args.Error(1)
}

func TestTranscriptionService_CreateTranscription(t *testing.T) {
	tests := []struct {
		name        string
		videoID     string
		language    string
		audioPath   string
		setupMocks  func(*mockTranscriptionRepository, *mockSegmentRepository, *mockWhisperService)
		wantErr     bool
		checkResult func(*testing.T, *model.Transcription)
	}{
		{
			name:      "successful transcription creation",
			videoID:   "test-video-123",
			language:  "auto",
			audioPath: "/tmp/audio.m4a",
			setupMocks: func(transcRepo *mockTranscriptionRepository, segRepo *mockSegmentRepository, whisperSvc *mockWhisperService) {
				// Mock: Check existing transcription (not found)
				transcRepo.On("GetByVideoIDAndLanguage", mock.Anything, "test-video-123", "auto").
					Return(nil, assert.AnError)

				// Mock: Create transcription record
				transcRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Transcription")).
					Return(nil)

				// Mock: Whisper transcription
				whisperResult := &model.WhisperResult{
					Text:     "Hello, this is a test.",
					Language: "en",
					Segments: []model.WhisperSegment{
						{
							ID:         0,
							Start:      0.0,
							End:        2.5,
							Text:       "Hello, this is a test.",
							Confidence: -0.5,
						},
					},
				}
				whisperSvc.On("TranscribeAudio", mock.Anything, "/tmp/audio.m4a", "auto").
					Return(whisperResult, nil)

				// Mock: Create segments
				segRepo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*model.TranscriptionSegment")).
					Return(nil)

				// Mock: Update transcription status to completed
				transcRepo.On("UpdateStatus", mock.Anything, mock.AnythingOfType("string"), "completed", (*string)(nil)).
					Return(nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *model.Transcription) {
				assert.NotNil(t, result)
				assert.Equal(t, "test-video-123", result.VideoID)
				assert.Equal(t, "auto", result.Language)
				assert.Equal(t, "completed", result.Status)
			},
		},
		{
			name:      "transcription already exists",
			videoID:   "test-video-123",
			language:  "en",
			audioPath: "/tmp/audio.m4a",
			setupMocks: func(transcRepo *mockTranscriptionRepository, segRepo *mockSegmentRepository, whisperSvc *mockWhisperService) {
				// Mock: Existing transcription found
				existing := &model.Transcription{
					ID:       "existing-123",
					VideoID:  "test-video-123",
					Language: "en",
					Status:   "completed",
				}
				transcRepo.On("GetByVideoIDAndLanguage", mock.Anything, "test-video-123", "en").
					Return(existing, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *model.Transcription) {
				assert.NotNil(t, result)
				assert.Equal(t, "existing-123", result.ID)
				assert.Equal(t, "completed", result.Status)
			},
		},
		{
			name:      "whisper transcription fails",
			videoID:   "test-video-123",
			language:  "ja",
			audioPath: "/tmp/audio.m4a",
			setupMocks: func(transcRepo *mockTranscriptionRepository, segRepo *mockSegmentRepository, whisperSvc *mockWhisperService) {
				// Mock: Check existing transcription (not found)
				transcRepo.On("GetByVideoIDAndLanguage", mock.Anything, "test-video-123", "ja").
					Return(nil, assert.AnError)

				// Mock: Create transcription record
				transcRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Transcription")).
					Return(nil)

				// Mock: Whisper transcription fails
				whisperSvc.On("TranscribeAudio", mock.Anything, "/tmp/audio.m4a", "ja").
					Return(nil, assert.AnError)

				// Mock: Update transcription status to failed
				errorMsg := "whisper transcription failed"
				transcRepo.On("UpdateStatus", mock.Anything, mock.AnythingOfType("string"), "failed", &errorMsg).
					Return(nil)
			},
			wantErr: true,
			checkResult: func(t *testing.T, result *model.Transcription) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transcRepo := new(mockTranscriptionRepository)
			segRepo := new(mockSegmentRepository)
			whisperSvc := new(mockWhisperService)

			tt.setupMocks(transcRepo, segRepo, whisperSvc)

			service := NewTranscriptionServiceWithDependencies(transcRepo, segRepo, whisperSvc)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := service.CreateTranscription(ctx, tt.videoID, tt.language, tt.audioPath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.checkResult(t, result)

			transcRepo.AssertExpectations(t)
			segRepo.AssertExpectations(t)
			whisperSvc.AssertExpectations(t)
		})
	}
}

func TestTranscriptionService_GetTranscription(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		setupMocks  func(*mockTranscriptionRepository, *mockSegmentRepository)
		wantErr     bool
		checkResult func(*testing.T, *model.Transcription, []*model.TranscriptionSegment)
	}{
		{
			name: "successful get transcription with segments",
			id:   "transcription-123",
			setupMocks: func(transcRepo *mockTranscriptionRepository, segRepo *mockSegmentRepository) {
				transcription := &model.Transcription{
					ID:               "transcription-123",
					VideoID:          "video-123",
					Language:         "en",
					Status:           "completed",
					DetectedLanguage: stringPtr("en"),
				}
				transcRepo.On("GetByID", mock.Anything, "transcription-123").
					Return(transcription, nil)

				segments := []*model.TranscriptionSegment{
					{
						ID:              "seg-1",
						TranscriptionID: "transcription-123",
						SegmentIndex:    0,
						StartTime:       "00:00:00",
						EndTime:         "00:00:02.5",
						Text:            "Hello, this is a test.",
					},
				}
				segRepo.On("GetByTranscriptionID", mock.Anything, "transcription-123").
					Return(segments, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, transcription *model.Transcription, segments []*model.TranscriptionSegment) {
				assert.NotNil(t, transcription)
				assert.Equal(t, "transcription-123", transcription.ID)
				assert.Equal(t, "completed", transcription.Status)
				assert.Len(t, segments, 1)
				assert.Equal(t, "Hello, this is a test.", segments[0].Text)
			},
		},
		{
			name: "transcription not found",
			id:   "nonexistent-123",
			setupMocks: func(transcRepo *mockTranscriptionRepository, segRepo *mockSegmentRepository) {
				transcRepo.On("GetByID", mock.Anything, "nonexistent-123").
					Return(nil, assert.AnError)
			},
			wantErr: true,
			checkResult: func(t *testing.T, transcription *model.Transcription, segments []*model.TranscriptionSegment) {
				assert.Nil(t, transcription)
				assert.Nil(t, segments)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transcRepo := new(mockTranscriptionRepository)
			segRepo := new(mockSegmentRepository)
			whisperSvc := new(mockWhisperService)

			tt.setupMocks(transcRepo, segRepo)

			service := NewTranscriptionServiceWithDependencies(transcRepo, segRepo, whisperSvc)

			ctx := context.Background()
			transcription, segments, err := service.GetTranscription(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.checkResult(t, transcription, segments)

			transcRepo.AssertExpectations(t)
			segRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
