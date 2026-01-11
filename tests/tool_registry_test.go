package tests

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/registry"
	"github.com/aretw0/trellis/pkg/runner"
)

func TestToolRegistryIntegration(t *testing.T) {
	// 1. Setup Registry
	reg := registry.NewRegistry()
	reg.Register("add", func(ctx context.Context, args map[string]any) (any, error) {
		a, okA := args["a"].(float64)
		b, okB := args["b"].(float64) // JSON unmarshal often produces float64 for numbers
		if !okA || !okB {
			// Fallback for int tests
			ia, okIA := args["a"].(int)
			ib, okIB := args["b"].(int)
			if okIA && okIB {
				return ia + ib, nil
			}
			return nil, nil // fail
		}
		return a + b, nil
	})

	// 2. Setup Handler with Registry
	handler := runner.NewTextHandler(nil, nil) // Discard output
	handler.Registry = reg

	// 3. Create Tool Call
	call := domain.ToolCall{
		ID:   "call_test_1",
		Name: "add",
		Args: map[string]any{"a": 10, "b": 20},
	}

	// 4. Handle Tool
	result, err := handler.HandleTool(context.Background(), call)
	if err != nil {
		t.Fatalf("HandleTool returned error: %v", err)
	}

	// 5. Verify Result
	if result.IsError {
		t.Fatalf("Tool execution failed: %s", result.Error)
	}

	if result.Result != 30 {
		t.Errorf("Expected 30, got %v", result.Result)
	}
}

func TestToolRegistryNotFound(t *testing.T) {
	reg := registry.NewRegistry()
	handler := runner.NewTextHandler(nil, nil)
	handler.Registry = reg

	call := domain.ToolCall{
		ID:   "call_test_2",
		Name: "missing_tool",
		Args: map[string]any{},
	}

	result, err := handler.HandleTool(context.Background(), call)
	if err != nil {
		t.Fatalf("HandleTool returned unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing tool, got success")
	}
}
