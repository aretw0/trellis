package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/registry"
)

// TextHandler implements the standard text-based interface.
type TextHandler struct {
	source      io.Reader
	interactive bool // true if reading from CONIN$ (Windows) where EOF should be ignored
	Reader      *bufio.Reader
	Writer      io.Writer
	Renderer    ContentRenderer
	Registry    *registry.Registry

	inputChan chan inputResult
	startOnce sync.Once
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

// NewTextHandler creates a handler for standard text IO.
func NewTextHandler(r io.Reader, w io.Writer, opts ...TextHandlerOption) *TextHandler {
	if r == nil {
		r = os.Stdin
	}
	if w == nil {
		w = os.Stdout
	}
	h := &TextHandler{
		source: r,
		Writer: w,
	}

	// Windows Specific: Check if we are running in a terminal.
	// If so, we MUST use CONIN$ to read input.
	// Standard os.Stdin on Windows closes immediately on Ctrl+C (EOF),
	// which creates a race condition with the SignalManager.
	// CONIN$ stays open, allowing the SignalManager to catch the Interrupt first.
	if runtime.GOOS == "windows" {
		if f, ok := r.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
			conin, err := os.Open("CONIN$")
			if err == nil {
				h.source = conin
				h.interactive = true
			}
		}
	}

	h.Reader = bufio.NewReader(h.source)

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (h *TextHandler) initPump() {
	h.startOnce.Do(func() {
		h.inputChan = make(chan inputResult)
		go h.pump()
	})
}

func (h *TextHandler) pump() {
	for {
		text, err := h.Reader.ReadString('\n')

		// If we got text (even with EOF), send it
		if text != "" {
			h.inputChan <- inputResult{text: text, err: nil}
		}

		if err != nil {
			if err == io.EOF {
				if h.interactive {
					// In interactive mode (e.g. Windows CONIN$), EOF might mean
					// a signal interrupted the read, but the stream is still valid regarding the OS.
					// We pass the EOF to the consumer so they know the current read failed (likely due to signal),
					// but we DO NOT close the channel so future reads can happen (e.g. after signal handling).
					h.inputChan <- inputResult{text: "", err: io.EOF}
					// Prevent busy loop if EOFs are generated rapidly (e.g. holding Ctrl+C)
					time.Sleep(50 * time.Millisecond)
					continue
				}
				close(h.inputChan)
				return
			}
			// Send non-EOF errors
			h.inputChan <- inputResult{text: "", err: err}
			// Backoff for non-fatal errors to prevent CPU spikes on persistent failure
			time.Sleep(50 * time.Millisecond)
		}
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
