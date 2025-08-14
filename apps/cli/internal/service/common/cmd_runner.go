package common

import (
	"context"
	"os"
	"os/exec"
)

// Process represents a running process
type Process interface {
	Wait() error
	Kill() error
	Signal(sig os.Signal) error
}

// CmdRunner is interface for executing external commands
type CmdRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
	Start(ctx context.Context, name string, args ...string) (Process, error)
}

// realCmdRunner implements CmdRunner using os/exec
type realCmdRunner struct{}

// NewCmdRunner creates a new CmdRunner
func NewCmdRunner() CmdRunner {
	return &realCmdRunner{}
}

// processWrapper wraps exec.Cmd to implement Process interface
type processWrapper struct {
	cmd *exec.Cmd
}

func (p *processWrapper) Wait() error {
	return p.cmd.Wait()
}

func (p *processWrapper) Kill() error {
	if p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}

func (p *processWrapper) Signal(sig os.Signal) error {
	if p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Signal(sig)
}

// Run executes external command with given arguments
func (r *realCmdRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

// Start starts external command and returns Process for management
func (r *realCmdRunner) Start(ctx context.Context, name string, args ...string) (Process, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &processWrapper{cmd: cmd}, nil
}
