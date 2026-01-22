package runner

import (
	"context"
	"testing"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newToolCall(name string, args map[string]any) domain.ToolCall {
	return domain.ToolCall{
		ID:   "tool_1", // Default ID for test
		Name: name,
		Args: args,
	}
}

func TestRunner_HandleTool(t *testing.T) {
	tests := []struct {
		name           string
		nodeJSON       string
		toolCall       domain.ToolCall
		toolResult     domain.ToolResult
		handlerError   error
		targetNodeJSON string
		targetNodeID   string
		expectedNodeID string
		expectedError  string
	}{
		{
			name: "Tool Execution Success",
			nodeJSON: `{
				"id": "step_tool",
				"type": "tool",
				"tool_call": { "name": "calc", "args": { "op": "add" } },
				"transitions": [
					{ "condition": "", "to_node_id": "success" }
				]
			}`,
			toolCall: newToolCall("calc", map[string]any{"op": "add"}),
			toolResult: domain.ToolResult{
				ID: "tool_1", Result: "42", IsError: false,
			},
			targetNodeID:   "success",
			targetNodeJSON: `{ "id": "success", "type": "text" }`,
			expectedNodeID: "success",
		},
		{
			name: "Tool Error with Recovery (on_error)",
			nodeJSON: `{
				"id": "step_tool_fail",
				"type": "tool",
				"tool_call": { "name": "calc" },
				"on_error": "recovery_node"
			}`,
			toolCall: newToolCall("calc", nil),
			toolResult: domain.ToolResult{
				ID: "tool_1", Result: "division by zero", IsError: true, // Name is optional in Result
			},
			targetNodeID:   "recovery_node",
			targetNodeJSON: `{ "id": "recovery_node", "type": "text" }`,
			expectedNodeID: "recovery_node",
		},
		{
			name: "Tool Error Unhandled (Fail Fast)",
			nodeJSON: `{
				"id": "step_tool_fatal",
				"type": "tool",
				"tool_call": { "name": "calc" }
			}`,
			toolCall: newToolCall("calc", nil),
			toolResult: domain.ToolResult{
				ID: "tool_1", Result: "fatal error", IsError: true,
			},
			expectedError: "no 'on_error' handler is defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			handler := new(MockSignalHandler)
			runner := NewRunner(
				WithInputHandler(handler),
				WithHeadless(true),
			)

			loader := new(MockLoader)
			loader.On("GetNode", "current").Return([]byte(tt.nodeJSON), nil)
			if tt.targetNodeID != "" {
				loader.On("GetNode", tt.targetNodeID).Return([]byte(tt.targetNodeJSON), nil)
			}

			engine, err := trellis.New("", trellis.WithLoader(loader))
			assert.NoError(t, err)

			state := &domain.State{
				CurrentNodeID:   "current",
				Status:          domain.StatusWaitingForTool,
				PendingToolCall: "tool_1", // Key for matching
				Context:         make(map[string]any),
			}

			// Mock Handler
			handler.On("HandleTool", mock.Anything, tt.toolCall).Return(tt.toolResult, tt.handlerError)

			// Construct Actions manually (simulating Render output)
			actions := []domain.ActionRequest{
				{
					Type:    domain.ActionCallTool,
					Payload: tt.toolCall,
				},
			}

			// 1. Execute Tool
			input, err := runner.handleTool(
				context.Background(),
				actions,
				state,
				handler,
				func(ctx context.Context, call domain.ToolCall) (bool, domain.ToolResult, error) {
					return true, domain.ToolResult{}, nil // Allow all
				},
			)

			// 2. Assert Execution Result
			if tt.handlerError != nil {
				assert.Error(t, err)
				return // Stop if handler failed (SysCall error)
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toolResult, input)

			// 3. Navigate (Resolve Transitions / Fail Fast)
			nextState, navErr := engine.Navigate(context.Background(), state, input)

			if tt.expectedError != "" {
				if assert.Error(t, navErr) {
					t.Logf("Actual Error: %v", navErr)
					assert.Contains(t, navErr.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, navErr)
				if assert.NotNil(t, nextState) {
					assert.Equal(t, tt.expectedNodeID, nextState.CurrentNodeID)
				}
			}

			handler.AssertExpectations(t)
		})
	}
}
