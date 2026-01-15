package runtime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
)

// MockLoader implements ports.GraphLoader for testing
type MockLoader struct {
	Nodes map[string][]byte
}

func (m *MockLoader) GetNode(id string) ([]byte, error) {
	if content, ok := m.Nodes[id]; ok {
		return content, nil
	}
	return nil, fmt.Errorf("node not found: %s", id)
}

func (m *MockLoader) ListNodes() ([]string, error) {
	keys := make([]string, 0, len(m.Nodes))
	for k := range m.Nodes {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestEngine_Signal(t *testing.T) {
	// Setup with local mock
	loader := &MockLoader{
		Nodes: make(map[string][]byte),
	}

	startNodeRaw := `
{
	"id": "start",
	"type": "text",
	"wait": true,
	"on_signal": {
		"interrupt": "cancel_node"
	},
	"content": "SGVsbG8=" 
}`
	// content is []byte, so JSON unmarshal expects base64 string if mapped to []byte?
	// domain.Node.Content is []byte. json.Unmarshal decodes base64 string to []byte.
	// "SGVsbG8=" is "Hello".
	loader.Nodes["start"] = []byte(startNodeRaw)

	cancelNodeRaw := `
{
	"id": "cancel_node",
	"type": "text",
	"content": "Q2FuY2VsZWQ="
}`
	loader.Nodes["cancel_node"] = []byte(cancelNodeRaw)

	// Capture hooks
	leaveCalled := false
	hooks := domain.LifecycleHooks{
		OnNodeLeave: func(ctx context.Context, e *domain.NodeEvent) {
			if e.NodeID == "start" {
				leaveCalled = true
			}
		},
	}

	// Engine handles parser creation internally
	engine := runtime.NewEngine(loader, nil, nil, runtime.WithLifecycleHooks(hooks))

	// Start state
	state := domain.NewState("start")
	state.Context["foo"] = "bar" // Test context preservation

	// Execute Signal
	nextState, err := engine.Signal(context.Background(), state, "interrupt")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, nextState)
	assert.Equal(t, "cancel_node", nextState.CurrentNodeID)
	assert.Equal(t, "bar", nextState.Context["foo"], "Context should be preserved")
	assert.True(t, leaveCalled, "OnNodeLeave should be triggered for interrupting node")
}

func TestEngine_Signal_Unhandled(t *testing.T) {
	loader := &MockLoader{
		Nodes: make(map[string][]byte),
	}

	startNodeRaw := `
{
	"id": "start",
	"type": "text",
	"content": "Q29udGVudA=="
}`
	loader.Nodes["start"] = []byte(startNodeRaw)

	// Capture hooks
	leaveCalled := false
	hooks := domain.LifecycleHooks{
		OnNodeLeave: func(ctx context.Context, e *domain.NodeEvent) {
			if e.NodeID == "start" {
				leaveCalled = true
			}
		},
	}

	engine := runtime.NewEngine(loader, nil, nil, runtime.WithLifecycleHooks(hooks))
	state := domain.NewState("start")

	// Execute Signal
	_, err := engine.Signal(context.Background(), state, "interrupt")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnhandledSignal, err)
	assert.True(t, leaveCalled, "OnNodeLeave should be triggered even for unhandled signal (graceful exit logging)")
}
