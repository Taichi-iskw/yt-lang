package translation

import (
	"strings"
	"testing"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextFormatter(t *testing.T) {
	formatter := &TextFormatter{}

	trans := &model.Translation{
		ID:             1,
		TargetLanguage: "ja",
		TranslatedText: "こんにちは世界",
		Source:         "plamo",
		CreatedAt:      time.Now(),
	}

	segments := []*TranslationSegment{
		{Text: "Hello", TranslatedText: "こんにちは"},
		{Text: "World", TranslatedText: "世界"},
	}

	output, err := formatter.Format(trans, segments)
	require.NoError(t, err)

	assert.Contains(t, output, "Translation ID: 1")
	assert.Contains(t, output, "Target Language: ja")
	assert.Contains(t, output, "Source: plamo")
	assert.Contains(t, output, "[1] Hello")
	assert.Contains(t, output, "→ こんにちは")
	assert.Contains(t, output, "[2] World")
	assert.Contains(t, output, "→ 世界")
}

func TestJSONFormatter(t *testing.T) {
	formatter := &JSONFormatter{}

	trans := &model.Translation{
		ID:             1,
		TargetLanguage: "ja",
		TranslatedText: "テスト",
		Source:         "plamo",
		CreatedAt:      time.Now(),
	}

	output, err := formatter.Format(trans, nil)
	require.NoError(t, err)

	assert.Contains(t, output, `"id": 1`)
	assert.Contains(t, output, `"target_language": "ja"`)
	assert.Contains(t, output, `"translated_text": "テスト"`)
	assert.Contains(t, output, `"source": "plamo"`)
}

func TestSRTFormatter(t *testing.T) {
	formatter := &SRTFormatter{}

	trans := &model.Translation{
		ID:             1,
		TargetLanguage: "ja",
		TranslatedText: "こんにちは世界",
		Source:         "plamo",
	}

	t.Run("with segments", func(t *testing.T) {
		segments := []*TranslationSegment{
			{Text: "Hello", TranslatedText: "こんにちは"},
			{Text: "World", TranslatedText: "世界"},
		}

		output, err := formatter.Format(trans, segments)
		require.NoError(t, err)

		lines := strings.Split(output, "\n")
		assert.Equal(t, "1", lines[0])
		assert.Contains(t, lines[1], "-->")
		assert.Equal(t, "こんにちは", lines[2])

		assert.Equal(t, "2", lines[4])
		assert.Contains(t, lines[5], "-->")
		assert.Equal(t, "世界", lines[6])
	})

	t.Run("without segments", func(t *testing.T) {
		_, err := formatter.Format(trans, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SRT format requires segments")
	})
}

func TestGetFormatter(t *testing.T) {
	tests := []struct {
		format       string
		expectedType string
		wantErr      bool
	}{
		{"text", "*translation.TextFormatter", false},
		{"txt", "*translation.TextFormatter", false},
		{"json", "*translation.JSONFormatter", false},
		{"srt", "*translation.SRTFormatter", false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			formatter, err := GetFormatter(tt.format)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, formatter)
			}
		})
	}
}
