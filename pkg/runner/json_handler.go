package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/registry"
)

// JSONHandler implements the IOHandler interface for structured JSON-Lines communication.
type JSONHandler struct {
	Reader   *bufio.Reader
	Writer   io.Writer
	Encoder  *json.Encoder
	Decoder  *json.Decoder
	Registry *registry.Registry
}

// NewJSONHandler creates a handler for JSON IO.
func NewJSONHandler(r io.Reader, w io.Writer) *JSONHandler {
	if r == nil {
		r = os.Stdin
	}
	if w == nil {
		w = os.Stdout
	}
	return &JSONHandler{
		Reader:  bufio.NewReader(r),
		Writer:  w,
		Encoder: json.NewEncoder(w),
		Decoder: json.NewDecoder(r),
	}
}

func (h *JSONHandler) Output(ctx context.Context, actions []domain.ActionRequest) (bool, error) {
	if len(actions) == 0 {
		return false, nil
	}

	// Emit actions as a single JSON line
	if err := h.Encoder.Encode(actions); err != nil {
		return false, err
	}

	// Check if the engine is requesting input
	needsInput := false
	for _, act := range actions {
		if act.Type == domain.ActionRequestInput {
			needsInput = true
		}
	}

	return needsInput, nil
}

func (h *JSONHandler) Input(ctx context.Context) (string, error) {
	// Read a line of JSON (or plain text)
	// We expect either a JSON string "value" or a raw string value.
	// Complex JSON objects are not yet strictly validated/parsed into a struct,
	// but are essentially treated as strings for now.

	// Reading line-based
	text, err := h.Reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	text = strings.TrimSpace(text)

	// Try to unquote if it's a JSON string
	var val string
	if err := json.Unmarshal([]byte(text), &val); err == nil {
		return val, nil
	}

	// Fallback: return raw text (e.g. if they just sent plain text)
	return text, nil
}

// HandleTool for JSONHandler emits the tool call as JSON.
// In a real headless/JSON scenario, the Host should intercept this ActionRequest in the 'Output' phase
// and perform the action, or the Runner should handle this differently.
// For now, to satisfy the interface, we log it or return a mock if needed.
// Ideally, the JSON Runner shouldn't "execute" tools, it should just "pass through" the request to the caller.
// But the Runner loop calls this during StatusWaitingForTool.
func (h *JSONHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	// 1. Try Local Execution (Registry)
	if h.Registry != nil {
		// Check if the tool exists locally
		result, err := h.Registry.Execute(ctx, call.Name, call.Args)
		if err == nil {
			// Success: Return immediately, skip network IO
			return domain.ToolResult{
				ID:     call.ID,
				Result: result,
			}, nil
		}
		// If error is "tool not found", fall through to network/JSON fallback.
		// If it's a legitimate execution error, we should probably return it.
		// But how can we distinguish "not found" cleanly from Registry.Execute?
		// Registry.Execute returns error if not found.
		// Let's rely on the error string for now or peek.
		// Actually, standardizing: IF Registry is present, we SHOULD execute it or fail?
		// Current plan: Server-side tools (Registry) + Client-side instruments (JSON/Fallback).
		// So if Registry.Execute fails with "tool not found", PROCEED to fallback.
		// If it fails with execution error, RETURN error.
		if err.Error() != fmt.Sprintf("tool not found: %s", call.Name) {
			return domain.ToolResult{
				ID:      call.ID,
				IsError: true,
				Error:   err.Error(),
			}, nil
		}
		// Tool not found in registry -> Fallback to client side
	}

	// 2. Fallback: Network/JSON (Client Execution)
	// Emit Request
	req := domain.ActionRequest{
		Type:    domain.ActionCallTool,
		Payload: call,
	}
	if err := h.Encoder.Encode([]domain.ActionRequest{req}); err != nil {
		return domain.ToolResult{}, err
	}

	// 3. Read Response from Stdin
	var result domain.ToolResult
	if err := h.Decoder.Decode(&result); err != nil {
		return domain.ToolResult{}, fmt.Errorf("failed to decode tool result: %w", err)
	}
	return result, nil
}
