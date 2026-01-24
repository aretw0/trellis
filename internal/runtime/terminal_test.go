package runtime

import (
	"context"
	"fmt"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLoader for testing Render logic
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

func TestEngine_Render_IsTerminal(t *testing.T) {
	tests := []struct {
		name           string
		nodeContent    string
		expectTerminal bool
	}{
		{
			name: "Standard Terminal (No transitions)",
			nodeContent: `{
				"id": "end",
				"type": "text"
			}`,
			expectTerminal: true,
		},
		{
			name: "Standard Transition",
			nodeContent: `{
				"id": "step1",
				"type": "text",
				"transitions": [
					{ "to": "step2" }
				]
			}`,
			expectTerminal: false,
		},
		{
			name: "Timeout Transition Only",
			nodeContent: `{
				"id": "step_timeout",
				"type": "text",
				"timeout": "5s"
			}`,
			expectTerminal: false,
		},
		{
			name: "Signal Transition Only",
			nodeContent: fmt.Sprintf(`{
				"id": "step_signal",
				"type": "text",
				"on_signal": {
					"%s": "exit"
				}
			}`, domain.SignalInterrupt),
			expectTerminal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := new(MockLoader)
			loader.On("GetNode", "current").Return([]byte(tt.nodeContent), nil)

			engine := NewEngine(loader, nil, nil)

			state := &domain.State{
				CurrentNodeID: "current",
				Context:       make(map[string]any),
			}

			_, isTerminal, err := engine.Render(context.Background(), state)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectTerminal, isTerminal, "isTerminal mismatch")
		})
	}
}
