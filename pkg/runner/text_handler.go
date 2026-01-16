package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/registry"
)

// TextHandler implements the standard text-based interface.
type TextHandler struct {
	Reader   *bufio.Reader
	Writer   io.Writer
	Renderer ContentRenderer
	Registry *registry.Registry

	inputChan chan inputResult
	startOnce sync.Once
}

type inputResult struct {
	text string
	err  error
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

func (h *TextHandler) initPump() {
	h.startOnce.Do(func() {
		h.inputChan = make(chan inputResult)
		go func() {
			for {
				text, err := h.Reader.ReadString('\n')
				// Send result. This blocks until someone (Input) reads it.
				// This implies that if no one asks for input, we buffer exactly one line (OS buffer aside).
				h.inputChan <- inputResult{text: text, err: err}
				if err != nil {
					close(h.inputChan)
					return
				}
			}
		}()
	})
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
	// Ensure the pump is running
	h.initPump()

	for {
		// Only show prompt if context is not yet done
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			fmt.Fprint(h.Writer, "> ")
		}

		select {
		case <-ctx.Done():
			// Important: don't print anything here, just exit silently
			return "", ctx.Err()
		case res, ok := <-h.inputChan:
			if !ok {
				return "", io.EOF
			}
			if res.err != nil {
				return "", res.err
			}
			text := strings.TrimSpace(res.text)

			// Sanitize Input (Limit + Control Chars)
			clean, err := SanitizeInput(text)
			if err != nil {
				// User Feedback: Prompt retry
				fmt.Fprintf(h.Writer, "Error: %v. Please try again.\n", err)
				continue
			}
			return clean, nil
		}
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
