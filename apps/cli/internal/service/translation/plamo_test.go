package translation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlamoService_Translate(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		fromLang string
		toLang   string
		want     string
		wantErr  bool
	}{
		{
			name:     "translate english to japanese",
			text:     "Hello world",
			fromLang: "en",
			toLang:   "ja",
			want:     "こんにちは世界",
			wantErr:  false,
		},
		{
			name:     "translate japanese to english",
			text:     "こんにちは",
			fromLang: "ja",
			toLang:   "en",
			want:     "Hello",
			wantErr:  false,
		},
		{
			name:     "empty text returns error",
			text:     "",
			fromLang: "en",
			toLang:   "ja",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "unsupported language returns error",
			text:     "Hello",
			fromLang: "invalid",
			toLang:   "ja",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock cmd runner that returns expected translations
			mockCmdRunner := &MockCmdRunner{
				RunFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
					if tt.wantErr {
						return nil, errors.New("mock error")
					}
					return []byte(tt.want), nil
				},
			}
			
			service := NewPlamoService(mockCmdRunner)
			
			ctx := context.Background()
			got, err := service.Translate(ctx, tt.text, tt.fromLang, tt.toLang)
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}