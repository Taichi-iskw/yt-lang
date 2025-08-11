package translation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlamoServerService_Translate(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		fromLang string
		toLang   string
		setupMock func(*MockCmdRunner)
		want     string
		wantErr  bool
	}{
		{
			name:     "successful translation",
			text:     "Hello world",
			fromLang: "en",
			toLang:   "ja",
			setupMock: func(m *MockCmdRunner) {
				m.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
					if name == "plamo-translate" && len(args) >= 6 {
						// Simulate server startup and translation
						return []byte("こんにちは世界\n"), nil
					}
					return nil, errors.New("unexpected command")
				}
			},
			want:    "こんにちは世界",
			wantErr: false,
		},
		{
			name:     "empty text error",
			text:     "",
			fromLang: "en",
			toLang:   "ja",
			setupMock: func(m *MockCmdRunner) {},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "unsupported language error",
			text:     "Hello",
			fromLang: "invalid",
			toLang:   "ja",
			setupMock: func(m *MockCmdRunner) {},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "server startup failure",
			text:     "Hello",
			fromLang: "en",
			toLang:   "ja",
			setupMock: func(m *MockCmdRunner) {
				m.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
					return nil, errors.New("PLaMo not found")
				}
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCmdRunner := &MockCmdRunner{}
			if tt.setupMock != nil {
				tt.setupMock(mockCmdRunner)
			}
			
			service := NewPlamoServerService(mockCmdRunner)
			
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			got, err := service.Translate(ctx, tt.text, tt.fromLang, tt.toLang)
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			
			// Clean up server if started
			if serverService, ok := service.(*PlamoServerService); ok {
				serverService.StopServer()
			}
		})
	}
}

func TestPlamoServerService_ServerLifecycle(t *testing.T) {
	mockCmdRunner := &MockCmdRunner{
		RunFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			// Simulate successful server startup
			return []byte("PLaMo server started\n"), nil
		},
	}
	
	service := NewPlamoServerService(mockCmdRunner).(*PlamoServerService)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Test server startup
	err := service.StartServer(ctx)
	require.NoError(t, err)
	assert.True(t, service.serverStarted)
	
	// Test server is already started (should not error)
	err = service.StartServer(ctx)
	require.NoError(t, err)
	
	// Test server shutdown
	err = service.StopServer()
	require.NoError(t, err)
	assert.False(t, service.serverStarted)
	
	// Test stopping already stopped server (should not error)
	err = service.StopServer()
	require.NoError(t, err)
}

func TestMapLanguageToPLaMo(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"en", "English"},
		{"ja", "Japanese"},
		{"zh", "Chinese"},
		{"ko", "Korean"},
		{"es", "Spanish"},
		{"fr", "French"},
		{"de", "German"},
		{"it", "Italian"},
		{"ru", "Russian"},
		{"ar", "Arabic"},
		{"vi", "Vietnamese"},
		{"th", "Thai"},
		{"id", "Indonesian"},
		{"nl", "Dutch"},
		{"EN", "English"}, // Test case insensitive
		{"invalid", ""},   // Test unsupported language
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapLanguageToPLaMo(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}