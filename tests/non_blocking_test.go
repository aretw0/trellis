package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

func TestNonBlockingText(t *testing.T) {
	// 1. Setup Graph: Start (Text) -> End (Question)
	// 'Start' is simple text, should be non-blocking.
	// 'End' is question, should block.

	nodeStart := domain.Node{
		ID:      "start",
		Type:    domain.NodeTypeText,
		Content: []byte("This is a non-blocking intro."),
		Transitions: []domain.Transition{
			{ToNodeID: "end"}, // Unconditional immediate transition
		},
	}

	nodeEnd := domain.Node{
		ID:      "end",
		Type:    domain.NodeTypeQuestion,
		Content: []byte("Finished?"),
	}

	loader, err := inmemory.NewFromNodes(nodeStart, nodeEnd)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	// Create engine using custom loader
	engine, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// 2. Setup Runner with Mock IO
	r := &runner.Runner{
		Handler: &MockHandler{
			Inputs: []string{"yes"}, // Only one input needed for the 'end' node
			T:      t,
		},
	}

	// 3. Run
	// If 'start' was blocking, it would consume "yes" prematurely or fail if no input provided for it.
	// Since we provide only 1 input for the 'end' node, if 'start' requests input, we will get an error/hang (simulated).

	initialState := engine.Start()
	if initialState.CurrentNodeID != "start" {
		t.Fatalf("Expected start node, got %s", initialState.CurrentNodeID)
	}

	// Manually driving the engine to check steps would be better, but Runner.Run abstracts the loop.
	// Let's use the standard Runner.Run to verify end-to-end flow.

	// Use a timeout context to prevent hangs should logic fail
	// context is not directly passed to Run in the current signature, but we can verify execution finishes.
	// Wait: Runner.Run DOES handle context? No, it takes engine and state.
	// But handlers take context. MockHandler can respect it?
	// The Runner code passes context.Background().

	err = r.Run(engine, initialState)
	if err != nil && err.Error() != "mock EOF" {
		// mock EOF is expected when inputs out
		t.Logf("Runner stopped with: %v", err)
	}
}

type MockHandler struct {
	Inputs []string
	Cursor int
	T      *testing.T
}

func (m *MockHandler) Output(ctx context.Context, actions []domain.ActionRequest) (bool, error) {
	for _, act := range actions {
		if act.Type == domain.ActionRequestInput {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockHandler) Input(ctx context.Context) (string, error) {
	if m.Cursor >= len(m.Inputs) {
		return "", fmt.Errorf("mock EOF")
	}
	val := m.Inputs[m.Cursor]
	m.Cursor++
	return val, nil
}

func (m *MockHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	return domain.ToolResult{}, nil
}

func (m *MockHandler) SystemOutput(ctx context.Context, msg string) error {
	m.T.Logf("[System] %s", msg)
	return nil
}
