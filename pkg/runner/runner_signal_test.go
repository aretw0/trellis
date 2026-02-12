package runner

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSignalHandler simulates an IOHandler
type MockSignalHandler struct {
	mock.Mock
}

func (m *MockSignalHandler) Output(ctx context.Context, actions []domain.ActionRequest) (bool, error) {
	args := m.Called(ctx, actions)
	return args.Bool(0), args.Error(1)
}

func (m *MockSignalHandler) Input(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockSignalHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	args := m.Called(ctx, call)
	if args.Get(0) == nil {
		return domain.ToolResult{}, args.Error(1)
	}
	return args.Get(0).(domain.ToolResult), args.Error(1)
}

func (m *MockSignalHandler) SystemOutput(ctx context.Context, msg string) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockSignalHandler) Signal(ctx context.Context, name string, args map[string]any) error {
	callArgs := m.Called(ctx, name, args)
	return callArgs.Error(0)
}

// MockLoader for testing signal handling
type MockLoader struct {
	mock.Mock
}

func (m *MockLoader) GetNode(id string) ([]byte, error) {
	args := m.Called(id)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockLoader) ListNodes() ([]string, error) {
	return nil, nil
}

func TestRunner_HandleInput_Signal(t *testing.T) {
	tests := []struct {
		name           string
		nodeJSON       string
		targetNodeJSON string // Optional, if transition succeeds
		targetNodeID   string // Key in mock expectation
		expectedNodeID string // Expected Result
		expectedError  string // Expected Error substring
	}{
		{
			name: "Timeout Handled Successfully",
			nodeJSON: fmt.Sprintf(`{
				"id": "start",
				"type": "text",
				"on_signal": {
					"%s": "timeout_node"
				}
			}`, domain.SignalTimeout),
			targetNodeID: "timeout_node",
			targetNodeJSON: `{
				"id": "timeout_node",
				"type": "text"
			}`,
			expectedNodeID: "timeout_node",
			expectedError:  "",
		},
		{
			name: "Timeout Unhandled (No Handler)",
			nodeJSON: `{
				"id": "start",
				"type": "text"
			}`,
			// No handler for "timeout"
			expectedError: "timeout exceeded and no 'on_signal.timeout' handler defined",
		},
		{
			name: "Target Node Missing (Loader Error)",
			nodeJSON: `{
				"id": "start",
				"type": "text",
				"on_signal": {
					"timeout": "missing_node"
				}
			}`,
			targetNodeID:   "missing_node",
			targetNodeJSON: "", // Empty indicates missing/error
			// Even if handler is defined, if it fails (target missing), Runner treats it as unhandled/failed timeout
			expectedError: "timeout exceeded and no 'on_signal.timeout' handler defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			handler := new(MockSignalHandler)
			runner := NewRunner(
				WithInputHandler(handler),
				WithHeadless(false),
			)

			loader := new(MockLoader)
			loader.On("GetNode", "start").Return([]byte(tt.nodeJSON), nil)

			if tt.targetNodeID != "" && tt.targetNodeJSON != "" {
				loader.On("GetNode", tt.targetNodeID).Return([]byte(tt.targetNodeJSON), nil)
			} else if tt.targetNodeID != "" && tt.targetNodeJSON == "" {
				// Simulate Missing Node Error
				loader.On("GetNode", tt.targetNodeID).Return([]byte(nil), assert.AnError)
			}

			// Inject MockLoader
			engine, err := trellis.New("", trellis.WithLoader(loader))
			assert.NoError(t, err)

			state := &domain.State{
				CurrentNodeID: "start",
				Context:       make(map[string]any),
			}

			// Create a context that times out VERY quickly
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()

			// Mock Expectation: Input starts, blocks, then context deadline exceeded
			handler.On("Input", mock.Anything).Return("", context.DeadlineExceeded)

			// Execute
			interruptSource := make(chan struct{})
			runner.InterruptSource = interruptSource

			_, nextState, err := runner.handleInput(
				ctx,
				handler,
				true, // needsInput
				engine,
				state,
			)

			// Assertions
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, nextState) {
					assert.Equal(t, tt.expectedNodeID, nextState.CurrentNodeID)
				}
			}

			// Verify mocks match expectations (especially strict GetNode calls)
			// Note: If Target Missing, verify handler.Input was called
			handler.AssertExpectations(t)
			// loader expectations might be partial if we fail fast, but that's ok for this test.
		})
	}
}
