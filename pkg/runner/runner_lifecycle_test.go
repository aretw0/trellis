package runner_test

import (
	"context"
	"testing"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// LifecycleMockHandler tracks calls for lifecycle verification
type LifecycleMockHandler struct {
	mock.Mock
}

func (m *LifecycleMockHandler) Output(ctx context.Context, actions []domain.ActionRequest) (bool, error) {
	args := m.Called(ctx, actions)
	return args.Bool(0), args.Error(1)
}

func (m *LifecycleMockHandler) Input(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	// Return a newline to simulate "Enter"
	return "\n", args.Error(1)
}

func (m *LifecycleMockHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	args := m.Called(ctx, call)
	return args.Get(0).(domain.ToolResult), args.Error(1)
}

func (m *LifecycleMockHandler) SystemOutput(ctx context.Context, msg string) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func TestRunner_Lifecycle_Interactive_WaitsOnTerminal(t *testing.T) {
	// A single terminal node
	termNode := domain.Node{
		ID:      "end",
		Type:    domain.NodeTypeText,
		Content: []byte("Goodbye"),
	}

	loader, _ := inmemory.NewFromNodes(termNode)
	engine, _ := trellis.New("", trellis.WithLoader(loader))

	mockHandler := new(LifecycleMockHandler)
	// Output called for printing "Goodbye"
	mockHandler.On("Output", mock.Anything, mock.Anything).Return(false, nil)

	// Input MUST be called for terminal node in interactive mode
	mockHandler.On("Input", mock.Anything).Return("\n", nil)

	r := runner.NewRunner()
	r.Handler = mockHandler
	r.Headless = false // Interactive

	// Start at "end"
	initialState := &domain.State{CurrentNodeID: "end"}

	// Execute
	// Use a timeout to ensure it doesn't hang if logic is broken
	done := make(chan error)
	go func() {
		done <- r.Run(engine, initialState)
	}()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("Runner hung")
	}

	mockHandler.AssertExpectations(t)
}

func TestRunner_Lifecycle_Headless_SkipWaitOnTerminal(t *testing.T) {
	// A single terminal node
	termNode := domain.Node{
		ID:      "end",
		Type:    domain.NodeTypeText,
		Content: []byte("Goodbye"),
	}

	loader, _ := inmemory.NewFromNodes(termNode)
	engine, _ := trellis.New("", trellis.WithLoader(loader))

	mockHandler := new(LifecycleMockHandler)
	// Output called
	mockHandler.On("Output", mock.Anything, mock.Anything).Return(false, nil)

	// Input should NOT be called
	// We do not set expectation for Input. strict mock will fail if called.

	r := runner.NewRunner()
	r.Handler = mockHandler
	r.Headless = true // Headless

	initialState := &domain.State{CurrentNodeID: "end"}

	err := r.Run(engine, initialState)
	assert.NoError(t, err)

	mockHandler.AssertExpectations(t)
}
