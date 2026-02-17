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

// ContentRenderer transforms a raw string (e.g. markdown) into a rendered string (e.g. ANSI).
type ContentRenderer func(content string) (string, error)

// TextHandler implements the standard text-based interface.
type TextHandler struct {
	Writer   io.Writer
	Renderer ContentRenderer
	Registry *registry.Registry

	inputChan chan inputResult
	buffer    int
}

type inputResult struct {
	text string
	err  error
}

// TextHandlerOption defines configuration for TextHandler.
type TextHandlerOption func(*TextHandler)

// WithTextHandlerRegistry configures the tool registry.
func WithTextHandlerRegistry(reg *registry.Registry) TextHandlerOption {
	return func(h *TextHandler) {
		h.Registry = reg
	}
}

// WithTextHandlerRenderer configures the content renderer.
func WithTextHandlerRenderer(renderer ContentRenderer) TextHandlerOption {
	return func(h *TextHandler) {
		h.Renderer = renderer
	}
}

// WithTextInputBufferSize sets the size of the input buffer.
func WithTextInputBufferSize(size int) TextHandlerOption {
	return func(h *TextHandler) {
		h.buffer = size
	}
}

// WithStdin enables reading from os.Stdin for simple library usage.
func WithStdin() TextHandlerOption {
	return func(h *TextHandler) {
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				h.FeedInput(scanner.Text(), nil)
			}
			if err := scanner.Err(); err != nil {
				h.FeedInput("", err)
			} else {
				h.FeedInput("", io.EOF)
			}
		}()
	}
}

// NewTextHandler creates a handler for standard text IO.
func NewTextHandler(w io.Writer, opts ...TextHandlerOption) *TextHandler {
	if w == nil {
		w = os.Stdout
	}
	h := &TextHandler{
		Writer: w,
		buffer: DefaultInputBufferSize,
	}

	h.inputChan = make(chan inputResult, h.buffer)

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// FeedInput feeds a line of input into the handler.
// This is called by the bridge when lifecycle detects input.
func (h *TextHandler) FeedInput(text string, err error) {
	select {
	case h.inputChan <- inputResult{text: text, err: err}:
	default:
		// Drop if blocked (shouldn't happen with proper buffer/flow)
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
	for {
		fmt.Fprint(h.Writer, "> ")

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case res, ok := <-h.inputChan:
			if !ok {
				return "", io.EOF
			}
			if res.err != nil {
				return "", res.err
			}

			text := strings.TrimSpace(res.text)

			// Sanitize Input (Security & Consistency)
			clean, err := SanitizeInput(text)
			if err != nil {
				// For text/interactive handler, we provide feedback and retry.
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

func (h *TextHandler) Signal(ctx context.Context, name string, args map[string]any) error {
	// Visual feedback for signals (e.g. "thinking")
	if name == "interrupt" {
		fmt.Fprint(h.Writer, "[CTRL+C]\n")
	}
	return nil
}
