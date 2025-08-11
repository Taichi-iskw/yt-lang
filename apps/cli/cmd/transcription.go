package cmd

import (
	"github.com/Taichi-iskw/yt-lang/cmd/transcription"
)

func init() {
	// Add transcription command to root command
	rootCmd.AddCommand(transcription.NewTranscriptionCmd())
}
