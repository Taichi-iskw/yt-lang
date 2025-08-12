package translation

import (
	"context"
	"os"

	"github.com/Taichi-iskw/yt-lang/internal/service/common"
)

// MockProcess implements common.Process for testing
type MockProcess struct {
	WaitFunc   func() error
	KillFunc   func() error
	SignalFunc func(sig os.Signal) error
}

func (m *MockProcess) Wait() error {
	if m.WaitFunc != nil {
		return m.WaitFunc()
	}
	return nil
}

func (m *MockProcess) Kill() error {
	if m.KillFunc != nil {
		return m.KillFunc()
	}
	return nil
}

func (m *MockProcess) Signal(sig os.Signal) error {
	if m.SignalFunc != nil {
		return m.SignalFunc(sig)
	}
	return nil
}

// MockCmdRunner implements common.CmdRunner for testing
type MockCmdRunner struct {
	RunFunc   func(ctx context.Context, name string, args ...string) ([]byte, error)
	StartFunc func(ctx context.Context, name string, args ...string) (common.Process, error)
}

func (m *MockCmdRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, name, args...)
	}
	return []byte("mocked output"), nil
}

func (m *MockCmdRunner) Start(ctx context.Context, name string, args ...string) (common.Process, error) {
	if m.StartFunc != nil {
		return m.StartFunc(ctx, name, args...)
	}
	// Return a default mock process
	return &MockProcess{}, nil
}
