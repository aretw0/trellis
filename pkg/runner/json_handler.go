package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
)

// JSONHandler implements the IOHandler interface for structured JSON-Lines communication.
type JSONHandler struct {
	Reader  *bufio.Reader
	Writer  io.Writer
	Encoder *json.Encoder
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
