package translation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Taichi-iskw/yt-lang/internal/service/common"
)

// PlamoServerService implements PlamoService using PLaMo server mode
type PlamoServerService struct {
	cmdRunner     common.CmdRunner
	serverProcess common.Process
	serverStarted bool
	cancel        context.CancelFunc
	mu            sync.Mutex
}

// NewPlamoServerService creates a new PLaMo server service
func NewPlamoServerService(cmdRunner common.CmdRunner) PlamoService {
	return &PlamoServerService{
		cmdRunner: cmdRunner,
	}
}

// StartServer starts the PLaMo server if not already running
func (s *PlamoServerService) StartServer(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.serverStarted {
		return nil
	}

	// Create cancellable context for server
	serverCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// Start PLaMo server with default settings
	args := []string{
		"server",
		"--backend-type", "mlx", // Use MLX backend for Apple Silicon
		"--precision", "4bit", // Use 4bit precision for speed
		"--no-stream",   // Batch processing mode
		"--interactive", // Interactive mode for continuous translation
	}

	// Use CmdRunner to start server process
	process, err := s.cmdRunner.Start(serverCtx, "plamo-translate", args...)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to start PLaMo server: %w", err)
	}

	s.serverProcess = process
	s.serverStarted = true

	// Wait a moment to ensure server started
	time.Sleep(2 * time.Second)

	return nil
}

// StopServer stops the PLaMo server
func (s *PlamoServerService) StopServer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.serverStarted {
		return nil
	}

	// Cancel the server context first for graceful shutdown
	if s.cancel != nil {
		s.cancel()
	}

	// If we have a server process, wait for it to shutdown gracefully
	if s.serverProcess != nil {
		done := make(chan error, 1)
		go func() {
			done <- s.serverProcess.Wait()
		}()

		select {
		case <-time.After(5 * time.Second):
			// Force kill if graceful shutdown takes too long
			if err := s.serverProcess.Kill(); err != nil {
				return fmt.Errorf("failed to force kill PLaMo server: %w", err)
			}
			<-done // Wait for Wait() to return
			return fmt.Errorf("PLaMo server killed after timeout")
		case err := <-done:
			// Server shutdown completed
			if err != nil && !strings.Contains(err.Error(), "context canceled") {
				return fmt.Errorf("PLaMo server stopped with error: %w", err)
			}
		}
	}

	s.serverStarted = false
	s.serverProcess = nil
	s.cancel = nil
	return nil
}

// Translate translates text using PLaMo server
func (s *PlamoServerService) Translate(ctx context.Context, text string, fromLang, toLang string) (string, error) {
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

	// Start server if not running
	if !s.serverStarted {
		if err := s.StartServer(ctx); err != nil {
			return "", fmt.Errorf("failed to start PLaMo server: %w", err)
		}
	}

	// In server mode, we would send commands to the running server
	// For now, simulate by using the command runner directly
	args := []string{
		"--from", fromLangPLaMo,
		"--to", toLangPLaMo,
		"--input", text,
	}

	output, err := s.cmdRunner.Run(ctx, "plamo-translate", args...)
	if err != nil {
		return "", fmt.Errorf("PLaMo server translation failed: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", errors.New("empty response from PLaMo server")
	}

	return result, nil
}

// mapLanguageToPLaMo function is defined in plamo.go
