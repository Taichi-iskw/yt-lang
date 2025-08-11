package translation

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/service/translation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock translation service
type mockTranslationService struct {
	CreateTranslationFunc func(ctx context.Context, transcriptionID string, targetLang string) (*model.Translation, error)
	GetTranslationFunc    func(ctx context.Context, id string) (*model.Translation, []*translation.TranslationSegment, error)
}

func (m *mockTranslationService) CreateTranslation(ctx context.Context, transcriptionID string, targetLang string) (*model.Translation, error) {
	if m.CreateTranslationFunc != nil {
		return m.CreateTranslationFunc(ctx, transcriptionID, targetLang)
	}
	return nil, nil
}

func (m *mockTranslationService) GetTranslation(ctx context.Context, id string) (*model.Translation, []*translation.TranslationSegment, error) {
	if m.GetTranslationFunc != nil {
		return m.GetTranslationFunc(ctx, id)
	}
	return nil, nil, nil
}

func TestCreateCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		targetLang     string
		dryRun         bool
		setupMock      func(*mockTranslationService)
		expectedOutput string
		wantErr        bool
	}{
		{
			name:       "successful translation creation",
			args:       []string{"trans-123"},
			targetLang: "ja",
			dryRun:     false,
			setupMock: func(m *mockTranslationService) {
				m.CreateTranslationFunc = func(ctx context.Context, transcriptionID string, targetLang string) (*model.Translation, error) {
					return &model.Translation{
						ID:             1,
						TargetLanguage: "ja",
						Content:        "こんにちは世界",
						Source:         "plamo",
					}, nil
				}
			},
			expectedOutput: "Translation created successfully",
			wantErr:        false,
		},
		{
			name:       "dry run mode",
			args:       []string{"trans-123"},
			targetLang: "ja",
			dryRun:     true,
			setupMock: func(m *mockTranslationService) {
				// No actual translation should be created in dry-run
			},
			expectedOutput: "DRY RUN",
			wantErr:        false,
		},
		{
			name:       "missing transcription ID",
			args:       []string{},
			targetLang: "ja",
			dryRun:     false,
			setupMock:  func(m *mockTranslationService) {},
			wantErr:    true,
		},
		{
			name:       "translation service error",
			args:       []string{"trans-456"},
			targetLang: "en",
			dryRun:     false,
			setupMock: func(m *mockTranslationService) {
				m.CreateTranslationFunc = func(ctx context.Context, transcriptionID string, targetLang string) (*model.Translation, error) {
					return nil, errors.New("translation failed")
				}
			},
			expectedOutput: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := &mockTranslationService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			// Create command with mock
			cmd := NewCreateCommand(mockService)
			
			// Set flags
			cmd.Flags().Set("target-lang", tt.targetLang)
			if tt.dryRun {
				cmd.Flags().Set("dry-run", "true")
			}

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			
			// Set args
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, buf.String(), tt.expectedOutput)
				}
			}
		})
	}
}

func TestGetCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		format         string
		setupMock      func(*mockTranslationService)
		expectedOutput string
		wantErr        bool
	}{
		{
			name:   "get translation in text format",
			args:   []string{"1"},
			format: "text",
			setupMock: func(m *mockTranslationService) {
				m.GetTranslationFunc = func(ctx context.Context, id string) (*model.Translation, []*translation.TranslationSegment, error) {
					return &model.Translation{
						ID:             1,
						TargetLanguage: "ja",
						Content:        "こんにちは世界",
					}, []*translation.TranslationSegment{
						{Text: "Hello", TranslatedText: "こんにちは"},
						{Text: "World", TranslatedText: "世界"},
					}, nil
				}
			},
			expectedOutput: "こんにちは",
			wantErr:        false,
		},
		{
			name:   "get translation in json format",
			args:   []string{"1"},
			format: "json",
			setupMock: func(m *mockTranslationService) {
				m.GetTranslationFunc = func(ctx context.Context, id string) (*model.Translation, []*translation.TranslationSegment, error) {
					return &model.Translation{
						ID:             1,
						TargetLanguage: "ja",
						Content:        "テスト",
					}, nil, nil
				}
			},
			expectedOutput: `"target_language": "ja"`,
			wantErr:        false,
		},
		{
			name:      "missing translation ID",
			args:      []string{},
			format:    "text",
			setupMock: func(m *mockTranslationService) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := &mockTranslationService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			// Create command with mock
			cmd := NewGetCommand(mockService)
			
			// Set flags
			cmd.Flags().Set("format", tt.format)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			
			// Set args
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, buf.String(), tt.expectedOutput)
				}
			}
		})
	}
}