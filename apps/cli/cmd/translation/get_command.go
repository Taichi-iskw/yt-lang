package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/service/translation"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the get translation command
func NewGetCommand(service translation.TranslationService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [TRANSLATION_ID]",
		Short: "Get a translation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			translationID := args[0]

			// Get flags
			format, _ := cmd.Flags().GetString("format")

			// Use provided service if available (for testing), otherwise create real service
			var translationService translation.TranslationService
			var cleanup func()

			if service != nil {
				translationService = service
			} else {
				// Create service using factory
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				factory := NewServiceFactory()
				var err error
				translationService, cleanup, err = factory.CreateService(ctx)
				if err != nil {
					return fmt.Errorf("failed to create translation service: %w", err)
				}
				defer cleanup()
			}

			// Get translation
			ctx := context.Background()
			translation, segments, err := translationService.GetTranslation(ctx, translationID)
			if err != nil {
				return fmt.Errorf("failed to get translation: %w", err)
			}

			// Format output
			switch format {
			case "json":
				output, err := json.MarshalIndent(translation, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format as JSON: %w", err)
				}
				cmd.Println(string(output))
			case "srt":
				// Format as SRT subtitle file
				if segments != nil && len(segments) > 0 {
					for i, seg := range segments {
						// SRT format: sequence number, timing, content, blank line
						cmd.Printf("%d\n", i+1)
						// Since we don't have timing info from parsed segments,
						// use estimated timing based on segment index
						startTime := formatSRTTime(i * 3) // 3 seconds per segment estimate
						endTime := formatSRTTime((i + 1) * 3)
						cmd.Printf("%s --> %s\n", startTime, endTime)
						cmd.Printf("%s\n\n", seg.TranslatedText)
					}
				} else {
					// Fallback: split content into lines for SRT format
					lines := strings.Split(translation.TranslatedText, "\n")
					for i, line := range lines {
						line = strings.TrimSpace(line)
						if line == "" {
							continue
						}
						cmd.Printf("%d\n", i+1)
						startTime := formatSRTTime(i * 3)
						endTime := formatSRTTime((i + 1) * 3)
						cmd.Printf("%s --> %s\n", startTime, endTime)
						cmd.Printf("%s\n\n", line)
					}
				}
			default: // text
				cmd.Printf("Translation ID: %d\n", translation.ID)
				cmd.Printf("Target Language: %s\n", translation.TargetLanguage)
				cmd.Printf("Source: %s\n", translation.Source)
				cmd.Println("\nTranslatedText:")
				if segments != nil && len(segments) > 0 {
					for _, seg := range segments {
						cmd.Printf("%s -> %s\n", seg.Text, seg.TranslatedText)
					}
				} else {
					cmd.Println(translation.TranslatedText)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().String("format", "text", "Output format (text, json, srt)")

	return cmd
}
