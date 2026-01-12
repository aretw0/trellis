package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aretw0/trellis/pkg/domain"
)

func TestJSONHandler_AsyncInput(t *testing.T) {
	// 1. Test Cancellation
	t.Run("ContextCancellation", func(t *testing.T) {
		r, w := io.Pipe()
		defer w.Close()
		handler := NewJSONHandler(r, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := handler.Input(ctx)
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if ctx.Err() != context.DeadlineExceeded && ctx.Err() != context.Canceled {
			t.Errorf("Expected context error, got: %v", err)
		}
	})

	// 2. Test Success
	t.Run("Success", func(t *testing.T) {
		input := "\"Valid JSON String\"\n"
		r := bytes.NewBufferString(input)
		handler := NewJSONHandler(r, nil)

		ctx := context.Background() // No timeout needed for byte buffer as it's immediate
		val, err := handler.Input(ctx)
		if err != nil {
			t.Fatalf("Input failed: %v", err)
		}
		if val != "Valid JSON String" {
			t.Errorf("Expected 'Valid JSON String', got '%s'", val)
		}
	})

	// 3. Test EOF
	t.Run("EOF", func(t *testing.T) {
		r := bytes.NewBufferString("")
		handler := NewJSONHandler(r, nil)

		// We need to wait a small bit for the goroutine to process the EOF
		// But cleaner is to just call Input
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := handler.Input(ctx)
		if err != io.EOF {
			// Note: bytes.Buffer returns EOF immediately.
			// The scanner loop should catch it and send to errCh.
			t.Errorf("Expected EOF, got: %v", err)
		}
	})
}

func TestJSONHandler_AsyncHandleTool(t *testing.T) {
	// 1. Test Cancellation
	t.Run("Cancellation", func(t *testing.T) {
		r, w := io.Pipe()
		defer w.Close()
		// Capture output to avoid noise
		outBuf := &bytes.Buffer{}
		handler := NewJSONHandler(r, outBuf)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Call tool "test"
		call := domain.ToolCall{ID: "1", Name: "test"}
		_, err := handler.HandleTool(ctx, call)

		if err == nil {
			t.Error("Expected error from HandleTool, got nil")
		}
	})

	// 2. Test Success (JSONL Response)
	t.Run("Success", func(t *testing.T) {
		// Prepare a mocked response
		result := domain.ToolResult{
			ID:     "1",
			Result: "Success Result",
		}
		resultJSON, _ := json.Marshal(result)
		input := string(resultJSON) + "\n"

		r := bytes.NewBufferString(input)
		outBuf := &bytes.Buffer{}
		handler := NewJSONHandler(r, outBuf)

		ctx := context.Background()
		call := domain.ToolCall{ID: "1", Name: "test"}

		res, err := handler.HandleTool(ctx, call)
		if err != nil {
			t.Fatalf("HandleTool failed: %v", err)
		}
		if res.Result != "Success Result" {
			t.Errorf("Expected 'Success Result', got '%s'", res.Result)
		}

		// Verify request was written
		encodedReq := strings.TrimSpace(outBuf.String())
		if !strings.Contains(encodedReq, `"CALL_TOOL"`) {
			t.Errorf("Output did not contain 'CALL_TOOL', got: %s", encodedReq)
		}
	})
}
