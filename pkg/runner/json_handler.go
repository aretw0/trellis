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
	Writer   io.Writer
	Encoder  *json.Encoder
	Registry *registry.Registry

	// Async reading fields
	linesCh chan string
	errCh   chan error
}

// NewJSONHandler creates a handler for JSON IO.
// It starts a background goroutine to read from r.
func NewJSONHandler(r io.Reader, w io.Writer) *JSONHandler {
	if r == nil {
		r = os.Stdin
	}
	if w == nil {
		w = os.Stdout
	}

	h := &JSONHandler{
		Writer:  w,
		Encoder: json.NewEncoder(w),
		linesCh: make(chan string),
		errCh:   make(chan error),
	}

	// Start background reader
	go h.startReadLoop(r)

	return h
}

func (h *JSONHandler) startReadLoop(r io.Reader) {
	scanner := bufio.NewReader(r)
	for {
		line, err := scanner.ReadString('\n')
		if line != "" {
			// Send non-empty lines (even if err != nil, e.g. EOF with partial line)
			h.linesCh <- line
		}
		if err != nil {
			// If EOF or error, send error and exit loop
			h.errCh <- err
			return
		}
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
	// Wait for a line from linesCh OR context cancellation
	select {
	case line := <-h.linesCh:
		text := strings.TrimSpace(line)

		// Try to unquote if it's a JSON string
		var val string
		if err := json.Unmarshal([]byte(text), &val); err == nil {
			return val, nil
		}
		// Fallback: return raw text
		return text, nil

	case err := <-h.errCh:
		if err == io.EOF {
			return "", io.EOF
		}
		return "", fmt.Errorf("read error: %w", err)

	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// HandleTool for JSONHandler emits the tool call as JSON.
func (h *JSONHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	// 1. Try Local Execution (Registry)
	if h.Registry != nil {
		result, err := h.Registry.Execute(ctx, call.Name, call.Args)
		if err == nil {
			return domain.ToolResult{
				ID:     call.ID,
				Result: result,
			}, nil
		}
		if err.Error() != fmt.Sprintf("tool not found: %s", call.Name) {
			return domain.ToolResult{
				ID:      call.ID,
				IsError: true,
				Error:   err.Error(),
			}, nil
		}
	}

	// 2. Fallback: Network/JSON (Client Execution)
	req := domain.ActionRequest{
		Type:    domain.ActionCallTool,
		Payload: call,
	}
	if err := h.Encoder.Encode([]domain.ActionRequest{req}); err != nil {
		return domain.ToolResult{}, err
	}

	// 3. Read Response from Async Channel
	// Note: We expect the ToolResult to be sent as a single JSON line.
	select {
	case line := <-h.linesCh:
		var result domain.ToolResult
		// We expect the line to be the JSON object of ToolResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			return domain.ToolResult{}, fmt.Errorf("failed to decode tool result from JSONL: %w. Input was: %s", err, line)
		}
		return result, nil

	case err := <-h.errCh:
		return domain.ToolResult{}, fmt.Errorf("read error during tool wait: %w", err)

	case <-ctx.Done():
		return domain.ToolResult{}, ctx.Err()
	}
}

func (h *JSONHandler) SystemOutput(ctx context.Context, msg string) error {
	actions := []domain.ActionRequest{
		{
			Type:    domain.ActionSystemMessage,
			Payload: msg,
		},
	}
	return h.Encoder.Encode(actions)
}
