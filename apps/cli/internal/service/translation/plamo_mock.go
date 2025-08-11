package translation

import "context"

// MockCmdRunner implements common.CmdRunner for testing
type MockCmdRunner struct {
	RunFunc func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (m *MockCmdRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, name, args...)
	}
	return []byte("mocked output"), nil
}
