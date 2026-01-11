package tests

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestJSONHandlerLocalRegistry(t *testing.T) {
	// Verifies that JSONHandler prefers Local Registry execution if available.

	// 1. Setup Registry
	reg := registry.NewRegistry()
	reg.Register("local_echo", func(ctx context.Context, args map[string]any) (any, error) {
		return args["msg"], nil
	})

	// 2. Setup JSON Handler
	// We pass empty buffers because we expect NO IO if local tool is found.
	var inBuf bytes.Buffer
	var outBuf bytes.Buffer
	handler := runner.NewJSONHandler(&inBuf, &outBuf)
	handler.Registry = reg

	// 3. Exec
	call := domain.ToolCall{
		ID:   "call_local",
		Name: "local_echo",
		Args: map[string]any{"msg": "hello_world"},
	}

	result, err := handler.HandleTool(context.Background(), call)
	if err != nil {
		t.Fatalf("JSONHandler failed local tool: %v", err)
	}

	// 4. Verify Immediate Result
	if result.Result != "hello_world" {
		t.Errorf("Expected hello_world, got %v", result.Result)
	}

	// 5. Verify No Output (No ActionRequest emitted)
	if outBuf.Len() > 0 {
		t.Errorf("Expected no output for local tool, got %d bytes: %s", outBuf.Len(), outBuf.String())
	}
}

func TestJSONHandlerFallbackHelper(t *testing.T) {
	// Verifies that if Tool is NOT in Registry, JSONHandler emits ActionRequest and waits for Input.

	reg := registry.NewRegistry()
	// No tools registered

	// Setup IO
	// Input Buffer simulates the "Client" responding to the tool call
	expectedResult := domain.ToolResult{
		ID:     "call_remote",
		Result: "client_response",
	}
	resultJSON, _ := json.Marshal(expectedResult)

	inBuf := bytes.NewBuffer(resultJSON) // Pre-fill input
	var outBuf bytes.Buffer

	handler := runner.NewJSONHandler(inBuf, &outBuf)
	handler.Registry = reg

	call := domain.ToolCall{
		ID:   "call_remote",
		Name: "remote_tool",
		Args: map[string]any{},
	}

	// Exec
	result, err := handler.HandleTool(context.Background(), call)
	if err != nil {
		t.Fatalf("JSONHandler failed fallback: %v", err)
	}

	// Verify Result from Client
	if result.Result != "client_response" {
		t.Errorf("Expected client_response, got %v", result.Result)
	}

	// Verify Output was emitted
	if outBuf.Len() == 0 {
		t.Error("Expected ActionRequest output for remote tool, got empty")
	}
}
