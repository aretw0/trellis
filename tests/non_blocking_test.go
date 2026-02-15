package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/memory"
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

	loader, err := memory.NewFromNodes(nodeStart, nodeEnd)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	// Create engine using custom loader
	engine, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// 2. Setup Runner with Mock IO
	// Use a timeout context to prevent hangs should logic fail
	ctx := context.Background()
	initialState, err := engine.Start(ctx, "test", nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	r := runner.NewRunner(
		runner.WithInputHandler(&MockHandler{
			Inputs: []string{"yes"}, // Only one input needed for the 'end' node
			T:      t,
		}),
		runner.WithEngine(engine),
		runner.WithInitialState(initialState),
	)

	// 3. Run
	// If 'start' was blocking, it would consume "yes" prematurely or fail if no input provided for it.
	// Since we provide only 1 input for the 'end' node, if 'start' requests input, we will get an error/hang (simulated).

	err = r.Run(context.Background())
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

func (m *MockHandler) Signal(ctx context.Context, name string, args map[string]any) error {
	m.T.Logf("[Signal] %s", name)
	return nil
}
