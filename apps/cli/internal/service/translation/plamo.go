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

// Translate translates text using PLaMo CLI
func (s *plamoService) Translate(ctx context.Context, text string, fromLang, toLang string) (string, error) {
	// Validation
	if strings.TrimSpace(text) == "" {
		return "", errors.New("text cannot be empty")
	}

	// Supported languages validation
	supportedLangs := map[string]bool{
		"en": true,
		"ja": true,
	}
	if !supportedLangs[fromLang] || !supportedLangs[toLang] {
		return "", errors.New("unsupported language")
	}

	// Execute PLaMo CLI command
	args := []string{
		"--from", fromLang,
		"--to", toLang,
		text,
	}

	output, err := s.cmdRunner.Run(ctx, "plamo", args...)
	if err != nil {
		return "", errors.New("PLaMo CLI execution failed: " + err.Error())
	}

	return strings.TrimSpace(string(output)), nil
}
