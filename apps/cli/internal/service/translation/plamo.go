package translation

import (
	"context"
	"errors"
	"strings"

	"github.com/Taichi-iskw/yt-lang/internal/service/common"
)

// PlamoService defines interface for PLaMo CLI translation service
type PlamoService interface {
	Translate(ctx context.Context, text string, fromLang, toLang string) (string, error)
	// Server lifecycle management (for server-mode implementations)
	StartServer(ctx context.Context) error
	StopServer() error
}

// plamoService implements PlamoService using PLaMo CLI
type plamoService struct {
	cmdRunner common.CmdRunner
}

// NewPlamoService creates a new PLaMo service instance
func NewPlamoService(cmdRunner common.CmdRunner) PlamoService {
	return &plamoService{
		cmdRunner: cmdRunner,
	}
}

// Translate translates text using PLaMo CLI (simple mode - not server)
func (s *plamoService) Translate(ctx context.Context, text string, fromLang, toLang string) (string, error) {
	// Validation
	if strings.TrimSpace(text) == "" {
		return "", errors.New("text cannot be empty")
	}

	// Map language codes to PLaMo format
	fromLangPLaMo := mapLanguageToPLaMo(fromLang)
	toLangPLaMo := mapLanguageToPLaMo(toLang)

	if fromLangPLaMo == "" || toLangPLaMo == "" {
		return "", errors.New("unsupported language")
	}

	// Execute PLaMo CLI command in simple mode (not server)
	args := []string{
		"--from", fromLangPLaMo,
		"--to", toLangPLaMo,
		"--input", text,
	}

	output, err := s.cmdRunner.Run(ctx, "plamo-translate", args...)
	if err != nil {
		return "", errors.New("PLaMo CLI execution failed: " + err.Error())
	}

	return strings.TrimSpace(string(output)), nil
}

// StartServer is a no-op for simple mode
func (s *plamoService) StartServer(ctx context.Context) error {
	// Simple mode doesn't use persistent server
	return nil
}

// StopServer is a no-op for simple mode
func (s *plamoService) StopServer() error {
	// Simple mode doesn't use persistent server
	return nil
}

// mapLanguageToPLaMo maps our language codes to PLaMo language names
func mapLanguageToPLaMo(lang string) string {
	switch strings.ToLower(lang) {
	case "en":
		return "English"
	case "ja":
		return "Japanese"
	case "zh":
		return "Chinese"
	case "ko":
		return "Korean"
	case "es":
		return "Spanish"
	case "fr":
		return "French"
	case "de":
		return "German"
	case "it":
		return "Italian"
	case "ru":
		return "Russian"
	case "ar":
		return "Arabic"
	case "vi":
		return "Vietnamese"
	case "th":
		return "Thai"
	case "id":
		return "Indonesian"
	case "nl":
		return "Dutch"
	default:
		return ""
	}
}
