package trellis

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
)

// IOHandler defines the strategy for interacting with the user.
// This allows switching between Text (CLI/TUI) and JSON (Structured) modes.
type IOHandler interface {
	// Output presents the actions to the user.
	// Returns true if the output requires user input (e.g. asking a question),
	// or if the handler expects to read input after this.
	Output(ctx context.Context, actions []domain.ActionRequest) (bool, error)

	// Input reads a response from the user.
	Input(ctx context.Context) (string, error)
}

// TextHandler implements the standard text-based interface.
type TextHandler struct {
	Reader   *bufio.Reader
	Writer   io.Writer
	Renderer ContentRenderer
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
	hasContent := false
	for _, act := range actions {
		if act.Type == domain.ActionRenderContent {
			if msg, ok := act.Payload.(string); ok {
				hasContent = true
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
	}
	return hasContent, nil
}

func (h *TextHandler) Input(ctx context.Context) (string, error) {
	// Prompt
	fmt.Fprint(h.Writer, "> ")

	text, err := h.Reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}
