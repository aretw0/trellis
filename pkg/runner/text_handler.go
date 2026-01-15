package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/registry"
)

// TextHandler implements the standard text-based interface.
type TextHandler struct {
	Reader   *bufio.Reader
	Writer   io.Writer
	Renderer ContentRenderer
	Registry *registry.Registry
}

// NewTextHandler creates a handler for standard text IO.
func NewTextHandler(r io.Reader, w io.Writer) *TextHandler {
	if r == nil {
		r = os.Stdin
	}
	if w == nil {
		w = os.Stdout
	}
	return &TextHandler{
		Reader: bufio.NewReader(r),
		Writer: w,
	}
}

func (h *TextHandler) Output(ctx context.Context, actions []domain.ActionRequest) (bool, error) {
	needsInput := false
	for _, act := range actions {
		if act.Type == domain.ActionRenderContent {
			if msg, ok := act.Payload.(string); ok {
				output := msg
				if h.Renderer != nil {
					rendered, err := h.Renderer(msg)
					if err == nil {
						output = rendered
					}
				}
				fmt.Fprintln(h.Writer, strings.TrimSpace(output))
			}
		}
		if act.Type == domain.ActionRequestInput {
			needsInput = true
		}
	}
	return needsInput, nil
}

func (h *TextHandler) Input(ctx context.Context) (string, error) {
	// Prompt
	fmt.Fprint(h.Writer, "> ")

	type result struct {
		text string
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		text, err := h.Reader.ReadString('\n')
		ch <- result{text: text, err: err}
	}()

	select {
	case <-ctx.Done():
		// Context cancelled (e.g. signal received).
		// Note: The goroutine will leak until some input is provided/EOF.
		// Standard Go stdio is hard to interrupt cleanly without closing stdin.
		return "", ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return "", res.err
		}
		return strings.TrimSpace(res.text), nil
	}
}

// HandleTool for TextHandler mocks the execution by printing to stdout.
func (h *TextHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	fmt.Fprintf(h.Writer, "[Tool Call] ID=%s Name=%s Args=%v\n", call.ID, call.Name, call.Args)

	if h.Registry != nil {
		result, err := h.Registry.Execute(ctx, call.Name, call.Args)
		if err != nil {
			// If tool execution fails, we return it as an error result, but not necessarily a Go error.
			// However, depending on semantics, we might want to fail the transition.
			// For now, let's return the error in the result.
			return domain.ToolResult{
				ID:      call.ID,
				IsError: true,
				Error:   err.Error(),
			}, nil
		}
		return domain.ToolResult{
			ID:     call.ID,
			Result: result,
		}, nil
	}

	// For Phase 1, we just return a success mock.
	// In the future, this could ask the user "Allow execution?" or actually run a local script.
	return domain.ToolResult{
		ID:      call.ID,
		Result:  "mock_success", // Default mock result
		IsError: false,
	}, nil
}

func (h *TextHandler) SystemOutput(ctx context.Context, msg string) error {
	// For text output, we can perhaps style it differently (e.g. bold, or different stream).
	// Ideally we print to stderr for logs, or just stdout with a prefix.
	// Let's use "[System]" prefix for now.
	fmt.Fprintf(h.Writer, "\n[System] %s\n", msg)
	return nil
}
