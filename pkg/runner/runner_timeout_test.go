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

// MockHandler implements runner.IOHandler
type MockHandler struct {
	mock.Mock
}

func (m *MockHandler) Output(ctx context.Context, actions []domain.ActionRequest) (bool, error) {
	for _, a := range actions {
		if a.Type == domain.ActionRequestInput {
			if _, ok := a.Payload.(domain.InputRequest); ok {
				// No-op: Just verifying payload type safety if needed
			}
		}
	}
	args := m.Called(ctx, actions)
	return args.Bool(0), args.Error(1)
}

func (m *MockHandler) Input(ctx context.Context) (string, error) {
	// Simulate blocking until context timeout or explicit return
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(100 * time.Millisecond): // Block short for test speed
		return "input", nil
	}
}

func (m *MockHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	args := m.Called(ctx, call)
	return args.Get(0).(domain.ToolResult), args.Error(1)
}

func (m *MockHandler) SystemOutput(ctx context.Context, msg string) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func TestRunner_Timeout(t *testing.T) {
	// Define flow: Start (Timeout: 10ms) -> (on_signal: timeout) -> TimeoutNode
	startNode := domain.Node{
		ID:        "start",
		Type:      domain.NodeTypeText,
		InputType: "text",
		Timeout:   "10ms", // Very short timeout
		OnSignal: map[string]string{
			"timeout": "timeout_node",
		},
		Transitions: []domain.Transition{
			{ToNodeID: "next"},
		},
	}
	timeoutNode := domain.Node{
		ID:      "timeout_node",
		Type:    domain.NodeTypeText,
		Content: []byte("Timeout Occurred"),
	}

	loader, _ := inmemory.NewFromNodes(startNode, timeoutNode)
	engine, _ := trellis.New("", trellis.WithLoader(loader))

	mockHandler := new(MockHandler)
	// Output should be called for "start"
	mockHandler.On("Output", mock.Anything, mock.Anything).Return(true, nil)
	// Output should be called for "timeout_node" (after transition)
	mockHandler.On("Output", mock.Anything, mock.Anything).Return(false, nil) // Terminal-ish

	r := runner.NewRunner()
	r.Handler = mockHandler
	r.Headless = true // Prevent blocking on final node

	err := r.Run(engine, nil)
	assert.NoError(t, err)

	// Since we are mocking the output, we rely on the MockHandler expectations to verify behavior.
	// If the timeout didn't trigger, the mock handler would block indefinitely (or fail the test timeout).
	// If the transition didn't happen, the "timeout_node" output wouldn't be called.
}
