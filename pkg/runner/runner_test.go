package runner

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestRunner_Run_BasicFlow(t *testing.T) {
	// 1. Setup Engine with Memory Loader
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:      "start",
			Type:    domain.NodeTypeText,
			Content: []byte("Welcome to Trellis"),
			Transitions: []domain.Transition{
				{ToNodeID: "end"},
			},
		},
		domain.Node{
			ID:      "end",
			Type:    domain.NodeTypeText,
			Content: []byte("Goodbye"),
		},
	)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	engine, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// 2. Setup Runner with TextHandler (Inputs pre-filled via FeedInput)
	outputBuf := &bytes.Buffer{}
	handler := NewTextHandler(outputBuf)

	// Feed input for the 'start' node (since it asks for input by default in interactive mode)
	go handler.FeedInput("", nil)

	r := NewRunner(
		WithInputHandler(handler),
		WithInterceptor(AutoApproveMiddleware()),
	)

	// 3. Run in a goroutine to prevent deadlock
	done := make(chan error)
	go func() {
		_, err := r.Run(t.Context(), engine, nil)
		done <- err
	}()

	select {
	case err := <-done:
		// We expect nil OR context.Canceled/DeadlineExceeded since we removed "exit" command support
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Runner failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Runner timed out")
	}

	// 4. Verify Output
	out := outputBuf.String()
	if !strings.Contains(out, "Welcome to Trellis") {
		t.Error("Expected welcome message in output")
	}
	if !strings.Contains(out, "Goodbye") {
		t.Error("Expected goodbye message in output")
	}
}

func TestRunner_Run_Headless(t *testing.T) {
	// 1. Setup Engine
	loader, _ := memory.NewFromNodes(
		domain.Node{
			ID:      "start",
			Type:    domain.NodeTypeText,
			Content: []byte("Headless Mode"),
		},
	)
	engine, _ := trellis.New("", trellis.WithLoader(loader))

	// 2. Setup Runner (Headless)
	outBuf := &bytes.Buffer{}
	handler := NewJSONHandler(outBuf)

	r := NewRunner(
		WithInputHandler(handler),
		WithHeadless(true),
	)

	// 3. Run
	done := make(chan error)
	go func() {
		_, err := r.Run(t.Context(), engine, nil)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Runner failed: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Runner timed out")
	}

	// 4. Verify JSON Output
	out := outBuf.String()
	if !strings.Contains(out, "Headless Mode") {
		t.Errorf("Expected 'Headless Mode' in JSON output, got: %s", out)
	}
}

func TestRunner_Run_RollbackFlow(t *testing.T) {
	// Regression Test: Ensure Runner handles StatusRollingBack correctly by executing the undo tool.
	// Scenario: Step 1 (Do) -> Step 2 (Fail) -> Rollback -> Step 1 (Undo)

	// 1. Setup Engine
	loader, _ := memory.NewFromNodes(
		domain.Node{
			ID:   "start",
			Type: domain.NodeTypeTool,
			Do: &domain.ToolCall{
				ID:   "do_step1",
				Name: "do_step1",
			},
			Undo: &domain.ToolCall{
				ID:   "undo_step1",
				Name: "undo_step1",
			},
			Transitions: []domain.Transition{
				{ToNodeID: "step2"},
			},
		},
		domain.Node{
			ID:      "step2",
			Type:    domain.NodeTypeTool,
			Do:      &domain.ToolCall{Name: "fail_tool", ID: "fail_tool"},
			OnError: "rollback",
		},
	)

	// Mock Tools
	tools := map[string]func(map[string]any) (any, error){
		"do_step1": func(args map[string]any) (any, error) {
			return "done", nil
		},
		"undo_step1": func(args map[string]any) (any, error) {
			return "undone", nil
		},
		"fail_tool": func(args map[string]any) (any, error) {
			return nil, errors.New("failure") // trigger error
		},
	}

	engine, _ := trellis.New("", trellis.WithLoader(loader))

	// 2. Setup Runner with Mock Tool Handler
	// We use a custom Input Handler to intercept tool calls
	mockHandler := &MockToolHandler{
		Tools: tools,
	}

	r := NewRunner(
		WithInputHandler(mockHandler),
		WithHeadless(true),
	)

	// 3. Run
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	state, err := r.Run(ctx, engine, nil)

	// 4. Verification
	// We expect the undo tool to have been called.
	if !mockHandler.Called("undo_step1") {
		t.Fatal("Expected 'undo_step1' to be called during rollback, but it was not.")
	}

	// Check final state (should be terminated or similar, depending on rollback logic)
	// If rollback complete, it might be terminated.
	if err != nil && err != context.DeadlineExceeded {
		// It might return error if fail_tool error propagates, but here we handled it via rollback
	}
	_ = state
}

// MockToolHandler helper for tests
type MockToolHandler struct {
	Tools map[string]func(map[string]any) (any, error)
	Calls []string
}

func (m *MockToolHandler) Output(ctx context.Context, actions []domain.ActionRequest) (bool, error) {
	// No-op for output
	return false, nil
}
func (m *MockToolHandler) Input(ctx context.Context) (string, error) {
	return "", nil
}
func (m *MockToolHandler) SystemOutput(ctx context.Context, msg string) error {
	return nil
}
func (m *MockToolHandler) Signal(ctx context.Context, name string, args map[string]any) error {
	return nil
}
func (m *MockToolHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	m.Calls = append(m.Calls, call.Name)
	if fn, ok := m.Tools[call.Name]; ok {
		// Mock error for fail_tool
		if call.Name == "fail_tool" {
			return domain.ToolResult{ID: call.ID, IsError: true, Result: "failed"}, nil
		}
		res, _ := fn(call.Args)
		return domain.ToolResult{ID: call.ID, IsError: false, Result: res}, nil
	}
	return domain.ToolResult{}, nil
}
func (m *MockToolHandler) Called(name string) bool {
	for _, c := range m.Calls {
		if c == name {
			return true
		}
	}
	return false
}
