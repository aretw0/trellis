package runner

import (
	"context"
	"strings"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
)

// MockHandler for testing middleware inputs/outputs
type MockIOHandler struct {
	CapturedOutput []domain.ActionRequest
	InputBehavior  func() (string, error)
}

func (m *MockIOHandler) Output(ctx context.Context, actions []domain.ActionRequest) (bool, error) {
	m.CapturedOutput = append(m.CapturedOutput, actions...)
	return true, nil
}

func (m *MockIOHandler) Input(ctx context.Context) (string, error) {
	if m.InputBehavior != nil {
		return m.InputBehavior()
	}
	return "", nil
}

func (m *MockIOHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	return domain.ToolResult{ID: call.ID, Result: "mock_exec"}, nil
}

func (m *MockIOHandler) SystemOutput(ctx context.Context, msg string) error {
	// For testing, we capture this as a system action
	// Since Output() captures ActionRenderContent, let's capture this too
	// but maybe wrap it in an ActionRequest manually so test looks the same
	// Or we just modify the Capture logic.
	// Let's create an ActionSystemMessage for it.
	action := domain.ActionRequest{
		Type:    domain.ActionSystemMessage,
		Payload: msg,
	}
	m.CapturedOutput = append(m.CapturedOutput, action)
	return nil
}

func TestConfirmationMiddleware_Allow(t *testing.T) {
	mock := &MockIOHandler{
		InputBehavior: func() (string, error) { return "y\n", nil },
	}

	interceptor := ConfirmationMiddleware(mock)
	call := domain.ToolCall{ID: "1", Name: "delete_db"}

	allowed, _, err := interceptor(context.Background(), call)
	if err != nil {
		t.Fatalf("Middleware error: %v", err)
	}

	if !allowed {
		t.Error("Expected tool to be allowed with 'y'")
	}

	// Verify prompt was sent
	if len(mock.CapturedOutput) == 0 {
		t.Error("Expected output prompt actions")
	}
	// Basic check for content
	foundPrompt := false
	for _, act := range mock.CapturedOutput {
		if act.Type == domain.ActionSystemMessage {
			if strings.Contains(act.Payload.(string), "Allow execution?") {
				foundPrompt = true
			}
		}
	}
	if !foundPrompt {
		t.Error("Expected prompt message in output")
	}
}

func TestConfirmationMiddleware_Deny(t *testing.T) {
	mock := &MockIOHandler{
		InputBehavior: func() (string, error) { return "n\n", nil },
	}

	interceptor := ConfirmationMiddleware(mock)
	call := domain.ToolCall{ID: "1", Name: "delete_db"}

	allowed, res, err := interceptor(context.Background(), call)
	if err != nil {
		t.Fatalf("Middleware error: %v", err)
	}

	if allowed {
		t.Error("Expected tool to be denied with 'n'")
	}

	if !res.IsError || res.Error != "User denied execution by policy" {
		t.Errorf("Expected denial error, got: %v", res)
	}
}

func TestMultiInterceptor(t *testing.T) {
	// Chain: AutoApprove -> DenyAll -> AutoApprove
	// Should fail at DenyAll

	denyAll := func(ctx context.Context, call domain.ToolCall) (bool, domain.ToolResult, error) {
		return false, domain.ToolResult{Error: "Denied"}, nil
	}

	chain := MultiInterceptor(AutoApproveMiddleware(), denyAll, AutoApproveMiddleware())

	allowed, res, _ := chain(context.Background(), domain.ToolCall{})
	if allowed {
		t.Error("MultiInterceptor should stop at first denial")
	}
	if res.Error != "Denied" {
		t.Error("Expected denial result")
	}
}
