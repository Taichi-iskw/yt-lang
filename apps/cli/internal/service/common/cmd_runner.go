package service

import (
	"context"
	"os/exec"
)

// CmdRunner is interface for executing external commands
type CmdRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// realCmdRunner implements CmdRunner using os/exec
type realCmdRunner struct{}

// NewCmdRunner creates a new CmdRunner
func NewCmdRunner() CmdRunner {
	return &realCmdRunner{}
}

// Run executes external command with given arguments
func (r *realCmdRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}
